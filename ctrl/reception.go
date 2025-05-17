package main

import (
	"fmt"
	"gospital/utils"
	"strconv"
	"strings"
)

func (c *utils.Controller) handleAppMessage(rcvmsg string) {
	type_msg := utils.Findval(rcvmsg, "type", c.Nom)
	utils.Display_d("main", "TYPE de la demande en provenance de l'app : "+type_msg, c.Nom)

	switch type_msg {
	case "demandeSC":
		c.Horloge++
		c.Tab[c.Nom] = utils.EtatReqSite{
			TypeRequete: utils.Requete,
			Horloge:     c.Horloge,
		}
		utils.Display_f("demandeSC", "Demande de SC locale, horloge : "+strconv.Itoa(c.Horloge), c.Nom)
		fmt.Println(
			utils.Msg_format("type", string(utils.Requete)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", "1") +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge)))
	case "finSC":
		c.Horloge++
		c.Tab[c.Nom] = utils.EtatReqSite{
			TypeRequete: utils.Liberation,
			Horloge:     c.Horloge,
		}
		c.IsInSection = false
		utils.Display_f("finSC", "Fin de SC locale, horloge : "+strconv.Itoa(c.Horloge), c.Nom)
		newData := ""
		newData = utils.Findval(rcvmsg, "new_data", c.Nom) // ex: |app_1=4|app_2=3|app_3=7
		fmt.Println(
			utils.Msg_format("type", "liberation") +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", "finSC") +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge)) +
				utils.Msg_format("new_data", newData),
		)
	case "send":
		c.Horloge++
		var destApp string = utils.Findval(rcvmsg, "destinator", c.Nom)
		var destCtrl string = utils.App_to_ctrl(destApp)
		utils.Display_f("destinator :", destCtrl, c.Nom)
		fmt.Println(utils.Msg_format("type", "send") + utils.Msg_format("destinator", destCtrl) + utils.Msg_format("sender", c.Nom) + utils.Msg_format("msg", "send") + utils.Msg_format("hlg", strconv.Itoa(c.Horloge)))
	case "snapshot":
    	if !c.SnapshotEnCours {
            c.SnapshotEnCours = true
            c.Horloge++
            utils.Display_f("snapshot", "Demande de snapshot reçue de l'application", c.Nom)
            c.DebutSnapshot()
        }	
	default:
		//fmt.Println(utils.Msg_format("sender", c.Nom) + utils.Msg_format("msg", rcvmsg) + utils.Msg_format("hlg", strconv.Itoa(c.Horloge)))
	case "debutSC":
	case "receive":
		utils.Display_f("NOT", "for me", c.Nom)
	}

}

func (c *utils.Controller) handleCtrlMessage(rcvmsg string) {
	rcvVC := utils.DecodeVC(utils.Findval(rcvmsg, "VC", c.Nom))
	rcvHLG, _ := strconv.Atoi(utils.Findval(rcvmsg, "hlg", c.Nom))
	sndmsg := utils.Findval(rcvmsg, "msg", c.Nom)

	if len(rcvVC) != 0 {
		utils.Display_d("main", fmt.Sprintf("horloge reçue : %#v", rcvVC), c.Nom)
		for k, v := range rcvVC {
			if _, ok := c.VectorClock[k]; !ok {
				c.VectorClock[k] = 0
			}
			if v > c.VectorClock[k] {
				c.VectorClock[k] = v
			}
		}
		c.VectorClock[*p_nom]++
		//utils.Display_e("main", "Nouvelle horloge :"+utils.Encodehorloge(c.VectorClock))
	} else {
		c.VectorClock[*p_nom]++
	}
	c.Horloge = max(rcvHLG, c.Horloge) + 1

	msg_type := utils.Findval(rcvmsg, "type", c.Nom)
	sender := utils.Findval(rcvmsg, "sender", c.Nom)
	//utils.Display_f("TYPE", msg_type)
	switch msg_type {
	case "new_data":
		utils.Display_f("NOT", "for me", c.Nom)
	case string(utils.Requete):
		if sender != *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrêt
			c.Tab[sender] = utils.EtatReqSite{
				TypeRequete: utils.Requete,
				Horloge:     rcvHLG,
			}
			utils.Display_f(string(utils.Requete), "Requête reçue de "+sender+" | horloge="+strconv.Itoa(c.Horloge), c.Nom)
			//envoyer( [accusé] hi ) à Sj
			utils.Display_f(string(utils.Requete), rcvmsg, c.Nom)
			utils.Display_f(string(utils.Requete), fmt.Sprintf("mon tab %#v", c.Tab), c.Nom)
			fmt.Println(rcvmsg)

			fmt.Println(utils.Msg_format("destinator", sender) + utils.Msg_format("msg", "ack") + utils.Msg_format("type", "ack") + utils.Msg_format("sender", c.Nom) + utils.Msg_format("hlg", strconv.Itoa(c.Horloge)))
			if c.Tab[c.Nom].TypeRequete == utils.Requete && !c.IsInSection {
				if isFirstRequest(c.Tab, c.Nom, c.Tab[c.Nom].Horloge) {
					c.IsInSection = true
					utils.Display_f("SC", "\n ======================", c.Nom)
					utils.Display_f("SC", "Entrée en SC autorisée", c.Nom)
					utils.Display_f("SC", "\n ======================", c.Nom)
					fmt.Print("debutSC\n")
					utils.Display_e("je veux passer mon tab", fmt.Sprintf("%#v", c.Tab), c.Nom)
					fmt.Println(TabToString(c.Tab))
					//fmt.Println("\n")
				}
			}
		}
	case string(utils.Liberation):
		if sender != *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrêt
			c.Tab[sender] = utils.EtatReqSite{
				TypeRequete: utils.Liberation,
				Horloge:     rcvHLG,
			}
			new_data := utils.Findval(rcvmsg, "new_data", c.Nom)
			if new_data != "" {
				fmt.Println(
					utils.Msg_format("type", "new_data") +
						utils.Msg_format("new_data", new_data) + // ex : new_data|app_1=5|app_2=2|app_3=6
						utils.Msg_format("msg", "finSC"))
			}
			utils.Display_f("liberation", "Libération reçue de "+sender+" | horloge="+strconv.Itoa(c.Horloge), c.Nom)
			utils.Display_f("liberation", fmt.Sprintf("mon tab %#v", c.Tab), c.Nom)
			fmt.Println(rcvmsg)
			//envoyer( [accusé] hi ) à Sj
			if c.Tab[c.Nom].TypeRequete == utils.Requete && !c.IsInSection {
				if isFirstRequest(c.Tab, c.Nom, c.Tab[c.Nom].Horloge) {
					c.IsInSection = true
					utils.Display_f("SC", "\n ======================", c.Nom)
					utils.Display_f("SC", "Entrée en SC autorisée", c.Nom)
					utils.Display_f("SC", "\n ======================", c.Nom)
					fmt.Print("debutSC\n")
					utils.Display_e("je veux passer mon tab", fmt.Sprintf("%#v", c.Tab), c.Nom)
					fmt.Println(TabToString(c.Tab) + "\n")
				}
			}
			utils.Display_f("liberation", "libération reçue de "+sender+" | horloge="+strconv.Itoa(c.Horloge), c.Nom)
		}
	case "ack":
		if utils.Findval(rcvmsg, "destinator", c.Nom) == *p_nom+"-"+strconv.Itoa(pid) {
			if c.Tab[sender].TypeRequete != utils.Requete {
				c.Tab[sender] = utils.EtatReqSite{
					TypeRequete: utils.Accuse,
					Horloge:     rcvHLG,
				}
			}
			utils.Display_f("Accusé", "Accusé reçue de "+sender+" | horloge="+strconv.Itoa(c.Horloge), c.Nom)
			utils.Display_f("Accusé", fmt.Sprintf("mon c.Tab %#v", c.Tab), c.Nom)

			//envoyer( [accusé] hi ) à Sj
			if c.Tab[c.Nom].TypeRequete == utils.Requete && !c.IsInSection {
				utils.Display_f("TENTATIVE", "Je vais tenter de voir si je suis le premier", c.Nom)

				if isFirstRequest(c.Tab, c.Nom, c.Tab[c.Nom].Horloge) {
					c.IsInSection = true
					utils.Display_f("SC", "\n ======================", c.Nom)
					utils.Display_f("SC", "Entrée en SC autorisée", c.Nom)
					utils.Display_f("SC", "\n ======================", c.Nom)
					fmt.Print("debutSC\n")
					utils.Display_e("je veux passer mon tab", fmt.Sprintf("%#v", c.Tab), c.Nom)
					fmt.Println(TabToString(c.Tab))
					//fmt.Println("\n")"
				}
			}
		} else {
			fmt.Println(rcvmsg)
		}

	case "send":
		if utils.Findval(rcvmsg, "destinator", c.Nom) == *p_nom {
			utils.Display_f("send", "send pour oim"+rcvmsg, c.Nom)

			fmt.Println("receive")
		} else {
			utils.Display_f("send", "send pas pour oim"+rcvmsg, c.Nom)
			fmt.Println(rcvmsg)
		}
    case "snapshot":
	    c.handleSnapshotMessage(rcvmsg)
    case string(utils.EtatMsg):
		var etatRecu map[string]interface{}
		var bilanRecu int
		c.ReceptionMsgEtat(etatRecu, bilanRecu)
	case string(utils.PrePost):
		c.ReceptionMsgPrepost(rcvmsg)
	default:
		if sender == *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrête
			utils.Display_e("main", "Arret du message :"+rcvmsg, c.Nom)
		} else {
			utils.Display_e("main", "sender :"+utils.Findval(rcvmsg, "sender", c.Nom), c.Nom)
			utils.Display_e("main", "sender :"+*p_nom+"-"+strconv.Itoa(pid), c.Nom)
			utils.Display_d("main", "message msg reçu: "+sndmsg, c.Nom)
			fmt.Println(rcvmsg)
		}
	}
	if c.Snapshot.Couleur == utils.Rouge && sender != c.Nom {
		senderCouleur := utils.Findval(rcvmsg, "couleur", c.Nom)
		if senderCouleur == "" {
			senderCouleur = string(utils.Blanc) 
		}
		c.ReceptionMsgAppDeCtrl(rcvmsg, utils.Couleur(senderCouleur))
	}
}

