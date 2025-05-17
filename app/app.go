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

func (d *utils.DoctorInfo) SendDoctorInfo() utils.DoctorPayload {
	return utils.DoctorPayload{
		Sender:       *p_nom,
		DoctorsCount: d.DoctorsCount,
		ActivityLog:  d.ActivityLog,
	}
}



func NewApp(name string) *utils.App {
	return &utils.App{
		Name: name,
		DoctorInfo: utils.DoctorInfo{
			DoctorsCount: map[string]int{
				"app_1": 5,
				"app_2": 3,
				"app_3": 7,
			},
			//ActivityLog: []string{
			//	"DemSC", "DebSC", "FinSC", "DemSC",
			//},
		},
		Actions:   make(chan map[string]interface{}, 10),
		WaitingSC: false,
	}
}


func (a *utils.App) run() {
	var wsURL string
	switch a.Name {
	case "app_1":
		wsURL = ":8080"
	case "app_2":
		wsURL = ":8081"
	case "app_3":
		wsURL = ":8082"
	default:
		log.Fatalf("Nom inconnu pour WebSocket : %s", a.Name)
	}

	go ws.StartServer(wsURL, &a.DoctorInfo, a.Actions)
	go a.receive()

	for action := range a.Actions {
		utils.Display_w("action", fmt.Sprintf("%v", action["to"]), a.Name)
		if action["type"] == "send" && a.DoctorInfo.DoctorsCount[*p_nom] > 0 {
			destinator := strings.TrimSpace(action["to"].(string))
			go a.waitingFoSending(destinator)
		} else if action["type"] == "snapshot" {
		go a.snapshot()
	}
	}
}

func main() {
	flag.Parse()
	app := NewApp(*p_nom)
	app.run()
}
