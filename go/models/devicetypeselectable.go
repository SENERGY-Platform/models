/*
 * Copyright 2026 InfAI (CC SES)
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

// DeviceTypeSelectable is the shape of the device-repository device-type-selectables answer:
// which services of a device type offer what a filter criteria asked for, and where in their
// content the values sit.
//
// The neighbouring shape is Selectable, which answers the same question per device rather
// than per device type and spells its json camelCase. The two are not interchangeable.
type DeviceTypeSelectable struct {
	DeviceTypeId       string                         `json:"device_type_id,omitempty"`
	Services           []Service                      `json:"services,omitempty"`
	ServicePathOptions map[string][]ServicePathOption `json:"service_path_options,omitempty"`
}

// ServicePathOption is one way a service offers what a filter criteria asked for. It carries
// the service id, because the options of a whole device type are answered at once.
type ServicePathOption struct {
	ServiceId             string         `json:"service_id"`
	Path                  string         `json:"path"`
	CharacteristicId      string         `json:"characteristic_id"`
	AspectNode            AspectNode     `json:"aspect_node"` //deprecated: alias for a single element AspectNodes; holds the node with the alphabetically first id
	AspectNodes           []AspectNode   `json:"aspect_nodes,omitempty"`
	FunctionId            string         `json:"function_id"`
	IsVoid                bool           `json:"is_void"`
	Value                 interface{}    `json:"value,omitempty"`
	IsControllingFunction bool           `json:"is_controlling_function"`
	Configurables         []Configurable `json:"configurables,omitempty"`
	Type                  Type           `json:"type,omitempty"`
	Interaction           Interaction    `json:"interaction"`
}