func (c *utils.Controller) IsFromApp(rcvmsg string) bool {
	sndmsg := utils.Findval(rcvmsg, "msg", c.Nom)
	if sndmsg == "" {
		return true //si ce n'est pas formaté, ça veut dire qu'on récupère le message de l'app
	}
	return false //sinon de l'app
}

func (c *utils.Controller) HandleMessage() {
	var rcvmsg string
	for {
		fmt.Scanln(&rcvmsg)

		time.Sleep(50 * time.Millisecond) //temps d'attente du backend
		utils.Display_d("main", "received : "+rcvmsg, c.Nom)

		/*MESSAGE DE L'APP*/
		if c.IsFromApp(rcvmsg) {
			c.handleAppMessage(rcvmsg)
			/*MESSAGE DU CTRL*/
		} else {
			c.handleCtrlMessage(rcvmsg)
		}
	}
}

// reception d'un msg état
func (c *utils.Controller) ReceptionMsgEtat(etatRecu map[string]interface{}, bilanRecu int) {
    if c.Snapshot.Initiateur {
        // Fusionner avec l'état global
        for k, v := range etatRecu {
            c.Snapshot.EtatGlobal[k] = v
        }
        
        // Décrémenter le nombre d'états attendus
        c.Snapshot.NbEtatAttendu--
        
        // Incrémenter le nombre de messages prépost attendus
        c.Snapshot.NbMessagePrepostAttendu += bilanRecu
        
        // Vérifier si l'instantané est terminé
        c.VerifierFinSnapshot()
    } else {
        // Si non initiateur envoyer message sur l'anneau
        etatMsg := utils.EtatMessage{
            EtatLocal: etatRecu,
            Bilan:     bilanRecu,
        }
        c.EnvoyerSurAnneau(utils.EtatMsg, etatMsg)
    }
}

