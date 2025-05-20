package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gospital/utils"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var N = 3

var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)



func NewController(nomcourt, nom string) *Controller {
	return &Controller{
		Nom:         nom,
		NomCourt:    nomcourt,
		Horloge:     0,
		VectorClock: utils.InitVC(N),
		Tab:         make(map[string]EtatReqSite),
		IsInSection: false,
		Snapshot:    *NewSnapshot(),
		Speed:       50,
	}
}

var p_nom *string = flag.String("n", "ecrivain", "nom")

func (c *Controller) Msg_Horloge() string {
	msg := utils.Msg_format("hlg", strconv.Itoa(c.Horloge))
	c.VectorClock = utils.IncVC(c.VectorClock, c.NomCourt)
	msg += utils.Msg_format("vc", utils.EncodeVC(c.VectorClock))
	return msg
}

func (c *Controller) handleAppMessage(rcvmsg string) {
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

func TabToString(tab map[string]EtatReqSite) string {
	var result string = "TAB_REQ"

	for k, v := range tab {
		result += fmt.Sprintf("%s : Horloge=%d (%s), Type=%s,", k, v.Horloge, strings.ReplaceAll(utils.EncodeVC(v.VectorClock), ",", " "), v.TypeRequete)
		utils.Display_f("TEST", utils.EncodeVC(v.VectorClock), "XXX")
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

		time.Sleep(c.Speed * time.Millisecond) //temps d'attente du backend
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
			continue
		}
		if info.Horloge < h {
			return false
		} else if info.Horloge == h { //si c'est égalité alors on teste le nom du controleur, le plus petit l'emporte
			if !IsCtrlNumberLess(me, k) {
				return false
			}

		}
	}
	return true
}
func IsCtrlNumberLess(nom1, nom2 string) bool {
	getCtrlNumber := func(nom string) int {
		parts := strings.Split(nom, "_")
		if len(parts) < 2 {
			return -1
		}
		numStr := strings.Split(parts[1], "-")[0]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return -1
		}
		return num
	}

	x := getCtrlNumber(nom1)
	y := getCtrlNumber(nom2)

	return x < y
}

// initialisation de la sauvegarde
func NewSnapshot() *Snapshot {
	return &Snapshot{
		Couleur:                 Blanc,
		EtatGlobal:              make(map[string]interface{}),
		EtatLocal:               make(map[string]interface{}),
		Initiateur:              false,
		NbEtatAttendu:           N,
		Bilan:                   0,
		NbMessagePrepostAttendu: 0,
		EtatEnvoye:              false,
	}
}

// maj état local
func (s *Snapshot) UpdateEtatLocal(c *Controller) {
	s.EtatLocal = map[string]interface{}{
		"Horloge":      c.Horloge,
		"InSection":    c.IsInSection,
		"DoctorsCount": s.EtatLocal,
		"Tab":          c.Tab,
	}
}

// début
func (c *Controller) DebutSnapshot() {
	c.Snapshot.Couleur = Rouge
	c.Snapshot.Initiateur = true
	c.Snapshot.NbEtatAttendu = N // Pas N - 1 car on attent la valeur locale
	c.Snapshot.NbMessagePrepostAttendu = c.Snapshot.Bilan
	c.Snapshot.EtatEnvoye = true
	fmt.Println("askForState")
	msg := utils.Msg_format("type", "snapshot") +
		utils.Msg_format("sender", c.Nom) +
		utils.Msg_format("msg", "1") + //IMPORTANT POUR DIRE QUE CA VIENT DE APP
		c.Msg_Horloge() +
		utils.Msg_format("couleur", string(c.Snapshot.Couleur)) //ROUGE 

	c.EnvoyerSurAnneau(SnapshotMsg, msg)

}

// reception d'un msg état
func (c *Controller) ReceptionMsgEtat(rcvmsg string) {
	etatRecu := utils.Findval(rcvmsg, "etat", c.Nom)
	bilanRecu, _ := strconv.Atoi(utils.Findval(rcvmsg, "bilan", c.Nom))
	sender := utils.Findval(rcvmsg, "sender", c.Nom)
	if c.Snapshot.Initiateur {
		c.Snapshot.EtatGlobal[sender] = etatRecu
		c.Snapshot.NbEtatAttendu--
		stderr.Println(c.Snapshot.EtatGlobal[sender])
		c.Snapshot.NbMessagePrepostAttendu += bilanRecu
		c.VerifierFinSnapshot()
	} else {
	}
	if sender == c.Nom {
		return
	}
	c.EnvoyerSurAnneau("FORWARD", rcvmsg)
}

// reception msg prépost
func (c *Controller) ReceptionMsgPrepost(message string) {

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
		c.EnvoyerSurAnneau(PrePost, message)
	}
}

