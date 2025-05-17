package utils

/*les champs des types doivent avoir des majuscules
pour être exportés*/

type DoctorPayload struct {
	Sender       string         `json:"sender"`
	DoctorsCount map[string]int `json:"doctors"`
	ActivityLog  []string       `json:"activity_log"`
}

//utilisé dans app
type DoctorInfo struct {
	DoctorsCount map[string]int
	ActivityLog  []string
}

//utilisé dans app
type App struct {
	Name       string
	DoctorInfo DoctorInfo
	Actions    chan map[string]interface{}
	WaitingSC  bool
	InSC       bool
}