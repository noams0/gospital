package websocket

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Autorise toutes les origines
	},
}

type DoctorInfoSender interface {
	SendDoctorInfo() map[string]int
}

// handleWS gère la connexion WebSocket pour chaque client
func handleWS(w http.ResponseWriter, r *http.Request, addr string, infoSender DoctorInfoSender, actions chan map[string]interface{}) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	//fmt.Println("Client connecté")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var action map[string]interface{}

			if err := json.Unmarshal(msg, &action); err != nil {
				println("Invalid message format:", string(msg))
				return
			}
			actions <- action
		}
	}()
	for {
		select {
		case <-ticker.C:
			//message := fmt.Sprintf("Heure serveur : %s", t.Format("15:04:05")) + "from adress" + addr
			data := infoSender.SendDoctorInfo() // map[string]int
			jsonData, _ := json.Marshal(data)
			conn.WriteMessage(websocket.TextMessage, jsonData)

		}
	}
}

// StartServer démarre le serveur WebSocket à l'adresse spécifiée
func StartServer(address string, infoSender DoctorInfoSender, actions chan map[string]interface{}) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWS(w, r, address, infoSender, actions)
	})
	//fmt.Printf("WebSocket en écoute sur %s/ws\n", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		//fmt.Println("Erreur du serveur WebSocket: ", err)
	}
}
