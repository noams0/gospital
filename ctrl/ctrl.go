package main

import (
	"flag"
	"fmt"
	"gospital/utils"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

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

var tab map[string]EtatReqSite = make(map[string]EtatReqSite)

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

var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)

const fieldsep = "/"
const keyvalsep = "="

func msg_format(key string, val string) string {
	return fieldsep + keyvalsep + key + keyvalsep + val
}

func findval(msg string, key string) string {
	if len(msg) < 4 {
		display_w("findval", "message trop court : "+msg)
		return ""
	}
	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)

	for _, keyval := range tab_allkeyvals {
		if len(keyval) < 3 { // au moins 1 pour separateur, 1 pour key, 1 pour val
			display_w("findval", "message trop court : "+msg)
			continue
		}
		equ := keyval[0:1]
		tabkeyval := strings.SplitN(keyval[1:], equ, 2)
		if len(tabkeyval) != 2 {
			continue
		}
		if tabkeyval[0] == key {
			return tabkeyval[1]
		}
	}
	return ""
}

func display_d(where string, what string) {
	stderr.Printf("%s + [%.6s %d] %-8.8s : %s\n%s", utils.ColorBlue, *p_nom, pid, where, what, utils.ColorReset)
}

func display_w(where string, what string) {

	stderr.Printf("%s * [%.6s %d] %-8.8s : %s\n%s", utils.ColorYellow, *p_nom, pid, where, what, utils.ColorReset)
}

func display_e(where string, what string) {
	stderr.Printf("%s ! [%.6s %d] %-8.8s : %s\n%s", utils.ColorRed, *p_nom, pid, where, what, utils.ColorReset)
}

func display_f(where string, what string) {

	stderr.Printf("%s * [%.6s %d] %-8.8s : %s\n%s", utils.ColorPurple, *p_nom, pid, where, what, utils.ColorReset)
}

var p_nom *string = flag.String("n", "ecrivain", "nom")

