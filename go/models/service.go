package models

type Service struct {
	Id              string      `json:"id"`
	LocalId         string      `json:"local_id"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Interaction     Interaction `json:"interaction"`
	ProtocolId      string      `json:"protocol_id"`
	Inputs          []Content   `json:"inputs"`
	Outputs         []Content   `json:"outputs"`
	Attributes      []Attribute `json:"attributes"`
	ServiceGroupKey string      `json:"service_group_key"`
}

type Interaction string

const (
	EVENT             Interaction = "event"
	REQUEST           Interaction = "request"
	EVENT_AND_REQUEST Interaction = "event+request"
)

type ServiceGroup struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
