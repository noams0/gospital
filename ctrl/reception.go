package main

import (
	"fmt"
	"gospital/utils"
	"strconv"
	"strings"
	"time"
)

func (c *Controller) HandleMessage() {
	/*Boucle principale de réception et dispatch des messages*/
	var rcvmsg string
	for {
		fmt.Scanln(&rcvmsg)

		time.Sleep(c.Speed * time.Millisecond) //temps d'attente du backend
		utils.Display_d("main", "received : "+rcvmsg, c.Nom)
		net := utils.Findval(rcvmsg, "net", c.Nom)
		if net != "1" { //ON VERIFIE QUE CE N'EST PAS DU BRUIT DE NET
			if utils.Findval(rcvmsg, "new_site", c.Nom) != "" {
				if utils.ExtractIDt(c.NomCourt) != utils.Findval(rcvmsg, "new_site", c.Nom) {
					msg := "new_site" + utils.Findval(rcvmsg, "new_site", c.Nom)
					fmt.Println(msg)
					rcvmsg = utils.StripNetFields(rcvmsg)
					fmt.Println(rcvmsg)
				}
			} else {

				/*MESSAGE DE L'APP*/
				rcvmsg = utils.StripNetFields(rcvmsg)
				if c.IsFromApp(rcvmsg) {
					c.handleAppMessage(rcvmsg)
					/*MESSAGE DU CTRL*/
				} else {
					c.handleCtrlMessage(rcvmsg)
				}
			}
		} else {
			utils.Display_f("BRUIT", "BRUIT NET", c.Nom)
		}

	}
}

func (c *Controller) IsFromApp(rcvmsg string) bool {
	sndmsg := utils.Findval(rcvmsg, "msg", c.Nom)
	if sndmsg == "" {
		return true //si ce n'est pas formaté, ça veut dire qu'on récupère le message de l'app
	}
	return false //sinon du ctrl
}

func (c *Controller) handleSnapshotMessage(msg string) {
	/* Gère la réception d'un message de snapshot*/
	rcvVC := utils.DecodeVC(utils.Findval(msg, "vc", c.Nom))
	c.VectorClock = utils.MaxVC(c.VectorClock, rcvVC, c.NomCourt)
	sender := utils.Findval(msg, "sender", c.Nom)
	if sender == c.Nom {
		return
	}

	if c.Snapshot.Couleur == Blanc {
		//on ne le traite que si on ne la pas déjà traité
		fmt.Println("askForState")
		c.EnvoyerSurAnneau(SnapshotMsg, msg)
	}

}

