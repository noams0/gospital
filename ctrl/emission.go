package main

import (
    "fmt"
    "strconv"
    "gospital/utils"
)


func (c *Controller) ForwardToApp(message string) {
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
	Le format du message dépend du type de message*/
	
	var msg string
	
	switch msgType {
	case EtatMsg:
		if etatMsg, ok := content.(EtatMessage); ok {
			// Formater le message d'état (à adapter selon votre format)
			etatStr := fmt.Sprintf("%v", etatMsg.EtatLocal) // Simple string representation
			msg = utils.Msg_format("type", string(EtatMsg)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("etat", etatStr) +
				utils.Msg_format("bilan", strconv.Itoa(etatMsg.Bilan)) +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge))
		}
	case PrePost:
		if prepostMsg, ok := content.(string); ok {
			msg = utils.Msg_format("type", string(PrePost)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", prepostMsg) +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge))
		}
	case AppMsg:
		if appMsg, ok := content.(CtrlMessage); ok {
			msg = utils.Msg_format("type", string(AppMsg)) +
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