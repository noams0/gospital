package main

import (
	"TD06-SR05/utils"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
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

func sendperiodic() {
	var sndmsg string
	var i int

	i = 0

	for {
		mutex.Lock()
		i = i + 1
		sndmsg = "message_" + strconv.Itoa(i) + "\n"
		fmt.Print(sndmsg)
		mutex.Unlock()
		time.Sleep(time.Duration(2) * time.Second)
	}
}

func receive() {
	var rcvmsg string

	for {
		fmt.Scanln(&rcvmsg)
		mutex.Lock()
		display_w("receive", "reception <"+rcvmsg+">")
		//for i := 1; i < 6; i++ {
		//	display_w("receive", "traitement message"+strconv.Itoa(i))
		//
		//	time.Sleep(time.Duration(1) * time.Second)
		//}
		mutex.Unlock()
		rcvmsg = ""
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
