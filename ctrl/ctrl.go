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
		if len(rcvVC) != 0 {
			display_d("main", fmt.Sprintf("horloge reçue : %#v", rcvVC))
			for k, v := range rcvVC {
				if _, ok := vectorClock[k]; !ok {
					vectorClock[k] = 0
				}
				if v > vectorClock[k] {
					vectorClock[k] = v
				}
			}
			vectorClock[*p_nom]++
			display_e("main", "Nouvelle horloge :"+utils.EncodeVC(vectorClock))
		} else {
			vectorClock[*p_nom]++
		}
		sndmsg = findval(rcvmsg, "msg")
		if sndmsg == "" { //si ce n'est pas formaté, ça veut dire qu'on récupère le message de l'app
			fmt.Println(msg_format("sender", nom) + msg_format("msg", rcvmsg) + msg_format("hlg", utils.EncodeVC(vectorClock)))
		} else {
			if findval(rcvmsg, "sender") == *p_nom+"-"+strconv.Itoa(pid) { // Si le message a fait un tour, il faut qu'il s'arrête
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
