package main

import (
	"bufio"
	"flag"
	"fmt"
	"gospital/utils"
	"log"
	"os"
	"os/exec"
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
	asking   bool
	asker    int
}

func NewNet(nomcourt, nom string, nb_site int, rules []Rule) *Net {
	/*Crée et initialise une nouvelle instance de Controller*/
	return &Net{
		Nom:      nom,
		NomCourt: nomcourt,
		NbSite:   nb_site,
		Rules:    rules,
		asking:   false,
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

//func (c *Net) IsNotForMe(rcvmsg string) bool {
//	net := utils.Findval(rcvmsg, "net", c.Nom)
//	if net == "0" {
//		return true
//	}
//	return false
//}

// A VERIFIER
func (c *Net) IsNotForMe(rcvmsg string) bool {

	net := utils.Findval(rcvmsg, "net", c.Nom)
	if net == "0" { // msg d'un autre net pour son ctrl
		return true
	} else {
		destinator := utils.Findval(rcvmsg, "net_destinator", c.Nom)
		ctrl := utils.Findval(rcvmsg, "msg", c.Nom)
		if net == "" && destinator == "" && ctrl != "" { //si pas de destinataire et que msg n'est pas vide, ça vient de ctrl qui veut transmettre message sur anneau
			return false
		}
		if net == "1" && destinator == c.NomCourt { // Si ça vient de net et que je suis le destinataire
			return false
		}
	}
	return true
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

func (net *Net) IsLeaf() bool {
	if len(net.Rules) == 2 && net.Rules[0].From == net.Rules[1].To && net.Rules[1].From == net.Rules[0].To {
		utils.Display_n("NET", "je suis une feuille", net.NomCourt)
		return true
	}
	return false
}

func updateRulesOnLeave(rules []Rule, leaver string) []Rule {
	var newRules []Rule
	var toReplace string
	for _, r := range rules {
		// On mémorise la destination de la règle sortante du site quittant
		if r.From == leaver {
			toReplace = r.To
			continue // on ne garde pas la règle sortante du site quittant
		}
		if r.From != leaver && r.To != leaver {
			newRules = append(newRules, r)
		}
	}
	for _, r := range rules {
		// Si ctrl envoyait vers le site quittant, on le redirige vers sa cible à lui
		if r.From == "ctrl" && r.To == leaver {
			newRules = append(newRules, Rule{From: "ctrl", To: toReplace})
			continue
		}
	}

	return newRules
}

func updateRulesOnLeaveSucc(rules []Rule, leaver string, new_succ string) []Rule {
	var newRules []Rule
	for _, r := range rules {
		// On mémorise la destination de la règle sortante du site quittant
		if r.From == leaver {
			newRules = append(newRules, Rule{From: new_succ, To: r.To})
		} else if r.To == leaver {
			newRules = append(newRules, Rule{From: r.From, To: new_succ})
		} else {
			newRules = append(newRules, r)
		}
	}
	return newRules
}

func updateRulesOnLeavePred(rules []Rule, leaver string, new_pred string) []Rule {
	var newRules []Rule
	for _, r := range rules {
		// On mémorise la destination de la règle sortante du site quittant
		if r.From == leaver {
			newRules = append(newRules, Rule{From: new_pred, To: r.To})
		} else if r.To == leaver {
			newRules = append(newRules, Rule{From: r.From, To: new_pred})
		} else {
			newRules = append(newRules, r)
		}
	}
	return newRules
}

func (net *Net) accept_site(rcvmsg string) {
	asker := strconv.Itoa(net.asker)
	sender := "net_" + asker
	net.asking = false
	var newRules []Rule
	destinataire := ""
	if len(net.Rules) == 2 {
		for _, r := range net.Rules {
			if r.From == "ctrl" {
				destinataire = r.To
				// On remplace cette règle par ctrl -> sender
				newRules = append(newRules, Rule{From: r.From, To: sender})
				// Et on ajoute la nouvelle règle sender -> destinataire
				newRules = append(newRules, Rule{From: sender, To: destinataire})
			} else {
				newRules = append(newRules, r)
			}
		}
	} else {
		if utils.Findval(rcvmsg, "type", net.NomCourt) == "append" {
			for _, r := range net.Rules {
				if r.From == "ctrl" {
					destinataire = r.To
					// On remplace par ctrl -> sender
					newRules = append(newRules, Rule{From: "ctrl", To: sender})
				} else {
					newRules = append(newRules, r)
				}
			}
			if destinataire != "" {
				// Ajoute la chaîne sender -> oldTarget
				newRules = append(newRules, Rule{From: sender, To: destinataire})
			}
			net.Rules = newRules
		}
	}
	net.Rules = newRules
	utils.Display_n("NET maj", fmt.Sprintf("%#v", net.Rules), net.NomCourt)

	msg := utils.Msg_format("new_site", utils.ExtractIDt(sender)) +
		utils.Msg_format("net", "1") +
		utils.Msg_format("msg", "1") +
		utils.Msg_format("net_destinator", destinataire) +
		utils.Msg_format("net_sender", net.NomCourt)
	utils.Display_n("NET", "HERE je vais envoyer :", net.NomCourt)
	utils.Display_n("NET", msg, net.NomCourt)
	fmt.Println("\n")
	fmt.Println(msg)
	msg = utils.Msg_format("site_accepted", utils.ExtractIDt(sender)) +
		utils.Msg_format("net", "1") +
		utils.Msg_format("msg", "1") +
		utils.Msg_format("net_destinator", sender) +
		utils.Msg_format("net_sender", net.NomCourt)
	utils.Display_n("NET", "HERE je vais envoyer :", net.NomCourt)
	utils.Display_n("NET", msg, net.NomCourt)
	fmt.Println("\n")
	fmt.Println(msg)

}

func (net *Net) askToJoin(rcvmsg string) {
	utils.Display_n("NET", "here", net.NomCourt)

	sender := utils.Findval(rcvmsg, "net_sender", net.NomCourt)
	dest := ""
	for _, r := range net.Rules {
		if r.From == "ctrl" {
			dest = r.To
		}
	}
	msg := utils.Msg_format("type", "askToJoin") +
		utils.Msg_format("asker", utils.ExtractIDt(sender)) +
		utils.Msg_format("net", "1") +
		utils.Msg_format("net_destinator", dest) +
		utils.Msg_format("net_sender", net.NomCourt)
	utils.Display_n("NET", "HERE je vais envoyer :", net.NomCourt)
	utils.Display_n("NET", msg, net.NomCourt)
	fmt.Println("\n")
	fmt.Println(msg)
	net.asking = true
	net.asker, _ = strconv.Atoi(utils.ExtractIDt(sender))
}

func main() {
	flag.Parse()

	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())
	rules := ParseRoute(*routeStr)
	net := NewNet(*p_nom, nom, *totalSites, rules)

	utils.Display_n("NET", fmt.Sprintf("%#v", net.Rules), net.NomCourt)
	if len(rules) == 2 &&
		rules[0].From == rules[1].To &&
		rules[0].To == rules[1].From {
		dest := rules[1].To
		msg := utils.Msg_format("type", "append") +
			utils.Msg_format("net", "1") +
			utils.Msg_format("net_destinator", dest) +
			utils.Msg_format("net_sender", net.NomCourt)
		fmt.Println(msg)
		scanner := bufio.NewScanner(os.Stdin)
	loop:
		for scanner.Scan() {
			rcvmsg := scanner.Text()
			if utils.Findval(rcvmsg, "type", net.NomCourt) == "site_accepted" {
				utils.Display_n("NET", "NBREAK LOOP	", net.NomCourt)
				break loop
			}
			utils.Display_n("NET", "WAITING", net.NomCourt)
		}
		utils.Display_n("NET", "NOUVELLEMENT AJOUTÉ DYNAMIQUEMENT", net.NomCourt)
	}
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
		if utils.Findval(rcvmsg, "type", net.NomCourt) == "askToJoin" {
			utils.Display_n("NET, YES", "askToJoin", net.NomCourt)
			sender := utils.Findval(rcvmsg, "net_sender", net.NomCourt)
			dest := ""
			for _, rule := range net.Rules {
				if rule.From == sender {
					dest = rule.To
					break
				}
			}
			if dest == "ctrl" {
				for _, rule := range net.Rules {
					if rule.From == "ctrl" {
						dest = rule.To
						break
					}
				}
			}
			if net.asking {
				asker := utils.Findval(rcvmsg, "asker", net.NomCourt)
				asker_int, _ := strconv.Atoi(asker)
				if asker_int == net.asker { //LE MESSAGE A FAIT UN TOUR COMPLET SANS Ê ARRETÉ
					net.accept_site(rcvmsg)
					return
				} else {
					if asker_int > net.asker {
						rcvmsg = utils.StripNetFields(rcvmsg)
						msg := rcvmsg +
							utils.Msg_format("net", "1") +
							utils.Msg_format("net_destinator", dest) +
							utils.Msg_format("net_sender", net.NomCourt)
						fmt.Println(msg)

					} else {
						utils.Display_n("NET", "BLOQUE", net.NomCourt)
					}
				}
			} else {
				rcvmsg = utils.StripNetFields(rcvmsg)
				msg := rcvmsg +
					utils.Msg_format("net", "1") +
					utils.Msg_format("net_destinator", dest) +
					utils.Msg_format("net_sender", net.NomCourt)
				fmt.Println(msg)
			}
			continue
		}

		if utils.Findval(rcvmsg, "type", net.NomCourt) == "askToQuit" {
			if net.IsLeaf() {
				utils.Display_n("NET", "leaf", net.NomCourt)
				var dest string
				if net.Rules[0].From == "ctrl" {
					dest = net.Rules[0].To
				} else {
					dest = net.Rules[0].From
				}
				msg := utils.Msg_format("net", "1") +
					utils.Msg_format("msg", "1") +
					utils.Msg_format("type", "leaf_leave") +
					utils.Msg_format("net_destinator", dest) +
					utils.Msg_format("net_sender", net.NomCourt)
				fmt.Println(msg)
				time.Sleep(1 * time.Second)
				pid_app := utils.Findval(rcvmsg, "pid_app", net.NomCourt)
				pid_ctrl := utils.Findval(rcvmsg, "pid_ctrl", net.NomCourt)
				pid_net := strconv.Itoa(os.Getpid())
				site_id := utils.ExtractIDt(net.NomCourt)
				cmd := exec.Command("./leave_site.sh", pid_app, pid_ctrl, pid_net, site_id)
				if err := cmd.Run(); err != nil {
					utils.Display_e("NET RUN", fmt.Sprintf("❌ Erreur lors de leave_site.sh : %s", err), net.NomCourt)
				} else {
					utils.Display_e("NET RUN", "succes leave_site.sh", net.NomCourt)
				}
			} else {
				var pred string
				var succ string
				for _, r := range net.Rules {
					if r.To == "ctrl" {
						pred = r.From
					}
					if r.From == "ctrl" {
						succ = r.To
					}
				}
				msg := utils.Msg_format("net", "1") +
					utils.Msg_format("msg", "1") +
					utils.Msg_format("type", "pred_leave") +
					utils.Msg_format("new_succ", succ) +
					utils.Msg_format("net_destinator", pred) +
					utils.Msg_format("net_sender", net.NomCourt)
				fmt.Println(msg)
				msg = utils.Msg_format("net", "1") +
					utils.Msg_format("msg", "1") +
					utils.Msg_format("type", "succ_leave") +
					utils.Msg_format("new_pred", pred) +
					utils.Msg_format("net_destinator", succ) +
					utils.Msg_format("net_sender", net.NomCourt)
				fmt.Println(msg)

				utils.Display_n("NET", "not leaf", net.NomCourt)
			}
			continue
		}

		//SINON POUR MOI

		if utils.Findval(rcvmsg, "type", net.NomCourt) == "pred_leave" {
			new_succ := utils.Findval(rcvmsg, "new_succ", net.NomCourt)
			utils.Display_n("NET succ", new_succ, net.NomCourt)

			sender := utils.Findval(rcvmsg, "net_sender", net.NomCourt)
			net.Rules = updateRulesOnLeaveSucc(net.Rules, sender, new_succ)
			utils.Display_n("NET maj", fmt.Sprintf("%#v", net.Rules), net.NomCourt)
		}
		if utils.Findval(rcvmsg, "type", net.NomCourt) == "succ_leave" {
			new_pred := utils.Findval(rcvmsg, "new_pred", net.NomCourt)
			sender := utils.Findval(rcvmsg, "net_sender", net.NomCourt)
			net.Rules = updateRulesOnLeavePred(net.Rules, sender, new_pred)
			utils.Display_n("NET maj", fmt.Sprintf("%#v", net.Rules), net.NomCourt)
		}

		if utils.Findval(rcvmsg, "type", net.NomCourt) == "leaf_leave" {
			sender := utils.Findval(rcvmsg, "net_sender", net.NomCourt)
			net.Rules = updateRulesOnLeave(net.Rules, sender)
			utils.Display_n("NET maj", fmt.Sprintf("%#v", net.Rules), net.NomCourt)
		}
		if utils.Findval(rcvmsg, "type", net.NomCourt) == "append" {
			go net.askToJoin(rcvmsg)
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
			utils.Display_n("NET", "HERE dest :", net.NomCourt)
			utils.Display_n("NET", dest, net.NomCourt)
			msg += utils.Msg_format("net", "1") + utils.Msg_format("net_destinator", dest) + utils.Msg_format("net_sender", net.NomCourt)
		} else {
			utils.Display_n("NET, NON", "destinataire inconnu", net.NomCourt)
			continue
		}
		utils.Display_n("NET", "HERE je vais envoyer :", net.NomCourt)
		utils.Display_n("NET", msg, net.NomCourt)
		fmt.Println(msg)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[NET] Erreur de lecture : %v\n", err)
	}

}
