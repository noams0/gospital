package main

import "time"

type MessageType string

const (
	Requete    MessageType = "request"
	Liberation MessageType = "liberation"
	Accuse     MessageType = "ack"
)
const (
	EtatMsg     MessageType = "etat"
	PrePost     MessageType = "prepost"
	AppMsg      MessageType = "app_msg"
	SnapshotMsg MessageType = "snapshot"
)

type EtatReqSite struct {
	Horloge     int
	VectorClock map[string]int
	TypeRequete MessageType
}

type Couleur string

const (
	Blanc Couleur = "blanc"
	Rouge Couleur = "rouge"
)

type Snapshot struct {
	Couleur                 Couleur // "blanc" ou "rouge"
	EtatGlobal              map[string]interface{}
	EtatLocal               map[string]interface{}
	Initiateur              bool
	NbEtatAttendu           int
	NbMessagePrepostAttendu int
	Bilan                   int
	EtatEnvoye              bool
}

type Controller struct {
	Nom             string
	NomCourt        string
	Horloge         int
	VectorClock     map[string]int
	Tab             map[string]EtatReqSite
	IsInSection     bool
	Snapshot        Snapshot
	SnapshotEnCours bool
	Speed           time.Duration
	NbSite          int
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
