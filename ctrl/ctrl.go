package main

import (
	"flag"
	"log"
	"os"
	"strconv"
)

var N = 3

var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)

var tab = make(map[string]EtatReqSite)

var p_nom *string = flag.String("n", "ecrivain", "nom")


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

func main() {
	flag.Parse()

	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())
	ctrl := NewController(nom)
	ctrl.HandleMessage()
}
