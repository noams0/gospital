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

var tab = make(map[string]utils.EtatReqSite)


func NewController(nom string) *utils.Controller {
	return &utils.Controller{
		Nom:         nom,
		Horloge:     0,
		VectorClock: make(map[string]int),
		Tab:         make(map[string]utils.EtatReqSite),
		IsInSection: false,
		Snapshot: *NewSnapshot(),
	}
}

var p_nom *string = flag.String("n", "ecrivain", "nom")



func TabToString(tab map[string]utils.EtatReqSite) string {
	var result string = "TAB_REQ"

	for k, v := range tab {
		result += fmt.Sprintf("%s : Horloge=%d, Type=%s,", k, v.Horloge, v.TypeRequete)
	}
	return result
}



func isFirstRequest(tab map[string]utils.EtatReqSite, me string, h int) bool {
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
func NewSnapshot() *utils.Snapshot {
    return &utils.Snapshot{
        Couleur:                 utils.Blanc,
        EtatGlobal:              make(map[string]interface{}),
		EtatLocal:               make(map[string]interface{}),
        Initiateur:              false,
        NbEtatAttendu:           N,
        Bilan:                   0,
        NbMessagePrepostAttendu: 0,
    }
}

//maj état local
func (s *utils.Snapshot) UpdateEtatLocal(c *utils.Controller) {
	s.EtatLocal = map[string]interface{}{
		"Horloge":      c.Horloge,
		"InSection":    c.IsInSection,
		"DoctorsCount": CopyDoctorsCount(),
		"Tab":          c.Tab,
	}
}
// début 
/*func (c *utils.Controller) DebutSnapshot() {
    // Le site devient rouge
    c.Snapshot.Couleur = utils.Rouge
    
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
func (c *utils.Controller) DebutSnapshot() {
	// Devenir rouge
	c.Snapshot.Couleur = utils.Rouge
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
func (c *utils.Controller) VerifierFinSnapshot() {
	if c.Snapshot.NbEtatAttendu <= 0 && c.Snapshot.NbMessagePrepostAttendu <= 0 {
		utils.Display_e("SNAPSHOT", "Snapshot terminé!", c.Nom)
		utils.Display_e("SNAPSHOT", fmt.Sprintf("État global final: %#v", c.Snapshot.EtatGlobal), c.Nom)
		
		// Reset snapshot state for next use
		c.Snapshot = *NewSnapshot()
	}
}

func main() {
	flag.Parse()

	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())
	ctrl := NewController(nom)
	ctrl.HandleMessage()
}
