package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
)

func defineWSurl(name string) string {
	re := regexp.MustCompile(`app_(\d+)`)
	matches := re.FindStringSubmatch(name)
	if len(matches) != 2 {
		log.Fatalf("Nom d'application invalide : %s", name)
	}
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		log.Fatalf("Numéro d'application invalide : %s", matches[1])
	}
	// Génération dynamique du port (8080 + num - 1)
	return fmt.Sprintf(":%d", 8080+num-1)
}
