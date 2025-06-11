package main

import (
	"bufio"
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
var p_nom *string = flag.String("n", "ecrivain", "nom")
var totalSites = flag.Int("total", 3, "Nombre total de sites")

type Net struct {
	Nom      string
	NomCourt string
	Speed    time.Duration
	NbSite   int
}

func NewNet(nomcourt, nom string, nb_site int) *Net {
	/*Crée et initialise une nouvelle instance de Controller*/
	return &Net{
		Nom:      nom,
		NomCourt: nomcourt,
		NbSite:   nb_site,
	}
}

func (c *Net) FromCtrlForNet(rcvmsg string) bool {
	sndmsg := utils.Findval(rcvmsg, "msg", c.Nom)
	if sndmsg != "" && !c.IsFromNet(rcvmsg) {
		return true
	}
	return false
}
func (c *Net) FromNetForCtrl(rcvmsg string) bool {
	if c.IsFromNet(rcvmsg) {
		return true
	}
	return false
}

func (c *Net) IsNotForMe(rcvmsg string) bool {
	net := utils.Findval(rcvmsg, "net", c.Nom)
	if net == "0" {
		return true
	}
	return false
}

func (c *Net) FromNetForNet(rcvmsg string) bool {
	sndmsg := utils.Findval(rcvmsg, "msg", c.Nom)
	if sndmsg == "" {
		return false
	}
	return true
}

func (c *Net) IsFromNet(rcvmsg string) bool {
	sndmsg := utils.Findval(rcvmsg, "net", c.Nom)
	if sndmsg == "" {
		return false
	}
	return true
}
func (c *Net) IsFromMe(rcvmsg string) bool {
	sender := utils.Findval(rcvmsg, "net_sender", c.Nom)
	if sender == c.NomCourt {
		return true
	}
	return false
}

func main() {
	flag.Parse()

	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())
	net := NewNet(*p_nom, nom, *totalSites)

	utils.Display_n("NET", "démarre", net.NomCourt)

	net.run()

}

func (net *Net) run() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		rcvmsg := scanner.Text()

		// Re-transmet sur stdout (anneau logique)
		utils.Display_n("NET, rec", rcvmsg, net.NomCourt)
		if net.IsNotForMe(rcvmsg) {
			utils.Display_n("NET, NON", "non", net.NomCourt)
		} else if net.FromCtrlForNet(rcvmsg) { //si le message vient du ctrl et n'est pas pour app
			msg := utils.StripNetFields(rcvmsg) + utils.Msg_format("net_sender", net.NomCourt) + utils.Msg_format("net", "1")
			fmt.Println(msg)
		} else if net.FromNetForCtrl(rcvmsg) {
			msg := utils.StripNetFields(rcvmsg) + utils.Msg_format("net", "0")
			msg = msg + utils.Msg_format("net", "0")
			fmt.Println(msg)
		} else {
			utils.Display_n("NET, NON", "non", net.NomCourt)

		}
		//else if net.FromNetForNet(rcvmsg) {
		//	//A VOIR PLUS TARD; PAS VRAI POUR L'INSTANT
		//}

	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[NET] Erreur de lecture : %v\n", err)
	}

}
