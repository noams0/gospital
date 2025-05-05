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

type Hospital struct {
	name       string
	numDoctors int
	mu         sync.Mutex
}

var hospitalRegistry = make(map[string]*Hospital) // Répertoire des hôpitaux
var mu = &sync.Mutex{}

func registerHospital(name string) *Hospital {
	mu.Lock()
	defer mu.Unlock()
	hospital := &Hospital{name: name}
	hospitalRegistry[name] = hospital
	return hospital
}

func updateDoctorsCount(hospital *Hospital, count int) {
	hospital.mu.Lock()
	defer hospital.mu.Unlock()
	hospital.numDoctors = count
	//fmt.Printf("Mise à jour du nombre de médecins pour %s: %d\n", hospital.name, hospital.numDoctors)
}

func getDoctorsCount(hospital *Hospital) int {
	hospital.mu.Lock()
	defer hospital.mu.Unlock()
	return hospital.numDoctors
}

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
		if len(keyval) < 3 { // au moins 1 pour separateur, 1 pour key, 1 pour val
			display_w("findval", "message trop court : "+msg)
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

func sendperiodic() {
	var sndmsg string
	var i int

	i = 0

	for i < 5 {
		mutex.Lock()
		i = i + 1
		sndmsg = "message_" + strconv.Itoa(i) + "from" + strconv.Itoa(pid) + "\n"
		display_d("J'ENVOIE", sndmsg)
		fmt.Print(sndmsg)
		mutex.Unlock()
		time.Sleep(time.Duration(5) * time.Second)
	}
}

func receive() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		rcvmsg := scanner.Text()
		mutex.Lock()
		display_w("receive", "reception <"+rcvmsg+">")
		mutex.Unlock()
	}
	if err := scanner.Err(); err != nil {
		display_e("receive", "erreur de lecture: "+err.Error())
	}
}

var mutex = &sync.Mutex{}

func main() {

	flag.Parse()

	hospitalA := registerHospital("HospitalA")
	hospitalB := registerHospital("HospitalB")

	// Mise à jour du nombre de médecins
	updateDoctorsCount(hospitalA, 5)
	updateDoctorsCount(hospitalB, 3)

	var wsURL string
	switch *p_nom {
	case "app_1":
		wsURL = ":8080"
	case "app_2":
		wsURL = ":8081"
	case "app_3":
		wsURL = ":8082"
	default:
		log.Fatalf("Nom inconnu pour l'auto-déduction du WebSocket : %s", *p_nom)
	}
	go ws.StartServer(wsURL)

	go sendperiodic()
	go receive()
	for {
		time.Sleep(time.Duration(60) * time.Second)
	} // Pour attendre la fin des goroutines...
}
