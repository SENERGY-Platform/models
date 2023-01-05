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

type DeviceGroup struct {
	Id            string                      `json:"id"`
	Name          string                      `json:"name"`
	Image         string                      `json:"image"`
	Criteria      []DeviceGroupFilterCriteria `json:"criteria"`
	DeviceIds     []string                    `json:"device_ids"`
	CriteriaShort []string                    `json:"criteria_short,omitempty"`
	Attributes    []Attribute                 `json:"attributes"`
}

func (this *DeviceGroup) SetShortCriteria() {
	this.CriteriaShort = []string{}
	for _, criteria := range this.Criteria {
		this.CriteriaShort = append(this.CriteriaShort, criteria.Short())
	}
}

type DeviceGroupFilterCriteria struct {
	Interaction   Interaction `json:"interaction"`
	FunctionId    string      `json:"function_id"`
	AspectId      string      `json:"aspect_id"`
	DeviceClassId string      `json:"device_class_id"`
}

func (this DeviceGroupFilterCriteria) Short() string {
	return this.FunctionId + "_" + this.AspectId + "_" + this.DeviceClassId + "_" + string(this.Interaction)
}
