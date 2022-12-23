package models

type DeviceType struct {
	Id            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	ServiceGroups []ServiceGroup `json:"service_groups"`
	Services      []Service      `json:"services"`
	DeviceClassId string         `json:"device_class_id"`
	Attributes    []Attribute    `json:"attributes"`
}
