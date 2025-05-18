/*fonctions qui permettent la gestion de la sauvegarde
(initialisation, début, mise à jour, vérification).*/

package main

import (
    "fmt"
    "strconv"
    "gospital/utils"
)


func NewSnapshot() *Snapshot {
	/*initialisation de la sauvegarde*/
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
	/*Début sauvegarde. 
	- Marque le site comme initiateur, 
	- sauvegarde l’état local 
	- et envoie les messages de marqueur.*/
	
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

//maj état local
func (s *Snapshot) UpdateEtatLocal(c *Controller) {
	s.EtatLocal = map[string]interface{}{
		"Horloge":      c.Horloge,
		"InSection":    c.IsInSection,
		"DoctorsCount": CopyDoctorsCount(),
		"Tab":          c.Tab,
	}
}

func CopyDoctorsCount() map[string]int {
	//à quoi sert cette fonction ?

	return make(map[string]int)
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