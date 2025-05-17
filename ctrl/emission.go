package main

import (
	"fmt"
	"gospital/utils"
	"strconv"
)


func (c *utils.Controller) EnvoyerSurAnneau(msgType utils.MessageType, content interface{}) {
	/* Cette fonction construit un message et l'envoie sur l'anneau
	 Format du message dépend du type de message*/
	
	var msg string
	
	switch msgType {
	case utils.EtatMsg:
		if etatMsg, ok := content.(utils.EtatMessage); ok {
			// Formater le message d'état (à adapter selon votre format)
			etatStr := fmt.Sprintf("%v", etatMsg.EtatLocal) // Simple string representation
			msg = utils.Msg_format("type", string(utils.EtatMsg)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("etat", etatStr) +
				utils.Msg_format("bilan", strconv.Itoa(etatMsg.Bilan)) +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge))
		}
	case utils.PrePost:
		if prepostMsg, ok := content.(string); ok {
			msg = utils.Msg_format("type", string(utils.PrePost)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", prepostMsg) +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge))
		}
	case utils.AppMsg:
		if appMsg, ok := content.(utils.CtrlMessage); ok {
			msg = utils.Msg_format("type", string(utils.AppMsg)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", appMsg.Contenu) +
				utils.Msg_format("couleur",string(appMsg.Couleur)) +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge))
		}
	}
	
	if msg != "" {
		// Envoyer le message (pour l'anneau, on l'imprime simplement)
		fmt.Println(msg)
	}
}

func (c *utils.Controller) ForwardToApp(message string) {
    // Décrémenter le bilan (message sortant)
    c.Snapshot.Bilan--

    // Construire le message à transmettre avec la couleur
    ctrlMsg := utils.CtrlMessage{
        Type:    utils.AppMsg,
        Contenu: message,
        Couleur: c.Snapshot.Couleur,
    }

    // Envoyer directement sur l'anneau
    c.EnvoyerSurAnneau(utils.AppMsg, ctrlMsg)
}