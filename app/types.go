package main

type DoctorInfo struct {
	DoctorsCount map[string]int
	ActivityLog  []string
	Snapshot     string
}

// 
type App struct {
	name       string
	doctorInfo DoctorInfo
	actions    chan map[string]interface{}
	waitingSC  bool
	inSC       bool
}