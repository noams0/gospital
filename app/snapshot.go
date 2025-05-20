package main

import (
	"fmt"
	"gospital/utils"
	"strings"
)

func (a *App) snapshot() {
	utils.Display_e("app snapshot", fmt.Sprintf("!!!app snapshot"), a.name)

	a.doctorInfo.ActivityLog = append([]string{"Snapshot"}, a.doctorInfo.ActivityLog...)

	payload := a.doctorInfo.SendDoctorInfo()

	var new_data strings.Builder
	for site, count := range payload.DoctorsCount {
		new_data.WriteString(fmt.Sprintf("|%s=%d", site, count))
	}

	utils.Display_w("snapshot", fmt.Sprintf("Envoi snapshot: %v", payload.DoctorsCount), a.name)

	msg := utils.Msg_format("type", "snapshot") +
		utils.Msg_format("sender", payload.Sender) +
		utils.Msg_format("data", new_data.String())
	utils.Msg_format("data", new_data.String())

	fmt.Print(msg + "\n")
}

