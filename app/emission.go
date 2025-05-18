/*fonctions qui envoient/émettent des messages 
ou modifient l’état local avant émission. */

package main

import (
    "fmt"
    "gospital/utils"
    "strings"
    "time"
)

func (a *App) snapshot() {
    a.doctorInfo.ActivityLog = append([]string{"Snapshot"}, a.doctorInfo.ActivityLog...)
    
    // Formatage du message pour le contrôleur
    msg := utils.Msg_format("type", "snapshot")
    
    // Envoi du message au contrôleur
    utils.Display_w("snapshot", "Demande de snapshot envoyée", a.name)
    fmt.Print(msg + "\n")
}

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

func (a *App) waitingFoSending(destinator string) {
	fmt.Print(utils.Msg_format("type", "demandeSC") + "\n")
	a.waitingSC = true
	a.doctorInfo.ActivityLog = append([]string{"DemSC"}, a.doctorInfo.ActivityLog...)
	for !a.inSC {
		time.Sleep(100 * time.Millisecond)
	}
	a.doctorInfo.DoctorsCount[*p_nom]--
	if a.doctorInfo.DoctorsCount[*p_nom] < 0 {
		a.doctorInfo.DoctorsCount[*p_nom] = 0
	}
	new_data := ""
	for site, count := range a.doctorInfo.DoctorsCount {
		new_data += fmt.Sprintf("|%s=%d", site, count)
	}
	//msg = "send" + destinator

	a.doctorInfo.ActivityLog = append([]string{"Envoie"}, a.doctorInfo.ActivityLog...)

	msg := utils.Msg_format("type", "finSC") + utils.Msg_format("new_data", new_data)

	fmt.Print(msg + "\n")
	a.doctorInfo.ActivityLog = append([]string{"FinSC"}, a.doctorInfo.ActivityLog...)
	//LIBERATION SC PUIS SEND => SINON BUG

	msg = utils.Msg_format("type", "send") + utils.Msg_format("destinator", destinator)
	utils.Display_w("action :", msg, a.name)
	fmt.Print(msg + "\n")
	a.inSC = false
}