func (c *Controller) handleAppMessage(rcvmsg string) {
	/* traite les messages en provenance de l'application */
	type_msg := utils.Findval(rcvmsg, "type", c.Nom)
	utils.Display_d("main", "TYPE de la demande en provenance de l'app : "+type_msg, c.Nom)

	switch type_msg {
	case "speed":
		delay := utils.Findval(rcvmsg, "delay", c.Nom)
		delayInt, _ := strconv.Atoi(delay)
		stderr.Println(delayInt)
		c.Speed = time.Duration(delayInt)
		stderr.Println(c.Speed)
	case "demandeSC":
		c.Horloge++
		c.VectorClock = utils.IncVC(c.VectorClock, c.NomCourt)
		c.Tab[c.Nom] = EtatReqSite{
			TypeRequete: Requete,
			Horloge:     c.Horloge,
			VectorClock: c.VectorClock,
		}
		utils.Display_f("demandeSC", "Demande de SC locale, horloge : "+strconv.Itoa(c.Horloge), c.Nom)
		fmt.Println(
			utils.Msg_format("type", string(Requete)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", "1") +
				utils.Msg_format("couleur", string(c.Snapshot.Couleur)) +
				c.Msg_Horloge())
	case "finSC":
		c.Horloge++
		c.VectorClock = utils.IncVC(c.VectorClock, c.NomCourt)
		c.Tab[c.Nom] = EtatReqSite{
			TypeRequete: Liberation,
			Horloge:     c.Horloge,
			VectorClock: c.VectorClock,
		}
		c.IsInSection = false
		utils.Display_f("finSC", "Fin de SC locale, horloge : "+strconv.Itoa(c.Horloge), c.Nom)
		newData := ""
		newData = utils.Findval(rcvmsg, "new_data", c.Nom) // ex: |app_1=4|app_2=3|app_3=7
		fmt.Println(
			utils.Msg_format("type", "liberation") +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", "finSC") +
				c.Msg_Horloge() +
				utils.Msg_format("couleur", string(c.Snapshot.Couleur)) +
				utils.Msg_format("new_data", newData),
		)
	case "send":
		c.Horloge++
		c.VectorClock = utils.IncVC(c.VectorClock, c.NomCourt)
		var destApp string = utils.Findval(rcvmsg, "destinator", c.Nom)
		var destCtrl string = utils.App_to_ctrl(destApp)
		utils.Display_f("destinator :", destCtrl, c.Nom)
		fmt.Println(
			utils.Msg_format("type", "send") +
				utils.Msg_format("destinator", destCtrl) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", "send") +
				utils.Msg_format("couleur", string(c.Snapshot.Couleur)) +
				c.Msg_Horloge())
	case "yourState":
		etat_local := utils.Findval(rcvmsg, "etat_local", c.Nom)
		etat_local_full := map[string]string{c.Nom: etat_local}
		converted := make(map[string]interface{})

		for k, v := range etat_local_full {
			converted[k] = v
		}
		etatMsg := EtatMessage{
			EtatLocal: converted,
			Bilan:     c.Snapshot.Bilan,
		}
		c.EnvoyerSurAnneau(EtatMsg, etatMsg)

	case "snapshot":
		sender := utils.Findval(rcvmsg, "sender", c.Nom)
		if strings.HasPrefix(sender, "app") {
			// Demande locale de l'app
			if c.SnapshotEnCours {
				utils.Display_f("snapshot", "Un snapshot est déjà en cours. Requête ignorée.", c.Nom)
			} else {
				c.SnapshotEnCours = true
				c.Horloge++
				utils.Display_f("snapshot", "Début du snapshot, horloge : "+strconv.Itoa(c.Horloge), c.Nom)
				c.DebutSnapshot()
			}
		} else {
			// Marqueur reçu d'un autre contrôleur
			c.handleSnapshotMessage(rcvmsg)
		}

	default:
	case "debutSC":
	case "receive":
		utils.Display_f("NOT", "for me", c.Nom)
	}

}

func (c *Controller) handleCtrlMessage(rcvmsg string) {
	/* traite les messages en provenance d'un autre contrôleur */
	rcvVC := utils.DecodeVC(utils.Findval(rcvmsg, "vc", c.Nom))
	c.VectorClock = utils.MaxVC(c.VectorClock, rcvVC, c.NomCourt)
	rcvHLG, _ := strconv.Atoi(utils.Findval(rcvmsg, "hlg", c.Nom))
	sndmsg := utils.Findval(rcvmsg, "msg", c.Nom)

	sender := utils.Findval(rcvmsg, "sender", c.Nom)

	c.Horloge = max(rcvHLG, c.Horloge) + 1

	msg_type := utils.Findval(rcvmsg, "type", c.Nom)
	switch msg_type {
	case "new_data":
		utils.Display_f("NOT", "for me", c.Nom)
	case string(Requete):
		if sender != *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrêt
			c.Tab[sender] = EtatReqSite{
				TypeRequete: Requete,
				Horloge:     rcvHLG,
				VectorClock: rcvVC,
			}
			utils.Display_f(string(Requete), "Requête reçue de "+sender+" | horloge="+strconv.Itoa(c.Horloge), c.Nom)
			//envoyer( [accusé] hi ) à Sj
			utils.Display_f(string(Requete), rcvmsg, c.Nom)
			utils.Display_f(string(Requete), fmt.Sprintf("mon tab %#v", c.Tab), c.Nom)
			fmt.Println(rcvmsg)

			fmt.Println(
				utils.Msg_format("destinator", sender) +
					utils.Msg_format("msg", "ack") +
					utils.Msg_format("type", "ack") +
					utils.Msg_format("sender", c.Nom) +
					utils.Msg_format("couleur", string(c.Snapshot.Couleur)) +
					c.Msg_Horloge())
			if c.Tab[c.Nom].TypeRequete == Requete && !c.IsInSection {
				if isFirstRequest(c.Tab, c.Nom, c.Tab[c.Nom].Horloge) {
					c.IsInSection = true
					utils.Display_f("SC", "\n ======================", c.Nom)
					utils.Display_f("SC", "Entrée en SC autorisée", c.Nom)
					utils.Display_f("SC", "\n ======================", c.Nom)
					fmt.Print("debutSC\n")
					utils.Display_e("je veux passer mon tab", fmt.Sprintf("%#v", c.Tab), c.Nom)
					fmt.Println(TabToString(c.Tab))
				}
			}
		}
	case string(Liberation):
		if sender != *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrêt
			c.Tab[sender] = EtatReqSite{
				TypeRequete: Liberation,
				Horloge:     rcvHLG,
				VectorClock: rcvVC,
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
			if c.Tab[c.Nom].TypeRequete == Requete && !c.IsInSection {
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
			if c.Tab[sender].TypeRequete != Requete {
				c.Tab[sender] = EtatReqSite{
					TypeRequete: Accuse,
					Horloge:     rcvHLG,
					VectorClock: rcvVC,
				}
			}
			utils.Display_f("Accusé", "Accusé reçue de "+sender+" | horloge="+strconv.Itoa(c.Horloge), c.Nom)
			utils.Display_f("Accusé", fmt.Sprintf("mon c.Tab %#v", c.Tab), c.Nom)

			//envoyer( [accusé] hi ) à Sj
			if c.Tab[c.Nom].TypeRequete == Requete && !c.IsInSection {
				utils.Display_f("TENTATIVE", "Je vais tenter de voir si je suis le premier", c.Nom)

				if isFirstRequest(c.Tab, c.Nom, c.Tab[c.Nom].Horloge) {
					c.IsInSection = true
					utils.Display_f("SC", "\n ======================", c.Nom)
					utils.Display_f("SC", "Entrée en SC autorisée", c.Nom)
					utils.Display_f("SC", "\n ======================", c.Nom)
					fmt.Print("debutSC\n")
					utils.Display_e("je veux passer mon tab", fmt.Sprintf("%#v", c.Tab), c.Nom)
					fmt.Println(TabToString(c.Tab))
				}
			}
		} else {
			fmt.Println(rcvmsg)
		}

	case "send":
		//On check que l'on n'est pas sur un message prepost
		couleur := utils.Findval(rcvmsg, "couleur", c.Nom)
		if couleur == string(Blanc) && c.Snapshot.Couleur == Rouge {
			stderr.Println("PREPOST")
			msg := fmt.Sprintf("from %s to %s", sender, utils.Findval(rcvmsg, "destinator", c.Nom))
			prepost, ok := c.Snapshot.EtatGlobal["PREPOST"].([]string)
			if !ok {
				// Soit la clé n'existe pas, soit ce n'est pas un []string — on initialise
				prepost = []string{}
			}
			prepost = append([]string{msg}, prepost...)
			c.Snapshot.EtatGlobal["PREPOST"] = prepost
		}
		if utils.Findval(rcvmsg, "destinator", c.Nom) == *p_nom {
			utils.Display_f("send", "send pour oim"+rcvmsg, c.Nom)
			fmt.Println("receive")
		} else {
			utils.Display_f("send", "send pas pour oim"+rcvmsg, c.Nom)
			fmt.Println(rcvmsg)
		}
	case "snapshot":
		stderr.Println("je recois un snapshot")
		if !c.SnapshotEnCours {
			c.SnapshotEnCours = true
			c.DebutSnapshot()
		}
	case string(EtatMsg):
		c.ReceptionMsgEtat(rcvmsg)
	case string(PrePost):
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

}