func main() {
	flag.Parse()

	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())
	ctrl := NewController(nom)
	ctrl.HandleMessage()
}
func (c *Controller) HandleMessage() {
	var rcvmsg string

	var sndmsg string
	for {
		fmt.Scanln(&rcvmsg)
		time.Sleep(200 * time.Millisecond)
		display_d("main", "received : "+rcvmsg)
		rcvVC := utils.DecodeVC(findval(rcvmsg, "VC"))
		rcvHLG, _ := strconv.Atoi(findval(rcvmsg, "hlg"))
		sndmsg = findval(rcvmsg, "msg")
		if sndmsg == "" { //si ce n'est pas formaté, ça veut dire qu'on récupère le message de l'app
			if strings.HasPrefix(rcvmsg, "finSC") {
				c.Horloge++
				tab[c.Nom] = EtatReqSite{
					TypeRequete: Liberation,
					Horloge:     c.Horloge,
				}
				c.IsInSection = false
				display_f("finSC", "Fin de SC locale, horloge : "+strconv.Itoa(c.Horloge))
				newData := ""
				if len(rcvmsg) > len("finSC") && rcvmsg[len("finSC")] == '|' {
					newData = rcvmsg[len("finSC"):] // ex: |app_1=4|app_2=3|app_3=7
				}
				fmt.Println(
					msg_format("type", "liberation") +
						msg_format("sender", c.Nom) +
						msg_format("msg", "finSC") +
						msg_format("hlg", strconv.Itoa(c.Horloge)) +
						msg_format("new_data", newData),
				)
			} else if strings.HasPrefix(rcvmsg, "send") {
				c.Horloge++
				var destApp string = strings.TrimPrefix(rcvmsg, "send")
				var destCtrl string = utils.App_to_ctrl(destApp)
				display_f("destinator :", destCtrl)
				fmt.Println(msg_format("type", "send") + msg_format("destinator", destCtrl) + msg_format("sender", c.Nom) + msg_format("msg", "send") + msg_format("hlg", strconv.Itoa(c.Horloge)))

			} else {

				switch rcvmsg {
				case "debutSC":
				case "receive":
					display_f("NOT", "for me")
				case "demandeSC":
					c.Horloge++
					tab[c.Nom] = EtatReqSite{
						TypeRequete: Requete,
						Horloge:     c.Horloge,
					}
					display_f("demandeSC", "Demande de SC locale, horloge : "+strconv.Itoa(c.Horloge))
					fmt.Println(msg_format("type", string(Requete)) + msg_format("sender", c.Nom) + msg_format("msg", rcvmsg) + msg_format("hlg", strconv.Itoa(c.Horloge)))
				default:
					fmt.Println(msg_format("sender", c.Nom) + msg_format("msg", rcvmsg) + msg_format("hlg", strconv.Itoa(c.Horloge)))
				}

			}

			//sinon, c'est un message provenant d'un ctrl
		} else {
			if len(rcvVC) != 0 {
				display_d("main", fmt.Sprintf("horloge reçue : %#v", rcvVC))
				for k, v := range rcvVC {
					if _, ok := c.VectorClock[k]; !ok {
						c.VectorClock[k] = 0
					}
					if v > c.VectorClock[k] {
						c.VectorClock[k] = v
					}
				}
				c.VectorClock[*p_nom]++
				//display_e("main", "Nouvelle horloge :"+utils.Encodehorloge(c.VectorClock))
			} else {
				c.VectorClock[*p_nom]++
			}
			c.Horloge = max(rcvHLG, c.Horloge) + 1

			msg_type := findval(rcvmsg, "type")
			sender := findval(rcvmsg, "sender")
			//display_f("TYPE", msg_type)
			switch msg_type {
			case "new_data":
				display_f("NOT", "for me")
			case string(Requete):
				if sender != *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrêt
					tab[sender] = EtatReqSite{
						TypeRequete: Requete,
						Horloge:     rcvHLG,
					}
					display_f(string(Requete), "Requête reçue de "+sender+" | horloge="+strconv.Itoa(c.Horloge))
					//envoyer( [accusé] hi ) à Sj
					display_f(string(Requete), rcvmsg)
					display_f(string(Requete), fmt.Sprintf("mon tab %#v", tab))
					fmt.Println(rcvmsg)

					fmt.Println(msg_format("destinator", sender) + msg_format("msg", "ack") + msg_format("type", "ack") + msg_format("sender", c.Nom) + msg_format("hlg", strconv.Itoa(c.Horloge)))
					if tab[c.Nom].TypeRequete == Requete && !c.IsInSection {
						if isFirstRequest(tab, c.Nom, tab[c.Nom].Horloge) {
							c.IsInSection = true
							display_f("SC", "\n ======================")
							display_f("SC", "Entrée en SC autorisée")
							display_f("SC", "\n ======================")
							fmt.Print("debutSC\n")
						}
					}
				}
			case string(Liberation):
				if sender != *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrêt
					tab[sender] = EtatReqSite{
						TypeRequete: Liberation,
						Horloge:     rcvHLG,
					}
					new_data := findval(rcvmsg, "new_data")
					if new_data != "" {
						fmt.Println(
							msg_format("type", "new_data") +
								msg_format("new_data", new_data) + // ex : new_data|app_1=5|app_2=2|app_3=6
								msg_format("msg", "finSC"))
					}
					display_f("liberation", "Libération reçue de "+sender+" | horloge="+strconv.Itoa(c.Horloge))
					display_f("liberation", fmt.Sprintf("mon tab %#v", tab))
					fmt.Println(rcvmsg)
					//envoyer( [accusé] hi ) à Sj
					if tab[c.Nom].TypeRequete == Requete && !c.IsInSection {
						if isFirstRequest(tab, c.Nom, tab[c.Nom].Horloge) {
							c.IsInSection = true
							display_f("SC", "\n ======================")
							display_f("SC", "Entrée en SC autorisée")
							display_f("SC", "\n ======================")
							fmt.Print("debutSC\n")
						}
					}
					display_f("liberation", "libération reçue de "+sender+" | horloge="+strconv.Itoa(c.Horloge))
				}
			case "ack":
				if findval(rcvmsg, "destinator") == *p_nom+"-"+strconv.Itoa(pid) {
					if tab[sender].TypeRequete != Requete {
						tab[sender] = EtatReqSite{
							TypeRequete: Accuse,
							Horloge:     rcvHLG,
						}
					}
					display_f("Accusé", "Accusé reçue de "+sender+" | horloge="+strconv.Itoa(c.Horloge))
					display_f("Accusé", fmt.Sprintf("mon tab %#v", tab))

					//envoyer( [accusé] hi ) à Sj
					if tab[c.Nom].TypeRequete == Requete && !c.IsInSection {
						display_f("TENTATIVE", "Je vais tenter de voir si je suis le premier")

						if isFirstRequest(tab, c.Nom, tab[c.Nom].Horloge) {
							c.IsInSection = true
							display_f("SC", "\n ======================")
							display_f("SC", "Entrée en SC autorisée")
							display_f("SC", "\n ======================")
							fmt.Print("debutSC\n")
						}
					}
				} else {
					display_f("ack", "message ack pas pour oim"+rcvmsg)
					fmt.Println(rcvmsg)
				}

			case "send":
				if findval(rcvmsg, "destinator") == *p_nom {
					display_f("send", "send pour oim"+rcvmsg)

					fmt.Println("receive")
				} else {
					display_f("send", "send pas pour oim"+rcvmsg)
					fmt.Println(rcvmsg)
				}

			default:
				if sender == *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrête
					display_e("main", "Arret du message :"+rcvmsg)
					continue
				} else {
					display_e("main", "sender :"+findval(rcvmsg, "sender"))
					display_e("main", "sender :"+*p_nom+"-"+strconv.Itoa(pid))
					display_d("main", "message msg reçu: "+sndmsg)
					fmt.Println(rcvmsg)
				}
			}

		}
	}
}

func isFirstRequest(tab map[string]EtatReqSite, me string, h int) bool {
	for k, info := range tab {
		if k == me {
			display_f("TENTATIVE", "c'est moi, je passe")
			continue
		}
		if info.TypeRequete == Requete {
			if info.Horloge < h {
				display_f("TENTATIVE RATEE", fmt.Sprintf("ca passe pas pour %d >= %d", info.Horloge, h))

				return false
			}
			display_f("TENTATIVE", fmt.Sprintf("ca passe pour %d >= %d", info.Horloge, h))

		}
	}
	return true
}

func less(vc1 map[string]int, name1 string, vc2 map[string]int, name2 string) bool {
	// Compare (vc1, name1) < (vc2, name2)
	v1 := vc1[name1]
	v2 := vc2[name2]
	if v1 != v2 {
		return v1 < v2
	}
	return name1 < name2
}
