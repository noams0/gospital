package main

import (
    "fmt"
    "strconv"
    "strings"
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
		result += fmt.Sprintf("%s : Horloge=%d, Type=%s,", k, v.Horloge, v.TypeRequete)
	}
	return result
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

func Encodehorloge(vc map[string]int) string {
	var result string
	for k, v := range vc {
		result += fmt.Sprintf("%s=%d|", k, v)
	}
	return strings.TrimSuffix(result, "|")
}

