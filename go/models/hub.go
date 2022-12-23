package models

type Hub struct {
	Id             string   `json:"id"`
	Name           string   `json:"name"`
	Hash           string   `json:"hash"`
	DeviceLocalIds []string `json:"device_local_ids"`
	DeviceIds      []string `json:"device_ids"` //not user defined; set by device-manager by finding device-ids of this.DeviceLocalIds
}
