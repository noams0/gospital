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

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

var N = 3

var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)

var tab = make(map[string]EtatReqSite)


func NewController(nom string) *Controller {
	return &Controller{
		Nom:         nom,
		Horloge:     0,
		VectorClock: make(map[string]int),
		Tab:         make(map[string]EtatReqSite),
		IsInSection: false,
		Snapshot: *NewSnapshot(),
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
	case "snapshot":
    	if !c.SnapshotEnCours {
            c.SnapshotEnCours = true
            c.Horloge++
            utils.Display_f("snapshot", "Demande de snapshot reçue de l'application", c.Nom)
            c.DebutSnapshot()
        }	
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
    case "snapshot":
	    c.handleSnapshotMessage(rcvmsg)
    case string(EtatMsg):
		var etatRecu map[string]interface{}
		var bilanRecu int
		c.ReceptionMsgEtat(etatRecu, bilanRecu)
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
	if c.Snapshot.Couleur == Rouge && sender != c.Nom {
		senderCouleur := utils.Findval(rcvmsg, "couleur", c.Nom)
		if senderCouleur == "" {
			senderCouleur = string(Blanc) 
		}
		c.ReceptionMsgAppDeCtrl(rcvmsg, Couleur(senderCouleur))
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
		} else if info.Horloge == h { //si c'est égalité alors on teste le nom du controleur, le plus petit l'emporte
			if !IsCtrlNumberLess(me, k) {
				return false
			}

		}
		//utils.Display_f("TENTATIVE", fmt.Sprintf("ca passe pour %d >= %d", info.Horloge, h))

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

//initialisation de la sauvegarde
func NewSnapshot() *Snapshot {
    return &Snapshot{
        Couleur:                 Blanc,
        EtatGlobal:              make(map[string]interface{}),
		EtatLocal:               make(map[string]interface{}),
        Initiateur:              false,
        NbEtatAttendu:           N,
        Bilan:                   0,
        NbMessagePrepostAttendu: 0,
    }
}

//maj état local
func (s *Snapshot) UpdateEtatLocal(c *Controller) {
	s.EtatLocal = map[string]interface{}{
		"Horloge":      c.Horloge,
		"InSection":    c.IsInSection,
		"DoctorsCount": CopyDoctorsCount(),
		"Tab":          c.Tab,
	}
}
// début 
/*func (c *Controller) DebutSnapshot() {
    // Le site devient rouge
    c.Snapshot.Couleur = Rouge
    
    // Initialiser l'état global avec l'état local
    c.Snapshot.UpdateEtatLocal(c)
    c.Snapshot.EtatGlobal = c.Snapshot.EtatLocal
    
    // Ce site est l'initiateur
    c.Snapshot.Initiateur = true
    
    // Nombre d'états attendus = N-1
    c.Snapshot.NbEtatAttendu = N - 1
    
    //Le site initiateur donne son bilan
    c.Snapshot.NbMessagePrepostAttendu = c.Snapshot.Bilan
    
}*/
func (c *Controller) DebutSnapshot() {
	// Devenir rouge
	c.Snapshot.Couleur = Rouge
	c.Snapshot.Initiateur = true

	// Sauvegarde de l'état local (ex: horloge, SC, DoctorsCount)
	etat := map[string]interface{}{
		"Horloge":     c.Horloge,
		"InSection":   c.IsInSection,
		"DoctorsCount": CopyDoctorsCount(), // à adapter à ta logique
	}
	c.Snapshot.EtatGlobal[c.Nom] = etat
	c.Snapshot.NbEtatAttendu = N - 1

	utils.Display_e("SNAPSHOT", fmt.Sprintf("Début snapshot par %s", c.Nom), c.Nom)
	utils.Display_e("SNAPSHOT", fmt.Sprintf("État local sauvegardé : %#v", etat), c.Nom)
	// Envoi de messages de "marqueur" aux autres (ex: broadcast VC)
	for i := 1; i <= N; i++ {
		if fmt.Sprintf("ctrl_%d", i) == c.Nom {
			continue
		}
		msg := utils.Msg_format("type", "snapshot") +
			utils.Msg_format("sender", c.Nom) +
			utils.Msg_format("VC", Encodehorloge(c.VectorClock)) +
			utils.Msg_format("hlg", strconv.Itoa(c.Horloge))

		fmt.Println(msg)
	}
}


// reception d'un msg état
func (c *Controller) ReceptionMsgEtat(etatRecu map[string]interface{}, bilanRecu int) {
    if c.Snapshot.Initiateur {
        // Fusionner avec l'état global
        for k, v := range etatRecu {
            c.Snapshot.EtatGlobal[k] = v
        }
        
        // Décrémenter le nombre d'états attendus
        c.Snapshot.NbEtatAttendu--
        
        // Incrémenter le nombre de messages prépost attendus
        c.Snapshot.NbMessagePrepostAttendu += bilanRecu
        
        // Vérifier si l'instantané est terminé
        c.VerifierFinSnapshot()
    } else {
        // Si non initiateur envoyer message sur l'anneau
        etatMsg := EtatMessage{
            EtatLocal: etatRecu,
            Bilan:     bilanRecu,
        }
        c.EnvoyerSurAnneau(EtatMsg, etatMsg)
    }
}

//reception msg prépost
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
//reception msg app de ctl
func (c *Controller) ReceptionMsgAppDeCtrl(message string, couleurRecue Couleur) {
    // Décrémenter le bilan de 1
    c.Snapshot.Bilan--
    
    if couleurRecue == Rouge && c.Snapshot.Couleur == Blanc {
        // Le site devient rouge
        c.Snapshot.Couleur = Rouge
        
        // Mettre à jour l'état local et l'envoyer sur l'anneau
        c.Snapshot.UpdateEtatLocal(c)
        etatMsg := EtatMessage{
            EtatLocal: c.Snapshot.EtatLocal,
            Bilan:     c.Snapshot.Bilan,
        }
        c.EnvoyerSurAnneau(EtatMsg, etatMsg)
    } else if couleurRecue == Blanc && c.Snapshot.Couleur == Rouge {
        // C'est un message prépost
        c.EnvoyerSurAnneau(PrePost, message)
    }
    
    // Transmettre le message à l'application
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
	sender := utils.Findval(msg, "msg", c.Nom)


	if c.Snapshot.Couleur == Blanc {
		// Premier marqueur reçu : devenir rouge et sauvegarder l'état local
		c.Snapshot.Couleur = Rouge

		etat := map[string]interface{}{
			"Horloge":      c.Horloge,
			"InSection":    c.IsInSection,
			"DoctorsCount": CopyDoctorsCount(), // à adapter
		}
		c.Snapshot.EtatGlobal[c.Nom] = etat
		c.Snapshot.NbEtatAttendu = N - 1

		utils.Display_e("SNAPSHOT", fmt.Sprintf("%s devient rouge suite à snapshot reçu de %s", c.Nom, sender), c.Nom)

		// Propagation du marqueur aux autres
		for i := 1; i <= N; i++ {
			target := fmt.Sprintf("ctrl_%d", i)
			if target == c.Nom || target == sender {
				continue
			}

			msgSnapshot := utils.Msg_format("type", "snapshot") +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("VC", Encodehorloge(c.VectorClock)) +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge))

			fmt.Println(msgSnapshot)
		}
	} else {
		// Si déjà rouge, on ignore ce marqueur (état déjà pris)
		utils.Display_e("SNAPSHOT", fmt.Sprintf("Marqueur reçu de %s, mais état déjà pris.", sender), c.Nom)
	}
}
func CopyDoctorsCount() map[string]int {

	return make(map[string]int)
}

