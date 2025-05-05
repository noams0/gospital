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
	var rcvmsg string
	var vectorClock = make(map[string]int)

	var sndmsg string
	flag.Parse()
	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())

	for {
		fmt.Scanln(&rcvmsg)
		time.Sleep(1 * time.Second)
		display_d("main", "received : "+rcvmsg)
		rcvVC := utils.DecodeVC(findval(rcvmsg, "hlg"))

		sndmsg = findval(rcvmsg, "msg")
		if sndmsg == "" { //si ce n'est pas formaté, ça veut dire qu'on récupère le message de l'app
			switch rcvmsg {
			case "demandeSC":
                                acksReceived = make(map[string]bool) // Réinitialisation avant une nouvelle demande de SC
				vectorClock[*p_nom]++
				tab[*p_nom] = Horloge{
					Type: Requete,
					VC:   utils.CloneVC(vectorClock),
				}
				display_f("demandeSC", "Demande de SC locale, horloge : "+utils.EncodeVC(vectorClock))
				// envoi de la requête aux autre processus
                                fmt.Println(msg_format("type", "request") + msg_format("sender", nom) + msg_format("msg", "demandeSC") + msg_format("hlg", utils.EncodeVC(vectorClock)))
			case "finSC":
				vectorClock[*p_nom]++
				tab[*p_nom] = Horloge{
					Type: Requete,
					VC:   utils.CloneVC(vectorClock),
				}
				display_f("finSC", "Fin de SC locale, horloge : "+utils.EncodeVC(vectorClock))
				// envoi de la libération aux autres processus

                                fmt.Println(msg_format("type", "liberation") + msg_format("sender", nom) + msg_format("msg", "finSC") + msg_format("hlg", utils.EncodeVC(vectorClock)))
			default:
				fmt.Println(msg_format("sender", nom) + msg_format("msg", "default") + msg_format("hlg", utils.EncodeVC(vectorClock)))
			}
			//sinon, c'est un message provenant d'un ctrl
		} else {
                        if findval(rcvmsg, "sender") == nom {
				display_e("main", "Arret du message :"+rcvmsg)
				continue
			}
			if len(rcvVC) != 0 {
				// Mise à jour de l'horloge vectorielle
				for k, v := range rcvVC {
					if v > vectorClock[k] {
						vectorClock[k] = v
					}
				}
				vectorClock[*p_nom]++
			}
			msg_type := findval(rcvmsg, "type")
			sender := findval(rcvmsg, "sender")
			switch msg_type {
			case "request":
      				// Ajout de la requête à la file
				tab[sender] = Horloge{Type: Requete, VC: utils.CloneVC(rcvVC)}
				fmt.Println(msg_format("type", "ack") + msg_format("sender", nom) + msg_format("hlg", utils.EncodeVC(vectorClock)))
			case "liberation":
				// Suppression de la requête du processus libéré
				tab[sender] = Horloge{Type: Liberation, VC: utils.CloneVC(rcvVC)}
			case "ack":
				// Marque qu'un ack a été reçu
				acksReceived[sender] = true
			}
			if tab[*p_nom].Type == Requete && isFirstRequest(tab, nom, tab[*p_nom].VC) && allAcksReceived() {
				display_f("SC", "Accès à la section critique autorisé")
				fmt.Print("débutSC\n")
			}

		}
	}
}

func isFirstRequest(tab map[string]Horloge, me string, myVC map[string]int) bool {
	for k, info := range tab {
		if k == me {
			continue
		}
		if info.Type == Requete {
			if less(info.VC, k, myVC, me) {
				return false
			}
		}
	}
	return true
}
// Vérifie si tous les acks ont été reçus
func allAcksReceived() bool {
	for _, received := range acksReceived {
		if !received {
			return false
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

type MessageType string

const (
	Requete    MessageType = "requête"
	Liberation MessageType = "libération"
	Accuse     MessageType = "ack"
)

type Horloge struct {
	Type MessageType
	VC   map[string]int
}

var tab map[string]Horloge = make(map[string]Horloge)
var acksReceived = make(map[string]bool) // suivi des acks reçus
