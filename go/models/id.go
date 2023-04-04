/*
 * Copyright 2023 InfAI (CC SES)
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

import "github.com/google/uuid"

func (characteristic *Characteristic) GenerateId() {
	if characteristic.Id == "" {
		characteristic.Id = URN_PREFIX + "characteristic:" + uuid.New().String()
	}
	for i, v := range characteristic.SubCharacteristics {
		v.GenerateId()
		characteristic.SubCharacteristics[i] = v
	}
}

func (class *DeviceClass) GenerateId() {
	if class.Id == "" {
		class.Id = URN_PREFIX + "device-class:" + uuid.New().String()
	}
}

func (this *Location) GenerateId() {
	if this.Id == "" {
		this.Id = URN_PREFIX + "location:" + uuid.New().String()
	}
}

func (function *Function) GenerateId() {
	if function.Id == "" {
		switch function.RdfType {
		case SES_ONTOLOGY_CONTROLLING_FUNCTION:
			function.Id = URN_PREFIX + "controlling-function:" + uuid.New().String()
		case SES_ONTOLOGY_MEASURING_FUNCTION:
			function.Id = URN_PREFIX + "measuring-function:" + uuid.New().String()
		default:
			function.Id = ""
		}
	}
}

func (aspect *Aspect) GenerateId() {
	if aspect.Id == "" {
		aspect.Id = URN_PREFIX + "aspect:" + uuid.New().String()
	}
	for i, sub := range aspect.SubAspects {
		sub.GenerateId()
		aspect.SubAspects[i] = sub
	}
}

func (concept *Concept) GenerateId() {
	if concept.Id == "" {
		concept.Id = URN_PREFIX + "concept:" + uuid.New().String()
	}
}

func (device *Device) GenerateId() {
	device.Id = URN_PREFIX + "device:" + uuid.New().String()
}

func (deviceType *DeviceType) GenerateId() {
	if deviceType.Id == "" {
		deviceType.Id = URN_PREFIX + "device-type:" + uuid.New().String()
	}
	for i, service := range deviceType.Services {
		service.GenerateId()
		deviceType.Services[i] = service
	}
}

func (deviceGroup *DeviceGroup) GenerateId() {
	if deviceGroup.Id == "" {
		deviceGroup.Id = URN_PREFIX + "device-group:" + uuid.New().String()
	}
}

func (service *Service) GenerateId() {
	if service.Id == "" {
		service.Id = URN_PREFIX + "service:" + uuid.New().String()
	}
	for i, content := range service.Inputs {
		content.GenerateId()
		service.Inputs[i] = content
	}
	for i, content := range service.Outputs {
		content.GenerateId()
		service.Outputs[i] = content
	}
}

func (hub *Hub) GenerateId() {
	hub.Id = URN_PREFIX + "hub:" + uuid.New().String()
}

func (protocol *Protocol) GenerateId() {
	if protocol.Id == "" {
		protocol.Id = URN_PREFIX + "protocol:" + uuid.New().String()
	}
	for i, segment := range protocol.ProtocolSegments {
		segment.GenerateId()
		protocol.ProtocolSegments[i] = segment
	}
}

func (segment *ProtocolSegment) GenerateId() {
	if segment.Id == "" {
		segment.Id = URN_PREFIX + "protocol-segment:" + uuid.New().String()
	}
}

func (content *Content) GenerateId() {
	if content.Id == "" {
		content.Id = URN_PREFIX + "content:" + uuid.New().String()
	}
	content.ContentVariable.GenerateId()
}

func (variable *ContentVariable) GenerateId() {
	if variable.Id == "" {
		variable.Id = URN_PREFIX + "content-variable:" + uuid.New().String()
	}
	for i, v := range variable.SubContentVariables {
		v.GenerateId()
		variable.SubContentVariables[i] = v
	}
}
