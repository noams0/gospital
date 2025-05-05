package websocket

import (
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
	SendDoctorInfo() string
}

// handleWS gère la connexion WebSocket pour chaque client
func handleWS(w http.ResponseWriter, r *http.Request, addr string, infoSender DoctorInfoSender) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	//fmt.Println("Client connecté")

	// Exemple : envoyer un message toutes les 2 secondes
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			message := infoSender.SendDoctorInfo()
			//message := fmt.Sprintf("Heure serveur : %s", t.Format("15:04:05")) + "from adress" + addr
			if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				//fmt.Println("Client déconnecté")
				return
			}
		}
	}
}

// StartServer démarre le serveur WebSocket à l'adresse spécifiée
func StartServer(address string, infoSender DoctorInfoSender) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Appeler la fonction avec un paramètre
		handleWS(w, r, address, infoSender)
	})
	//fmt.Printf("WebSocket en écoute sur %s/ws\n", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		//fmt.Println("Erreur du serveur WebSocket: ", err)
	}
}
