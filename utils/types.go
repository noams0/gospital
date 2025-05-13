package utils

type DoctorPayload struct {
	Sender       string         `json:"sender"`
	DoctorsCount map[string]int `json:"doctors"`
	ActivityLog  []string       `json:"activity_log"`
}
