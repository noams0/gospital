package main

import (
	"encoding/json"
	"fmt"
	"gospital/utils"
	"strconv"
)

// initialisation de la sauvegarde
func NewSnapshot(nbSite int) *Snapshot {
	return &Snapshot{
		Couleur:                 Blanc,
		EtatGlobal:              make(map[string]interface{}),
		EtatLocal:               make(map[string]interface{}),
		Initiateur:              false,
		NbEtatAttendu:           5,
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
	stderr.Println("NB SITE", c.NbSite)
	c.Snapshot.NbEtatAttendu = c.NbSite // Pas N - 1 car on attend la valeur locale
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

func (c *Controller) VerifierFinSnapshot() {
	/*Vérifie si le snapshot est terminé et effectue la sauvegarde.*/

	utils.Display_e("NB???", fmt.Sprintf("%d et %d", c.Snapshot.NbEtatAttendu, c.Snapshot.NbMessagePrepostAttendu), "ccc")
	if c.Snapshot.NbEtatAttendu <= 1 && c.Snapshot.NbMessagePrepostAttendu <= 0 {
		utils.Display_e("SNAPSHOT", "Snapshot terminé!", c.Nom)
		utils.Display_e("SNAPSHOT", fmt.Sprintf("État global final: %#v", c.Snapshot.EtatGlobal), c.Nom)
		etatBytes, _ := json.Marshal(c.Snapshot.EtatGlobal)
		etatStr := string(etatBytes)

		fmt.Println("endSnapshot", etatStr)
		utils.SaveSnapshot(c.NomCourt, etatStr, utils.EncodeVC(c.VectorClock))

		c.Snapshot = *NewSnapshot(c.NbSite)
		c.SnapshotEnCours = false
	}
}
