package models

type Device struct {
	Id           string      `json:"id"`
	LocalId      string      `json:"local_id"`
	Name         string      `json:"name"`
	Attributes   []Attribute `json:"attributes"`
	DeviceTypeId string      `json:"device_type_id"`
}
