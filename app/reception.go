package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"gospital/utils"
)


func (a *utils.App) receive() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		globalMutex.Lock()
		utils.Display_w("receive", "reception <"+msg+">", a.Name)
		if msg == "receive" {
			go a.waitingFoReceivng()
		}
		if strings.HasPrefix(msg, "TAB_REQ") {
			a.DoctorInfo.ActivityLog = append([]string{msg}, a.DoctorInfo.ActivityLog...)

		} else if msg == "debutSC" && a.WaitingSC {
			a.InSC = true
			a.WaitingSC = false
			a.DoctorInfo.ActivityLog = append([]string{"DebSC"}, a.DoctorInfo.ActivityLog...)
		} else if utils.Findval(msg, "type", a.Name) == "new_data" {
			utils.Display_w("NEW_DATA", "IL FAUT MAJ", a.Name)
			data := utils.Findval(msg, "new_data", a.Name)
			pairs := strings.Split(data, "|")
			for _, pair := range pairs {
				parts := strings.Split(pair, "=")
				if len(parts) == 2 {
					appName := parts[0]
					val, err := strconv.Atoi(parts[1])
					if err == nil {
						a.DoctorInfo.DoctorsCount[appName] = val
						utils.Display_w("NEW_DATA", fmt.Sprintf("Mise à jour : %s -> %d", appName, val), a.Name)
						//a.DoctorInfo.ActivityLog = append([]string{"NewData"}, a.DoctorInfo.ActivityLog...)

					} else {
						utils.Display_e("NEW_DATA", "Erreur de conversion pour "+pair, a.Name)
					}
				}
			}
		}
		globalMutex.Unlock()
	}
	if err := scanner.Err(); err != nil {
		utils.Display_e("receive", "erreur de lecture: "+err.Error(), a.Name)
	}
}

func (a *utils.App) waitingFoReceivng() {
	a.DoctorInfo.ActivityLog = append([]string{"Receive"}, a.DoctorInfo.ActivityLog...)

	fmt.Print(utils.Msg_format("type", "demandeSC") + "\n")

	a.WaitingSC = true
	a.DoctorInfo.ActivityLog = append([]string{"DemSC"}, a.DoctorInfo.ActivityLog...)
	for !a.InSC {
		time.Sleep(100 * time.Millisecond)
	}
	a.DoctorInfo.DoctorsCount[*p_nom]++
	new_data := ""
	for site, count := range a.DoctorInfo.DoctorsCount {
		new_data += fmt.Sprintf("|%s=%d", site, count)
	}
	msg := utils.Msg_format("type", "finSC") + utils.Msg_format("new_data", new_data)

	fmt.Print(msg + "\n")
	a.DoctorInfo.ActivityLog = append([]string{"FinSC"}, a.DoctorInfo.ActivityLog...)

}