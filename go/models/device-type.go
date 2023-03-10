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

type DeviceType struct {
	Id            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	ServiceGroups []ServiceGroup `json:"service_groups"`
	Services      []Service      `json:"services"`
	DeviceClassId string         `json:"device_class_id"`
	Attributes    []Attribute    `json:"attributes"`
}