func Encodehorloge(vc map[string]int) string {
	var result string
	for k, v := range vc {
		result += fmt.Sprintf("%s=%d|", k, v)
	}
	return strings.TrimSuffix(result, "|")
}
// Vérification de la fin du snapshot
func (c *Controller) VerifierFinSnapshot() {
	if c.Snapshot.NbEtatAttendu <= 0 && c.Snapshot.NbMessagePrepostAttendu <= 0 {
		utils.Display_e("SNAPSHOT", "Snapshot terminé!", c.Nom)
		utils.Display_e("SNAPSHOT", fmt.Sprintf("État global final: %#v", c.Snapshot.EtatGlobal), c.Nom)
		
		// Reset snapshot state for next use
		c.Snapshot = *NewSnapshot()
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
			// Formater le message d'état (à adapter selon votre format)
			etatStr := fmt.Sprintf("%v", etatMsg.EtatLocal) // Simple string representation
			msg = utils.Msg_format("type", string(EtatMsg)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("etat", etatStr) +
				utils.Msg_format("bilan", strconv.Itoa(etatMsg.Bilan)) +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge))
		}
	case PrePost:
		if prepostMsg, ok := content.(string); ok {
			msg = utils.Msg_format("type", string(PrePost)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", prepostMsg) +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge))
		}
	case AppMsg:
		if appMsg, ok := content.(CtrlMessage); ok {
			msg = utils.Msg_format("type", string(AppMsg)) +
				utils.Msg_format("sender", c.Nom) +
				utils.Msg_format("msg", appMsg.Contenu) +
				utils.Msg_format("couleur",string(appMsg.Couleur)) +
				utils.Msg_format("hlg", strconv.Itoa(c.Horloge))
		}
	}
	
	if msg != "" {
		// Envoyer le message (pour l'anneau, on l'imprime simplement)
		fmt.Println(msg)
	}
}
func main() {
	flag.Parse()

	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())
	ctrl := NewController(nom)
	ctrl.HandleMessage()
}
