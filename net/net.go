package main

import (
	"bufio"
	"flag"
	"fmt"
	"gospital/utils"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Rule struct {
	From string
	To   string
}

func ParseRoute(s string) []Rule {
	var rules []Rule
	entries := strings.Split(s, ",")
	for _, entry := range entries {
		parts := strings.Split(entry, ":")
		if len(parts) == 2 {
			from := strings.TrimPrefix(parts[0], "from=")
			to := strings.TrimPrefix(parts[1], "to=")
			rules = append(rules, Rule{From: from, To: to})
		}
	}
	return rules
}

var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)
var p_nom *string = flag.String("n", "ecrivain", "nom")
var totalSites = flag.Int("total", 3, "Nombre total de sites")
var routeStr = flag.String("route", "", "table de routage")

type Net struct {
	Nom      string
	NomCourt string
	Speed    time.Duration
	NbSite   int
	Rules    []Rule
}

func NewNet(nomcourt, nom string, nb_site int, rules []Rule) *Net {
	/*Crée et initialise une nouvelle instance de Controller*/
	return &Net{
		Nom:      nom,
		NomCourt: nomcourt,
		NbSite:   nb_site,
		Rules:    rules,
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
	rules := ParseRoute(*routeStr)
	net := NewNet(*p_nom, nom, *totalSites, rules)

	utils.Display_n("NET", fmt.Sprintf("%#v", net.Rules), net.NomCourt)

	net.run()

}

func (net *Net) run() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		rcvmsg := scanner.Text()

		// Re-transmet sur stdout (anneau logique)
		utils.Display_n("NET, rec", rcvmsg, net.NomCourt)
		// Si pas concerné
		if net.IsNotForMe(rcvmsg) {
			utils.Display_n("NET, NON", "not for me", net.NomCourt)
			continue
		}

		// Extraire l’expéditeur (ex: "ctrl_1", "net_2", etc.)
		sender := utils.Findval(rcvmsg, "net_sender", net.NomCourt)
		if sender == "" {
			sender = "ctrl" // fallback si message local sans sender explicite
		}

		// Chercher le destinataire dans la table
		dest := ""
		for _, rule := range net.Rules {
			if rule.From == sender {
				dest = rule.To
				break
			}
		}
		if dest == "" {
			utils.Display_n("NET, NON", "aucune règle", net.NomCourt)
			continue
		}
		// Recomposer le message proprement
		msg := utils.StripNetFields(rcvmsg)
		if dest == "ctrl" {
			msg += utils.Msg_format("net", "0")
		} else if strings.HasPrefix(dest, "net_") {
			msg += utils.Msg_format("net", "1") + utils.Msg_format("net_destinator", dest) + utils.Msg_format("net_sender", net.NomCourt)
		} else {
			utils.Display_n("NET, NON", "destinataire inconnu", net.NomCourt)
			continue
		}

		fmt.Println(msg)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[NET] Erreur de lecture : %v\n", err)
	}

}
