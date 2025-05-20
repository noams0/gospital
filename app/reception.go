package main 

import(
	"bufio"
	"fmt"
	"gospital/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

func (a *App) waitingFoReceivng() {
	a.doctorInfo.ActivityLog = append([]string{"Receive"}, a.doctorInfo.ActivityLog...)

	fmt.Print(utils.Msg_format("type", "demandeSC") + "\n")

	a.waitingSC = true
	a.doctorInfo.ActivityLog = append([]string{"DemSC"}, a.doctorInfo.ActivityLog...)
	for !a.inSC {
		time.Sleep(100 * time.Millisecond)
	}
	a.doctorInfo.DoctorsCount[*p_nom]++
	new_data := ""
	for site, count := range a.doctorInfo.DoctorsCount {
		new_data += fmt.Sprintf("|%s=%d", site, count)
	}
	msg := utils.Msg_format("type", "finSC") + utils.Msg_format("new_data", new_data)

	fmt.Print(msg + "\n")
	a.doctorInfo.ActivityLog = append([]string{"FinSC"}, a.doctorInfo.ActivityLog...)

}


func (a *App) receive() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		globalMutex.Lock()
		utils.Display_w("receive", "reception <"+msg+">", a.name)
		if msg == "receive" {
			go a.waitingFoReceivng()
		} else if msg == "askForState" {
			yrstate := a.doctorInfo.DoctorsCount[*p_nom]
			str_yrstate := strconv.Itoa(yrstate)
			stderr.Println(str_yrstate)
			msg = utils.Msg_format("type", "yourState") + utils.Msg_format("etat_local", str_yrstate)
			fmt.Println(msg)

		} else if strings.HasPrefix(msg, "endSnapshot") {
			snapshotData := strings.TrimPrefix(msg, "endSnapshot")
			a.doctorInfo.Snapshot = snapshotData

		} else if strings.HasPrefix(msg, "TAB_REQ") {
			a.doctorInfo.ActivityLog = append([]string{msg}, a.doctorInfo.ActivityLog...)

		} else if msg == "debutSC" && a.waitingSC {
			a.inSC = true
			a.waitingSC = false
			a.doctorInfo.ActivityLog = append([]string{"DebSC"}, a.doctorInfo.ActivityLog...)
		} else if utils.Findval(msg, "type", a.name) == "new_data" {
			utils.Display_w("NEW_DATA", "IL FAUT MAJ", a.name)
			data := utils.Findval(msg, "new_data", a.name)
			pairs := strings.Split(data, "|")
			for _, pair := range pairs {
				parts := strings.Split(pair, "=")
				if len(parts) == 2 {
					appName := parts[0]
					val, err := strconv.Atoi(parts[1])
					if err == nil {
						a.doctorInfo.DoctorsCount[appName] = val
						utils.Display_w("NEW_DATA", fmt.Sprintf("Mise à jour : %s -> %d", appName, val), a.name)
						//a.doctorInfo.ActivityLog = append([]string{"NewData"}, a.doctorInfo.ActivityLog...)

					} else {
						utils.Display_e("NEW_DATA", "Erreur de conversion pour "+pair, a.name)
					}
				}
			}
		}
		globalMutex.Unlock()
	}
	if err := scanner.Err(); err != nil {
		utils.Display_e("receive", "erreur de lecture: "+err.Error(), a.name)
	}
}