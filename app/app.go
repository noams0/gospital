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
	/*Retourne un objet contenant l’état courant du médecin
	(compteurs, logs, snapshot) pour l’envoi.*/
	return utils.DoctorPayload{
		Sender:       *p_nom,
		DoctorsCount: d.DoctorsCount,
		ActivityLog:  d.ActivityLog,
		Snapshot:     d.Snapshot,
	}
}

func NewApp(name string, totalSites int) *App {
	/*Initialise et retourne une nouvelle instance de l’application */
	doctors := make(map[string]int)
	for i := 1; i <= totalSites; i++ {
		appName := fmt.Sprintf("app_%d", i)
		doctors[appName] = 5 // nombre initial de médecins pour ce site
	}

	return &App{
		name: name,
		doctorInfo: DoctorInfo{
			DoctorsCount: doctors,
		},
		actions:   make(chan map[string]interface{}, 10),
		waitingSC: false,
	}
}

func (a *App) run() {
	/*Lance WebSocket, démarre la réception des actions,
	puis traite les actions reçues*/
	var wsURL string
	wsURL = defineWSurl(a.name)

	go ws.StartServer(wsURL, &a.doctorInfo, a.actions)
	go a.receive()

	for action := range a.actions {
		utils.Display_w("action", fmt.Sprintf("%v", action["type"]), a.name)
		if action["type"] == "send" && a.doctorInfo.DoctorsCount[*p_nom] > 0 {
			destinator := strings.TrimSpace(action["to"].(string))
			go a.waitingFoSending(destinator)
		} else if action["type"] == "snapshot" {
			go a.snapshot()
		} else if action["type"] == "askToLeave" {
			go a.leave()
		} else if action["type"] == "speed" {
			delayStr, ok := action["delay"].(string)
			if ok {
				fmt.Println(utils.Msg_format("type", "speed") + utils.Msg_format("delay", delayStr))
			}
		}
	}
}

func main() {
	/* Analyse les arguments de la ligne de commande,
	crée l’application et lance son exécution.*/
	totalSites := flag.Int("total", 3, "Nombre total de sites")
	flag.Parse()
	app := NewApp(*p_nom, *totalSites)
	app.run()
}