//reception msg prépost
func (c *utils.Controller) ReceptionMsgPrepost(message string) {
    if c.Snapshot.Initiateur {
        // Ajouter le message à l'état global
        prepostKey := "prepost_" + strconv.Itoa(len(c.Snapshot.EtatGlobal))
        c.Snapshot.EtatGlobal[prepostKey] = message
        
        // Décrémenter le nombre de messages prépost attendus
        c.Snapshot.NbMessagePrepostAttendu--
        
        // Vérifier si l'instantané est terminé
        c.VerifierFinSnapshot()
    } else {
        // Si non initiateur, simplement transmettre le message sur l'anneau
        c.EnvoyerSurAnneau(utils.PrePost, message)
    }
}
//reception msg app de ctl
func (c *utils.Controller) ReceptionMsgAppDeCtrl(message string, couleurRecue utils.Couleur) {
    // Décrémenter le bilan de 1
    c.Snapshot.Bilan--
    
    if couleurRecue == utils.Rouge && c.Snapshot.Couleur == utils.Blanc {
        // Le site devient rouge
        c.Snapshot.Couleur = utils.Rouge
        
        // Mettre à jour l'état local et l'envoyer sur l'anneau
        c.Snapshot.UpdateEtatLocal(c)
        etatMsg := utils.EtatMessage{
            EtatLocal: c.Snapshot.EtatLocal,
            Bilan:     c.Snapshot.Bilan,
        }
        c.EnvoyerSurAnneau(utils.EtatMsg, etatMsg)
    } else if couleurRecue == utils.Blanc && c.Snapshot.Couleur == utils.Rouge {
        // C'est un message prépost
        c.EnvoyerSurAnneau(utils.PrePost, message)
    }
    
    // Transmettre le message à l'application
    c.ForwardToApp(message)
}

