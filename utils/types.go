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

type MessageType string  //ok

const (
	Requete    MessageType = "request" //ok
	Liberation MessageType = "liberation"//ok
	Accuse     MessageType = "ack"//ok
)
const (
	EtatMsg MessageType = "etat" //ok
	PrePost MessageType = "prepost" //ok
	AppMsg  MessageType = "app_msg" //ok
)

type EtatReqSite struct {  //ok
	Horloge     int
	TypeRequete MessageType
}

type Couleur string

const (
    Blanc Couleur = "blanc" //ok
    Rouge Couleur = "rouge" //ok
)


type Snapshot struct { //ok
	Couleur                  Couleur                      // "blanc" ou "rouge"
	EtatGlobal               map[string]interface{}      
	EtatLocal                map[string]interface{}
	Initiateur               bool
	NbEtatAttendu            int
	NbMessagePrepostAttendu  int
	Bilan                    int
}

type Controller struct { //ok
	Nom         string
	Horloge     int
	VectorClock map[string]int
	Tab         map[string]EtatReqSite
	IsInSection bool
	Snapshot Snapshot
	SnapshotEnCours bool

}
type EtatMessage struct { //ok
	EtatLocal map[string]interface{}
	Bilan     int
}
type CtrlMessage struct { //ok
	Type    MessageType
	Contenu string
	Couleur Couleur
}