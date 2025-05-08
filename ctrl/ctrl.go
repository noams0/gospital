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
	Requete    MessageType = "requête"
	Liberation MessageType = "libération"
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
}

func NewController(nom string) *Controller {
	return &Controller{
		Nom:         nom,
		Horloge:     0,
		VectorClock: make(map[string]int),
		Tab:         make(map[string]EtatReqSite),
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
		time.Sleep(1 * time.Second)
		display_d("main", "received : "+rcvmsg)
		rcvVC := utils.DecodeVC(findval(rcvmsg, "hlg"))

		sndmsg = findval(rcvmsg, "msg")
		if sndmsg == "" { //si ce n'est pas formaté, ça veut dire qu'on récupère le message de l'app
			switch rcvmsg {
			case "demandeSC":
				c.Horloge++
				tab[*p_nom] = EtatReqSite{
					TypeRequete: Requete,
					Horloge:     c.Horloge,
				}
				display_f("demandeSC", "Demande de SC locale, horloge : "+strconv.Itoa(c.Horloge))
				fmt.Println(msg_format("type", "request") + msg_format("sender", c.Nom) + msg_format("msg", rcvmsg) + msg_format("hlg", strconv.Itoa(c.Horloge)))
			case "finSC":
				c.Horloge++
				tab[*p_nom] = EtatReqSite{
					TypeRequete: Liberation,
					Horloge:     c.Horloge,
				}
				display_f("finSC", "Fin de SC locale, horloge : "+strconv.Itoa(c.Horloge))
				//fmt.Println(msg_format("type", "liberation") + msg_format("sender", nom) + msg_format("msg", rcvmsg) + msg_format("hlg", utils.EncodeVC(vectorClock)))
			default:
				fmt.Println(msg_format("sender", c.Nom) + msg_format("msg", rcvmsg) + msg_format("hlg", strconv.Itoa(c.Horloge)))
			}

			//sinon, c'est un message provenant d'un ctrly
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
				display_e("main", "Nouvelle horloge :"+utils.EncodeVC(c.VectorClock))
			} else {
				c.VectorClock[*p_nom]++
			}

			msg_type := findval(rcvmsg, "type")
			sender := findval(rcvmsg, "sender")
			//display_f("TYPE", msg_type)
			switch msg_type {
			case "request":
				if sender != *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrêt
					tab[sender] = EtatReqSite{
						TypeRequete: Requete,
						Horloge:     c.Horloge,
					}
					display_f("request", "Requête reçue de "+sender+" | VC="+strconv.Itoa(c.Horloge))
					//envoyer( [accusé] hi ) à Sj
					fmt.Println(rcvmsg)
					display_f("request", rcvmsg)
					fmt.Println(msg_format("destinator", sender) + msg_format("msg", "ack") + msg_format("type", "ack") + msg_format("sender", c.Nom) + msg_format("hlg", strconv.Itoa(c.Horloge)))
					if tab[c.Nom].TypeRequete == "request" {
						if isFirstRequest(tab, c.Nom, tab[c.Nom].Horloge) {
							display_f("SC", "\n ======================")
							display_f("SC", "Entrée en SC autorisée")
							display_f("SC", "\n ======================")
							fmt.Print("débutSC\n")
						}
					}
				}
			case "liberation":
				tab[sender] = EtatReqSite{
					TypeRequete: Liberation,
					Horloge:     c.Horloge,
				}
				display_f("liberation", "Libération reçue de "+sender+" | VC="+strconv.Itoa(c.Horloge))
				//envoyer( [accusé] hi ) à Sj
				if tab[c.Nom].TypeRequete == "request" {
					if isFirstRequest(tab, c.Nom, tab[c.Nom].Horloge) {
						display_f("SC", "\n ======================")
						display_f("SC", "Entrée en SC autorisée")
						display_f("SC", "\n ======================")
						fmt.Print("débutSC\n")
					}
				}
				display_f("liberation", "libération reçue de "+sender+" | VC="+strconv.Itoa(c.Horloge))
			case "ack":
				if findval(rcvmsg, "destinator") == *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrête
					display_f("Accusé", "Accusé reçue de "+sender+" | VC="+strconv.Itoa(c.Horloge))
					//envoyer( [accusé] hi ) à Sj
					if tab[c.Nom].TypeRequete == "request" {
						if isFirstRequest(tab, c.Nom, tab[c.Nom].Horloge) {
							display_f("SC", "\n ======================")
							display_f("SC", "Entrée en SC autorisée")
							display_f("SC", "\n ======================")
							fmt.Print("débutSC\n")
						}
					}
				} else {
					display_f("ack", "message ack pas pour oim"+rcvmsg)
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
			continue
		}
		if info.TypeRequete == Requete {
			if info.Horloge < h {
				return false
			}
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