func (c *utils.Controller) handleSnapshotMessage(msg string) {
	sender := utils.Findval(msg, "msg", c.Nom)


	if c.Snapshot.Couleur == utils.Blanc {
		// Premier marqueur reçu : devenir rouge et sauvegarder l'état local
		c.Snapshot.Couleur = utils.Rouge

		etat := map[string]interface{}{
			"Horloge":      c.Horloge,
			"InSection":    c.IsInSection,
			"DoctorsCount": CopyDoctorsCount(), // à adapter
		}
		c.Snapshot.EtatGlobal[c.Nom] = etat
		c.Snapshot.NbEtatAttendu = N - 1

		utils.Display_e("SNAPSHOT", fmt.Sprintf("%s devient rouge suite à snapshot reçu de %s", c.Nom, sender), c.Nom)

		// Propagation du marqueur aux autres
		for i := 1; i <= N; i++ {
			target := fmt.Sprintf("ctrl_%d", i)
			if target == c.Nom || target == sender {
				continue
			}

			msgSnapshot := utils.Msg_format("type", "snapshot") +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("VC", Encodehorloge(c.VectorClock)) +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge))

			fmt.Println(msgSnapshot)
		}
	} else {
		// Si déjà rouge, on ignore ce marqueur (état déjà pris)
		utils.Display_e("SNAPSHOT", fmt.Sprintf("Marqueur reçu de %s, mais état déjà pris.", sender), c.Nom)
	}
}