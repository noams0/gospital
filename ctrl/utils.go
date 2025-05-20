package main

import (
    "fmt"
    "strconv"
    "strings"
    "gospital/utils"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func TabToString(tab map[string]EtatReqSite) string {
	var result string = "TAB_REQ"

	for k, v := range tab {
		result += fmt.Sprintf("%s : Horloge=%d (%s), Type=%s,", k, v.Horloge, strings.ReplaceAll(utils.EncodeVC(v.VectorClock), ",", " "), v.TypeRequete)
		utils.Display_f("TEST", utils.EncodeVC(v.VectorClock), "XXX")
	}
	return result
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