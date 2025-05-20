package main

import (
	"flag"
	"gospital/utils"
	"log"
	"os"
	"strconv"
)

var N = 3
var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)
var p_nom *string = flag.String("n", "ecrivain", "nom")

func NewController(nomcourt, nom string) *Controller {
	/*Crée et initialise une nouvelle instance de Controller*/
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



func (c *Controller) Msg_Horloge() string {
	msg := utils.Msg_format("hlg", strconv.Itoa(c.Horloge))
	c.VectorClock = utils.IncVC(c.VectorClock, c.NomCourt)
	msg += utils.Msg_format("vc", utils.EncodeVC(c.VectorClock))
	return msg
}


func main() {

	/*Point d’entrée du programme : parse les arguments, 
	crée un Controller et lance la boucle de gestion des messages.*/

	flag.Parse()
	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())
	ctrl := NewController(*p_nom, nom)
	ctrl.HandleMessage()
}
