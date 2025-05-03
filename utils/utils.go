package utils

import "log"

func Recaler(x, y int) int {
	if x < y {
		return y + 1
	}
	return x + 1
}

func Display_d(stderr *log.Logger, p_nom *string, pid int, where string, what string) {
	stderr.Printf("%s + [%.6s %d] %-8.8s : %s\n%s", ColorBlue, *p_nom, pid, where, what, ColorReset)
}

func Display_w(stderr *log.Logger, p_nom *string, pid int, where string, what string) {

	stderr.Printf("%s * [%.6s %d] %-8.8s : %s\n%s", ColorYellow, *p_nom, pid, where, what, ColorReset)
}

func Display_e(stderr *log.Logger, p_nom *string, pid int, where string, what string) {
	stderr.Printf("%s ! [%.6s %d] %-8.8s : %s\n%s", ColorRed, *p_nom, pid, where, what, ColorReset)
}
