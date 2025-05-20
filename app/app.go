package main

import (
	"flag"
	"fmt"
	"gospital/utils"
	ws "gospital/websocket"
	"log"
	"os"
	"strings"
	"sync"
)

var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)
var p_nom *string = flag.String("n", "nom", "nom")
var globalMutex = &sync.Mutex{}

func (d *DoctorInfo) SendDoctorInfo() utils.DoctorPayload {
	return utils.DoctorPayload{
		Sender:       *p_nom,
		DoctorsCount: d.DoctorsCount,
		ActivityLog:  d.ActivityLog,
		Snapshot:     d.Snapshot,
	}
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
			//ActivityLog: []string{
			//	"DemSC", "DebSC", "FinSC", "DemSC",
			//},
		},
		actions:   make(chan map[string]interface{}, 10),
		waitingSC: false,
	}
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
		utils.Display_w("action", fmt.Sprintf("%v", action["type"]), a.name)
		if action["type"] == "send" && a.doctorInfo.DoctorsCount[*p_nom] > 0 {
			destinator := strings.TrimSpace(action["to"].(string))
			go a.waitingFoSending(destinator)
		} else if action["type"] == "snapshot" {
			go a.snapshot()
		} else if action["type"] == "speed" {
			delayStr, ok := action["delay"].(string)
			if ok {
				fmt.Println(utils.Msg_format("type", "speed") + utils.Msg_format("delay", delayStr))
			}
		}
	}
}


func main() {
	flag.Parse()
	app := NewApp(*p_nom)
	app.run()
}
