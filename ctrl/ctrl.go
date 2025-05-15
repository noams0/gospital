package main

import (
	"flag"
	"fmt"
	"gospital/utils"
	"log"
	"os"
	"strconv"
	"time"
)

var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)

type MessageType string

const (
	Requete    MessageType = "request"
	Liberation MessageType = "liberation"
	Accuse     MessageType = "ack"
)

type EtatReqSite struct {
	Horloge     int
	TypeRequete MessageType
}

var tab = make(map[string]EtatReqSite)

type Controller struct {
	Nom         string
	Horloge     int
	VectorClock map[string]int
	Tab         map[string]EtatReqSite
	IsInSection bool
}

func NewController(nom string) *Controller {
	return &Controller{
		Nom:         nom,
		Horloge:     0,
		VectorClock: make(map[string]int),
		Tab:         make(map[string]EtatReqSite),
		IsInSection: false,
	}
}

var p_nom *string = flag.String("n", "ecrivain", "nom")

func (c *Controller) handleAppMessage(rcvmsg string) {
	type_msg := utils.Findval(rcvmsg, "type", c.Nom)
	utils.Display_d("main", "TYPE de la demande en provenance de l'app : "+type_msg, c.Nom)

	switch type_msg {
	case "demandeSC":
		c.Horloge++
		c.Tab[c.Nom] = EtatReqSite{
			TypeRequete: Requete,
			Horloge:     c.Horloge,
		}
		utils.Display_f("demandeSC", "Demande de SC locale, horloge : "+strconv.Itoa(c.Horloge), c.Nom)
		fmt.Println(
			utils.Msg_format("type", string(Requete)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", "1") +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge)))
	case "finSC":
		c.Horloge++
		c.Tab[c.Nom] = EtatReqSite{
			TypeRequete: Liberation,
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
	default:
		//fmt.Println(utils.Msg_format("sender", c.Nom) + utils.Msg_format("msg", rcvmsg) + utils.Msg_format("hlg", strconv.Itoa(c.Horloge)))
	case "debutSC":
	case "receive":
		utils.Display_f("NOT", "for me", c.Nom)
	}

}

func (c *Controller) handleCtrlMessage(rcvmsg string) {
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
	case string(Requete):
		if sender != *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrêt
			c.Tab[sender] = EtatReqSite{
				TypeRequete: Requete,
				Horloge:     rcvHLG,
			}
			utils.Display_f(string(Requete), "Requête reçue de "+sender+" | horloge="+strconv.Itoa(c.Horloge), c.Nom)
			//envoyer( [accusé] hi ) à Sj
			utils.Display_f(string(Requete), rcvmsg, c.Nom)
			utils.Display_f(string(Requete), fmt.Sprintf("mon tab %#v", c.Tab), c.Nom)
			fmt.Println(rcvmsg)

			fmt.Println(utils.Msg_format("destinator", sender) + utils.Msg_format("msg", "ack") + utils.Msg_format("type", "ack") + utils.Msg_format("sender", c.Nom) + utils.Msg_format("hlg", strconv.Itoa(c.Horloge)))
			if c.Tab[c.Nom].TypeRequete == Requete && !c.IsInSection {
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
	case string(Liberation):
		if sender != *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrêt
			c.Tab[sender] = EtatReqSite{
				TypeRequete: Liberation,
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

func TabToString(tab map[string]EtatReqSite) string {
	var result string = "TAB_REQ"

	for k, v := range tab {
		result += fmt.Sprintf("%s : Horloge=%d, Type=%s,", k, v.Horloge, v.TypeRequete)
	}
	return result
}

func (c *Controller) IsFromApp(rcvmsg string) bool {
	sndmsg := utils.Findval(rcvmsg, "msg", c.Nom)
	if sndmsg == "" {
		return true //si ce n'est pas formaté, ça veut dire qu'on récupère le message de l'app
	}
	return false //sinon de l'app
}

func (c *Controller) HandleMessage() {
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

func isFirstRequest(tab map[string]EtatReqSite, me string, h int) bool {
	for k, info := range tab {
		if k == me {
			//utils.Display_f("TENTATIVE", "c'est moi, je passe")
			continue
		}
		if info.Horloge < h {
			//utils.Display_f("TENTATIVE RATEE", fmt.Sprintf("ca passe pas pour %d >= %d", info.Horloge, h))

			return false
		}
		//utils.Display_f("TENTATIVE", fmt.Sprintf("ca passe pour %d >= %d", info.Horloge, h))

	}
	return true
}

func main() {
	flag.Parse()

	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())
	ctrl := NewController(nom)
	ctrl.HandleMessage()
}
