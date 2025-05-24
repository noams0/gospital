package utils

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

/*
FONCTIONS D'AFFICHAGE
*/
var stderr = log.New(os.Stderr, "", 0)

func Display_d(where string, what string, nom string) {
	stderr.Printf("%s + [%s] %-8.8s : %s\n%s", ColorBlue, nom, where, what, ColorReset)
}
func Display_w(where string, what string, nom string) {

	stderr.Printf("%s * [%s] %-8.8s : %s\n%s", ColorYellow, nom, where, what, ColorReset)
}
func Display_e(where string, what string, nom string) {
	stderr.Printf("%s ! [%s] %-8.8s : %s\n%s", ColorRed, nom, where, what, ColorReset)
}
func Display_f(where string, what string, nom string) {
	stderr.Printf("%s * [%s] %-8.8s : %s\n%s", ColorPurple, nom, where, what, ColorReset)
}

/*
FONCTIONS ENCODAGE ET DECODAGE DE MESSAGE
*/
const fieldsep = "/"
const keyvalsep = "="

func Msg_format(key string, val string) string {
	return fieldsep + keyvalsep + key + keyvalsep + val
}

func Findval(msg string, key string, nom string) string {
	if len(msg) < 4 {
		//Display_w("findval", "message trop court : "+msg, nom)
		return ""
	}
	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)

	for _, keyval := range tab_allkeyvals {
		if len(keyval) < 3 { // au moins 1 pour separateur, 1 pour key, 1 pour val
			//Display_w("findval", "message trop court : "+msg, nom)
			continue
		}
		equ := keyval[0:1]
		tabkeyval := strings.SplitN(keyval[1:], equ, 2)
		if len(tabkeyval) != 2 {
			continue
		}
		if tabkeyval[0] == key {
			return tabkeyval[1]
		}
	}
	return ""
}

/*
FONCTIONS D'AIDE POUR LES HORLOGES
*/
func Recaler(x, y int) int {
	if x < y {
		return y + 1
	}
	return x + 1
}

func CloneVC(vc map[string]int) map[string]int {
	clone := make(map[string]int)
	for k, v := range vc {
		clone[k] = v
	}
	return clone
}

func EncodeVC(vc map[string]int) string {
	var parts []string
	keys := make([]string, 0, len(vc))
	for k := range vc {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		parts = append(parts, k+"-"+strconv.Itoa(vc[k]))
	}
	return strings.Join(parts, ",")
}

func DecodeVC(s string) map[string]int {
	vc := make(map[string]int)
	if s == "" {
		return vc
	}
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "-")
		if len(kv) == 2 {
			val, err := strconv.Atoi(kv[1])
			if err == nil {
				vc[kv[0]] = val
			}
		}
	}
	return vc
}

func InitVC(n int) map[string]int {
	vc := make(map[string]int)
	for i := 0; i < n; i++ {
		vc["ctrl_"+strconv.Itoa(i+1)] = 0
	}
	return vc
}

func IncVC(vc map[string]int, name string) map[string]int {
	vc[name]++
	return vc
}

func MaxVC(vc1, vc2 map[string]int, name string) map[string]int {
	vc := make(map[string]int)
	for k, v := range vc1 {
		if v > vc2[k] {
			vc[k] = v
		} else {
			vc[k] = vc2[k]
		}
	}
	vc[name]++
	return vc
}

/*
 FONCTIONS ANNEXES
*/

func Int_to_ctrl(i int) string {
	return fmt.Sprintf("ctrl_%i", i)
}

func Int_to_app(i int) string {
	return fmt.Sprintf("app_%i", i)
}

func App_to_ctrl(app string) string {
	re := regexp.MustCompile(`app_(\d+)`)
	matches := re.FindStringSubmatch(app)
	if len(matches) != 2 {
		return "ctrl"
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return "ctrl"
	}

	return fmt.Sprintf("ctrl_%d", num)
}

func less(vc1 map[string]int, name1 string, vc2 map[string]int, name2 string) bool {
	// Compare (vc1, name1) < (vc2, name2)
	v1 := vc1[name1]
	v2 := vc2[name2]
	if v1 != v2 {
		return v1 < v2
	}
	return name1 < name2
}

// Sauvegarde du snapshot
func SaveSnapshot(name, snapshot string, horloge string) {
	filename := "snapshot_" + name + ".save"
	content := "horloge: " + horloge + "\n" + snapshot
	f, err := os.Create(filename)
	if err != nil {
		stderr.Println(err)
		return
	}
	l, err := f.WriteString(content)
	if err != nil {
		stderr.Println(err)
		f.Close()
		return
	}
	err = f.Close()
	if err != nil {
		stderr.Println(err)
		return
	}
	stderr.Printf("Saved %s : %d o\n", filename, l)
}
