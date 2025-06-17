package main

import (
	"fmt"
	"gospital/utils"
	"strconv"
)

func (a *App) leave() {

	utils.Display_w("leave", a.name, a.name)

	msg := utils.Msg_format("type", "askToLeave") +
		utils.Msg_format("sender", a.name)
	fmt.Print(msg + "\n")
}

func (a *App) quit(pid int) {

	utils.Display_w("quit", a.name, a.name)

	msg := utils.Msg_format("type", "askToQuit") +
		utils.Msg_format("sender", a.name) +
		utils.Msg_format("pid_app", strconv.Itoa(pid))
	fmt.Print(msg + "\n")
}
