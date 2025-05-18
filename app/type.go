/*définition des types/structures de données*/

package main

type DoctorInfo struct {
	DoctorsCount map[string]int
	ActivityLog  []string
}

type App struct {
	name       string
	doctorInfo DoctorInfo
	actions    chan map[string]interface{}
	waitingSC  bool
	inSC       bool
}