package main

import (
	"bufio"
	"flag"
	"fmt"
	"gospital/utils"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)
var p_nom *string = flag.String("n", "", "nom")

func display_d(where string, what string) {
	stderr.Printf("%s + [%.6s %d] %-8.8s : %s\n%s", utils.ColorBlue, *p_nom, pid, where, what, utils.ColorReset)
}

func display_w(where string, what string) {

	stderr.Printf("%s * [%.6s %d] %-8.8s : %s\n%s", utils.ColorYellow, *p_nom, pid, where, what, utils.ColorReset)
}

func display_e(where string, what string) {
	stderr.Printf("%s ! [%.6s %d] %-8.8s : %s\n%s", utils.ColorRed, *p_nom, pid, where, what, utils.ColorReset)
}

func findval(msg string, key string) string {
	if len(msg) < 4 {
		display_w("findval", "message trop court : "+msg)
		return ""
	}
	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)

	for _, keyval := range tab_allkeyvals {
		if len(keyval) < 3 { // au moins 1 pour separateur, 1 pour key, 1 pour val
			display_w("findval", "message trop court : "+msg)
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

func sendperiodic() {
	var sndmsg string
	var i int

	i = 0

	for {
		mutex.Lock()
		i = i + 1
		sndmsg = "message_" + strconv.Itoa(i) + "from" + strconv.Itoa(pid) + "\n"
		display_d("J'ENVOIE", sndmsg)
		fmt.Print(sndmsg)
		mutex.Unlock()
		time.Sleep(time.Duration(5) * time.Second)
	}
}

func receive() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		rcvmsg := scanner.Text()
		mutex.Lock()
		display_w("receive", "reception <"+rcvmsg+">")
		mutex.Unlock()
	}
	if err := scanner.Err(); err != nil {
		display_e("receive", "erreur de lecture: "+err.Error())
	}
}

var mutex = &sync.Mutex{}

func main() {
	go sendperiodic()
	go receive()
	for {
		time.Sleep(time.Duration(60) * time.Second)
	} // Pour attendre la fin des goroutines...
}
