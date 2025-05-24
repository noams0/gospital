package main

import (
	"fmt"
	"gospital/utils"
	"time"
)

func (a *App) waitingFoSending(destinator string) {
	/*
		waitingFoSending() envoie une demande de section critique (SC),
		attend l’accès à la SC, décrémente le compteur du médecin courant, met à jour les données,
		envoie les messages correspondants, puis libère la SC et notifie l’envoi.
	*/
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

	msg := utils.Msg_format("type", "finSC") + utils.Msg_format("new_data", new_data)

	fmt.Print(msg + "\n")
	a.doctorInfo.ActivityLog = append([]string{"FinSC"}, a.doctorInfo.ActivityLog...)
	//LIBERATION SC PUIS SEND => SINON BUG
	//test pour voir prepost cas n°2
	//time.Sleep(1 * time.Second)
	msg = utils.Msg_format("type", "send") + utils.Msg_format("destinator", destinator)
	utils.Display_w("action :", msg, a.name)
	fmt.Print(msg + "\n")
	a.doctorInfo.ActivityLog = append([]string{"Envoie"}, a.doctorInfo.ActivityLog...)

	a.inSC = false
}
