package main

import (
    "fmt"
    "strconv"
    "encoding/json"
    "log"
    "gospital/utils"
)

func (c *Controller) ForwardToApp(message string) {
	/*cette fonction transmet un message à l'application*/

	// Décrémenter le bilan (message sortant)
	c.Snapshot.Bilan--

	// Construire le message à transmettre avec la couleur
	ctrlMsg := CtrlMessage{
		Type:    AppMsg,
		Contenu: message,
		Couleur: c.Snapshot.Couleur,
	}

	// Envoyer directement sur l'anneau
	c.EnvoyerSurAnneau(AppMsg, ctrlMsg)
}



func (c *Controller) EnvoyerSurAnneau(msgType MessageType, content interface{}) {

	/* Cette fonction construit un message et l'envoie sur l'anneau
	Format du message dépend du type de message*/

	var msg string

	switch msgType {
	case EtatMsg:
		if etatMsg, ok := content.(EtatMessage); ok {

			etatBytes, err := json.Marshal(etatMsg.EtatLocal)
			if err != nil {
				log.Fatal("Erreur lors du marshalling de l'état local :", err)
			}
			etatStr := string(etatBytes)
			msg = utils.Msg_format("type", string(EtatMsg)) +
				utils.Msg_format("couleur", string(c.Snapshot.Couleur)) +
				utils.Msg_format("msg", "1") +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("etat", etatStr) +
				utils.Msg_format("bilan", strconv.Itoa(etatMsg.Bilan)) +
				c.Msg_Horloge()
		}
	case PrePost:

		if prepostMsg, ok := content.(string); ok {

			msg = utils.Msg_format("type", string(PrePost)) +
				utils.Msg_format("couleur", string(c.Snapshot.Couleur)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", prepostMsg) +
				c.Msg_Horloge()
		}
	case AppMsg:

		if appMsg, ok := content.(CtrlMessage); ok {
			msg = utils.Msg_format("type", string(AppMsg)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", appMsg.Contenu) +
				utils.Msg_format("couleur", string(appMsg.Couleur)) +
				c.Msg_Horloge()
		}
	case SnapshotMsg:

		if msgStr, ok := content.(string); ok {
			msg = msgStr
		}
	case "FORWARD":
		if msgStr, ok := content.(string); ok {
			msg = msgStr
		}
	}
	if msg != "" {
		utils.Display_f("SEND", fmt.Sprintf("%s - %s", msgType, msg), c.Nom)
		fmt.Println(msg)
	}
}

