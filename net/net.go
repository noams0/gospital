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

func main() {
	flag.Parse()

	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())
	net := NewNet(*p_nom, nom, *totalSites)

	utils.Display_n("NET", "démarre", net.Nom)

	net.run()

}

func (net *Net) run() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		utils.Display_n("NET, rec", line, "net")

		// Re-transmet sur stdout (anneau logique)
		//fmt.Println(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[NET] Erreur de lecture : %v\n", err)
	}

}
