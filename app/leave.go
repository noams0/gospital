package main

import (
	"fmt"
	"gospital/utils"
)

func (a *App) leave() {
	utils.Display_e("app snapshot", fmt.Sprintf("!!!app snapshot"), a.name)

	utils.Display_w("leave", a.name, a.name)

	msg := utils.Msg_format("type", "askToLeave") +
		utils.Msg_format("sender", a.name)
	fmt.Print(msg + "\n")
}
