package main

import (
	"fmt"
	"gospital/utils"
	"time"
)

func (a *utils.App) waitingFoSending(destinator string) {
	fmt.Print(utils.Msg_format("type", "demandeSC") + "\n")
	a.WaitingSC = true
	a.DoctorInfo.ActivityLog = append([]string{"DemSC"}, a.DoctorInfo.ActivityLog...)
	for !a.InSC {
		time.Sleep(100 * time.Millisecond)
	}
	a.DoctorInfo.DoctorsCount[*p_nom]--
	if a.DoctorInfo.DoctorsCount[*p_nom] < 0 {
		a.DoctorInfo.DoctorsCount[*p_nom] = 0
	}
	new_data := ""
	for site, count := range a.DoctorInfo.DoctorsCount {
		new_data += fmt.Sprintf("|%s=%d", site, count)
	}
	//msg = "send" + destinator

	a.DoctorInfo.ActivityLog = append([]string{"Envoie"}, a.DoctorInfo.ActivityLog...)

	msg := utils.Msg_format("type", "finSC") + utils.Msg_format("new_data", new_data)

	fmt.Print(msg + "\n")
	a.DoctorInfo.ActivityLog = append([]string{"FinSC"}, a.DoctorInfo.ActivityLog...)
	//LIBERATION SC PUIS SEND => SINON BUG

	msg = utils.Msg_format("type", "send") + utils.Msg_format("destinator", destinator)
	utils.Display_w("action :", msg, a.Name)
	fmt.Print(msg + "\n")
	a.InSC = false
}

func (a *utils.App) snapshot() {
    a.DoctorInfo.ActivityLog = append([]string{"Snapshot"}, a.DoctorInfo.ActivityLog...)
    
    // Formatage du message pour le contrôleur
    msg := utils.Msg_format("type", "snapshot")
    
    // Envoi du message au contrôleur
    utils.Display_w("snapshot", "Demande de snapshot envoyée", a.Name)
    fmt.Print(msg + "\n")
}