// reception msg app de ctl
func (c *Controller) ReceptionMsgAppDeCtrl(message string, couleurRecue Couleur) {
	c.Snapshot.Bilan--
	sender := utils.Findval(message, "sender", c.Nom)
	if sender == c.Nom {
		return
	}

	if couleurRecue == Rouge && c.Snapshot.Couleur == Blanc {
		c.Snapshot.Couleur = Rouge
		c.Snapshot.UpdateEtatLocal(c)

		if !c.Snapshot.EtatEnvoye {
			etatMsg := EtatMessage{
				EtatLocal: c.Snapshot.EtatLocal,
				Bilan:     c.Snapshot.Bilan,
			}
			c.EnvoyerSurAnneau(EtatMsg, etatMsg)
		}
	} else if couleurRecue == Blanc && c.Snapshot.Couleur == Rouge {
		// message en transit
		c.EnvoyerSurAnneau(PrePost, message)
	}

	c.ForwardToApp(message)
}

func (c *Controller) ForwardToApp(message string) {

	// Décrémenter le bilan (message sortant)
	c.Snapshot.Bilan--

	// Construire le message à transmettre avec la couleur
	ctrlMsg := CtrlMessage{
		Type:    AppMsg,
		Contenu: message,
		Couleur: c.Snapshot.Couleur,
	}

	// Envoyer directement sur l'anneau
	c.EnvoyerSurAnneau(AppMsg, ctrlMsg)
}
func (c *Controller) handleSnapshotMessage(msg string) {
	rcvVC := utils.DecodeVC(utils.Findval(msg, "vc", c.Nom))
	c.VectorClock = utils.MaxVC(c.VectorClock, rcvVC, c.NomCourt)
	sender := utils.Findval(msg, "sender", c.Nom)
	if sender == c.Nom {
		return
	}

	if c.Snapshot.Couleur == Blanc { //on ne le traite que si on ne la pas déjà traité
		
		fmt.Println("askForState")

		
		c.EnvoyerSurAnneau(SnapshotMsg, msg)
	}

}



// Vérification de la fin du snapshot
func (c *Controller) VerifierFinSnapshot() {

	utils.Display_e("NB???", fmt.Sprintf("%d et %d", c.Snapshot.NbEtatAttendu, c.Snapshot.NbMessagePrepostAttendu), "ccc")
	if c.Snapshot.NbEtatAttendu <= 0 && c.Snapshot.NbMessagePrepostAttendu <= 0 {
		utils.Display_e("SNAPSHOT", "Snapshot terminé!", c.Nom)
		utils.Display_e("SNAPSHOT", fmt.Sprintf("État global final: %#v", c.Snapshot.EtatGlobal), c.Nom)
		etatBytes, _ := json.Marshal(c.Snapshot.EtatGlobal)
		etatStr := string(etatBytes)

		fmt.Println("endSnapshot", etatStr)
		utils.SaveSnapshot(c.NomCourt, etatStr, utils.EncodeVC(c.VectorClock))

		c.Snapshot = *NewSnapshot()
		c.SnapshotEnCours = false
	}
}

// Envoyer un message sur l'anneau
func (c *Controller) EnvoyerSurAnneau(msgType MessageType, content interface{}) {

	// Cette fonction construit un message et l'envoie sur l'anneau
	// Format du message dépend du type de message

	var msg string

	switch msgType {
	case EtatMsg:
		if etatMsg, ok := content.(EtatMessage); ok {

			etatBytes, err := json.Marshal(etatMsg.EtatLocal)
			if err != nil {
				log.Fatal("Erreur lors du marshalling de l'état local :", err)
			}
			etatStr := string(etatBytes)
			msg = utils.Msg_format("type", string(EtatMsg)) +
				utils.Msg_format("couleur", string(c.Snapshot.Couleur)) +
				utils.Msg_format("msg", "1") +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("etat", etatStr) +
				utils.Msg_format("bilan", strconv.Itoa(etatMsg.Bilan)) +
				c.Msg_Horloge()
		}
	case PrePost:

		if prepostMsg, ok := content.(string); ok {

			msg = utils.Msg_format("type", string(PrePost)) +
				utils.Msg_format("couleur", string(c.Snapshot.Couleur)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", prepostMsg) +
				c.Msg_Horloge()
		}
	case AppMsg:

		if appMsg, ok := content.(CtrlMessage); ok {
			msg = utils.Msg_format("type", string(AppMsg)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", appMsg.Contenu) +
				utils.Msg_format("couleur", string(appMsg.Couleur)) +
				c.Msg_Horloge()
		}
	case SnapshotMsg:

		if msgStr, ok := content.(string); ok {
			msg = msgStr
		}
	case "FORWARD":
		if msgStr, ok := content.(string); ok {
			msg = msgStr
		}
	}
	if msg != "" {
		utils.Display_f("SEND", fmt.Sprintf("%s - %s", msgType, msg), c.Nom)
		fmt.Println(msg)
	}
}

func main() {
	flag.Parse()

	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())
	ctrl := NewController(*p_nom, nom)
	ctrl.HandleMessage()
}
