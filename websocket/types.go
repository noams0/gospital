package websocket

type Site struct {
	ID       int `json:"id"`
	Medecins int `json:"medecins"` // Nombre de médecins disponibles
}

type Register struct {
	Sites map[int]*Site `json:"sites"` // Clé: ID du site, valeur: informations du site
}

type SimulationInterface interface {
	GetRegister() map[int]*Site                       // Retourne l'état actuel du registre des sites
	ProcessAction(action map[string]interface{}) bool // Traitement des actions des clients
	GetEventStream() chan map[string]interface{}
	Run()
}
