package models

type Location struct {
	Id             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Image          string   `json:"image"`
	DeviceIds      []string `json:"device_ids"`
	DeviceGroupIds []string `json:"device_group_ids"`
}
