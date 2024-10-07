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

type Device struct {
	Id           string      `json:"id"`
	LocalId      string      `json:"local_id"`
	Name         string      `json:"name"`
	Attributes   []Attribute `json:"attributes"`
	DeviceTypeId string      `json:"device_type_id"`
	OwnerId      string      `json:"owner_id"`
}

type ExtendedDevice struct {
	Device
	ConnectionState ConnectionState `json:"connection_state"`
	DisplayName     string          `json:"display_name"  bson:"display_name"`
	DeviceTypeName  string          `json:"device_type_name" bson:"-"`      //computed on request, not stored
	DeviceType      *DeviceType     `json:"device_type,omitempty" bson:"-"` //optional
	Shared          bool            `json:"shared" bson:"-"`                //computed on request, not stored
	Permissions     Permissions     `json:"permissions" bson:"-"`           //computed on request, not stored
}

type ConnectionState = string

const ConnectionStateOnline = "online"
const ConnectionStateOffline = "offline"
const ConnectionStateUnknown = ""
