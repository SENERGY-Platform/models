/*
 * Copyright 2022 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package models

// Hub doesn't use embedded HubEdit because it's used by device-repository mongodb/bson operations
// and the `bson:",inline"`tag would violate "separation of concerns"
type Hub struct {
	Id             string   `json:"id"`
	Name           string   `json:"name"`
	Hash           string   `json:"hash"`
	DeviceLocalIds []string `json:"device_local_ids"`
	DeviceIds      []string `json:"device_ids"` //not user defined; set by finding device-ids of this.DeviceLocalIds
	OwnerId        string   `json:"owner_id"`
}

// HubEdit is Hub without Hub.DeviceIds
// used in device-manager operations with no valid use of Hub.DeviceIds
type HubEdit struct {
	Id             string   `json:"id"`
	Name           string   `json:"name"`
	Hash           string   `json:"hash"`
	DeviceLocalIds []string `json:"device_local_ids"`
	OwnerId        string   `json:"owner_id"`
}

func (this *Hub) ToHubEdit() HubEdit {
	return HubEdit{
		Id:             this.Id,
		Name:           this.Name,
		Hash:           this.Hash,
		DeviceLocalIds: this.DeviceLocalIds,
		OwnerId:        this.OwnerId,
	}
}

func (this *HubEdit) ToHub() Hub {
	return Hub{
		Id:             this.Id,
		Name:           this.Name,
		Hash:           this.Hash,
		DeviceLocalIds: this.DeviceLocalIds,
		OwnerId:        this.OwnerId,
	}
}
