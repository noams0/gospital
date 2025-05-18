/*fonctions qui reçoivent/lisent (lecture stdin)
et traitent les messages reçus, en provenance de ctrl*/

package main

import (
    "bufio"
    "fmt"
    "gospital/utils"
    "os"
    "strconv"
    "strings"
)


func (a *App) receive() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		globalMutex.Lock()
		utils.Display_w("receive", "reception <"+msg+">", a.name)
		if msg == "receive" {
			go a.waitingFoReceivng()
		}
		if strings.HasPrefix(msg, "TAB_REQ") {
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