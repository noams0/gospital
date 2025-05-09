package utils

type DoctorPayload struct {
	Sender       string         `json:"sender"`
	DoctorsCount map[string]int `json:"doctors"`
}
