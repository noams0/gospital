package main

import (
	"bufio"
	"flag"
	"fmt"
	"gospital/utils"
	ws "gospital/websocket"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)
var p_nom *string = flag.String("n", "nom", "nom")
var globalMutex = &sync.Mutex{}

var mu = &sync.Mutex{}

func display_d(where string, what string) {
	stderr.Printf("%s + [%.6s %d] %-8.8s : %s\n%s", utils.ColorBlue, *p_nom, pid, where, what, utils.ColorReset)
}
func display_w(where string, what string) {
	stderr.Printf("%s * [%.6s %d] %-8.8s : %s\n%s", utils.ColorYellow, *p_nom, pid, where, what, utils.ColorReset)
}
func display_e(where string, what string) {
	stderr.Printf("%s ! [%.6s %d] %-8.8s : %s\n%s", utils.ColorRed, *p_nom, pid, where, what, utils.ColorReset)
}

func findval(msg string, key string) string {
	if len(msg) < 4 {
		display_w("findval", "message trop court : "+msg)
		return ""
	}
	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)

	for _, keyval := range tab_allkeyvals {
		if len(keyval) < 3 {
			display_w("findval", "clé-valeur trop courte : "+keyval)
			continue
		}
		equ := keyval[0:1]
		tabkeyval := strings.SplitN(keyval[1:], equ, 2)
		if len(tabkeyval) != 2 {
			continue
		}
		if tabkeyval[0] == key {
			return tabkeyval[1]
		}
	}
	return ""
}

func (d *DoctorInfo) SendDoctorInfo() utils.DoctorPayload {
	return utils.DoctorPayload{
		Sender:       *p_nom,
		DoctorsCount: d.DoctorsCount,
	}
}

type DoctorInfo struct {
	DoctorsCount map[string]int
}

// Struct App
type App struct {
	name       string
	doctorInfo DoctorInfo
	actions    chan map[string]interface{}
	waitingSC  bool
	inSC       bool
}

func NewApp(name string) *App {
	return &App{
		name: name,
		doctorInfo: DoctorInfo{
			DoctorsCount: map[string]int{
				"app_1": 5,
				"app_2": 3,
				"app_3": 7,
			},
		},
		actions:   make(chan map[string]interface{}, 10),
		waitingSC: false,
	}
}

func (a *App) receive() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		globalMutex.Lock()
		display_w("receive", "reception <"+msg+">")
		if msg == "receive" {
			go a.waitingFoReceivng()
		}
		if msg == "debutSC" && a.waitingSC {
			a.inSC = true
			a.waitingSC = false
		} else if findval(msg, "type") == "new_data" {
			display_w("NEW_DATA", "IL FAUT MAJ")
			data := findval(msg, "new_data")
			pairs := strings.Split(data, "|")
			for _, pair := range pairs {
				parts := strings.Split(pair, "=")
				if len(parts) == 2 {
					appName := parts[0]
					val, err := strconv.Atoi(parts[1])
					if err == nil {
						a.doctorInfo.DoctorsCount[appName] = val
						display_w("NEW_DATA", fmt.Sprintf("Mise à jour : %s -> %d", appName, val))
					} else {
						display_e("NEW_DATA", "Erreur de conversion pour "+pair)
					}
				}
			}

		}

		globalMutex.Unlock()
	}
	if err := scanner.Err(); err != nil {
		display_e("receive", "erreur de lecture: "+err.Error())
	}
}

func (a *App) waitingFoReceivng() {
	fmt.Print("demandeSC\n")
	a.waitingSC = true
	for !a.inSC {
		time.Sleep(100 * time.Millisecond)
	}
	a.doctorInfo.DoctorsCount[*p_nom]++
	msg := "finSC"
	for site, count := range a.doctorInfo.DoctorsCount {
		msg += fmt.Sprintf("|%s=%d", site, count)
	}

	fmt.Print(msg + "\n")
}

func (a *App) waitingFoSending(destinator string) {
	fmt.Print("demandeSC\n")
	a.waitingSC = true
	for !a.inSC {
		time.Sleep(100 * time.Millisecond)
	}
	a.doctorInfo.DoctorsCount[*p_nom]--
	if a.doctorInfo.DoctorsCount[*p_nom] < 0 {
		a.doctorInfo.DoctorsCount[*p_nom] = 0
	}

	msg := "finSC"
	for site, count := range a.doctorInfo.DoctorsCount {
		msg += fmt.Sprintf("|%s=%d", site, count)
	}

	fmt.Print(msg + "\n")
	msg = "send" + destinator
	display_w("action :", msg)
	fmt.Print(msg + "\n")

	a.inSC = false
}

func (a *App) run() {
	var wsURL string
	switch a.name {
	case "app_1":
		wsURL = ":8080"
	case "app_2":
		wsURL = ":8081"
	case "app_3":
		wsURL = ":8082"
	default:
		log.Fatalf("Nom inconnu pour WebSocket : %s", a.name)
	}

	go ws.StartServer(wsURL, &a.doctorInfo, a.actions)
	go a.receive()

	for action := range a.actions {
		display_w("action", fmt.Sprintf("%v", action["to"]))
		if action["type"] == "send" {
			destinator := strings.TrimSpace(action["to"].(string))
			go a.waitingFoSending(destinator)
		}
	}
}

func main() {
	flag.Parse()
	app := NewApp(*p_nom)
	app.run()
}
