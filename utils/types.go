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


//types utilisés dans ctrl

type MessageType string 

const (
	Requete    MessageType = "request" 
	Liberation MessageType = "liberation"
	Accuse     MessageType = "ack"
)
const (
	EtatMsg MessageType = "etat" 
	PrePost MessageType = "prepost" 
	AppMsg  MessageType = "app_msg" 
)

type EtatReqSite struct { 
	Horloge     int
	TypeRequete MessageType
}

type Couleur string

const (
    Blanc Couleur = "blanc" 
    Rouge Couleur = "rouge" 
)


type Snapshot struct {
	Couleur                  Couleur                      // "blanc" ou "rouge"
	EtatGlobal               map[string]interface{}      
	EtatLocal                map[string]interface{}
	Initiateur               bool
	NbEtatAttendu            int
	NbMessagePrepostAttendu  int
	Bilan                    int
}

type Controller struct {
	Nom         string
	Horloge     int
	VectorClock map[string]int
	Tab         map[string]EtatReqSite
	IsInSection bool
	Snapshot Snapshot
	SnapshotEnCours bool

}
type EtatMessage struct {
	EtatLocal map[string]interface{}
	Bilan     int
}
type CtrlMessage struct {
	Type    MessageType
	Contenu string
	Couleur Couleur
}