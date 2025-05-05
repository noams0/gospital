package utils

import (
	"log"
	"strconv"
	"strings"
)

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

func CloneVC(vc map[string]int) map[string]int {
	clone := make(map[string]int)
	for k, v := range vc {
		clone[k] = v
	}
	return clone
}

func EncodeVC(vc map[string]int) string {
	var parts []string
	for k, v := range vc {
		parts = append(parts, k+"-"+strconv.Itoa(v))
	}
	return strings.Join(parts, ",")
}

func DecodeVC(s string) map[string]int {
	vc := make(map[string]int)
	if s == "" {
		return vc
	}
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "-")
		if len(kv) == 2 {
			val, err := strconv.Atoi(kv[1])
			if err == nil {
				vc[kv[0]] = val
			}
		}
	}
	return vc
}
