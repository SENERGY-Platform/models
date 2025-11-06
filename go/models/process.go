/*
 * Copyright 2025 InfAI (CC SES)
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

const CurrentDeploymentModelVersion int64 = 3

type Deployment struct {
	Version          int64                   `json:"version"`
	Id               string                  `json:"id"`
	Name             string                  `json:"name"`
	Description      string                  `json:"description"`
	Diagram          Diagram                 `json:"diagram"`
	Elements         []Element               `json:"elements"`
	Executable       bool                    `json:"executable"`
	IncidentHandling *IncidentHandling       `json:"incident_handling,omitempty"`
	StartParameter   []ProcessStartParameter `json:"start_parameter,omitempty"`
}

type ProcessStartParameter struct {
	Id         string            `json:"id"`
	Label      string            `json:"label"`
	Type       string            `json:"type"`
	Default    string            `json:"default"`
	Properties map[string]string `json:"properties"`
}

type Diagram struct {
	XmlRaw      string `json:"xml_raw"`
	XmlDeployed string `json:"xml_deployed"`
	Svg         string `json:"svg"`
}

type Element struct {
	BpmnId           string            `json:"bpmn_id"`
	Group            *string           `json:"group"`
	Name             string            `json:"name"`
	Order            int64             `json:"order"`
	TimeEvent        *TimeEvent        `json:"time_event"`
	Notification     *Notification     `json:"notification"`
	MessageEvent     *MessageEvent     `json:"message_event"`
	ConditionalEvent *ConditionalEvent `json:"conditional_event"`
	Task             *Task             `json:"task"`
}

type TimeEvent struct {
	Type string `json:"type"`
	Time string `json:"time"`
}

type Notification struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type MessageEvent struct {
	Value         string `json:"value"`
	EventId       string `json:"event_id"`
	UseMarshaller bool   `json:"use_marshaller"`

	Selection Selection `json:"selection"`

	//TODO: implement usage in Process-Deployment, Frontend, Event-Deployment, Ewent-Worker (currently unused)
	//information if event should be triggered by analytics value
	AnalyticsSelection *AnalyticsSelection `json:"analytics_selection,omitempty"`
}

type AnalyticsSelection struct {
	PipelineId    string `json:"pipeline_id"`    //auth required
	OperatorId    string `json:"operator_id"`    //the id of the operator in the flow
	OperatorTopic string `json:"operator_topic"` // "analytics-<name-of-operator>"
	OutputName    string `json:"output_name"`
}

type ConditionalEvent struct {
	Script        string            `json:"script"`
	ValueVariable string            `json:"value_variable"`
	Variables     map[string]string `json:"variables"`
	Qos           int               `json:"qos"`
	EventId       string            `json:"event_id"`
	Selection     Selection         `json:"selection"`
}

type Task struct {
	Retries     int64             `json:"retries"`
	Parameter   map[string]string `json:"parameter"`
	Selection   Selection         `json:"selection"`
	PreferEvent bool              `json:"prefer_event,omitempty"`
}

type IncidentHandling struct {
	Restart bool `json:"restart"`
	Notify  bool `json:"notify"`
}

type Selection struct {
	FilterCriteria             ProcessFilterCriteria `json:"filter_criteria"`
	SelectionOptions           []SelectionOption     `json:"selection_options"`
	SelectedDeviceId           *string               `json:"selected_device_id"`
	SelectedServiceId          *string               `json:"selected_service_id"`
	SelectedDeviceGroupId      *string               `json:"selected_device_group_id"`
	SelectedImportId           *string               `json:"selected_import_id"`
	SelectedGenericEventSource *GenericEventSource   `json:"selected_generic_event_source"`
	SelectedPath               *PathOption           `json:"selected_path"`
}

type ProcessFilterCriteria struct {
	CharacteristicId *string `json:"characteristic_id"`
	FunctionId       *string `json:"function_id"`
	DeviceClassId    *string `json:"device_class_id"`
	AspectId         *string `json:"aspect_id"`
}

type SelectionOption struct {
	Device      *SelectionDevice        `json:"device"`
	Services    []SelectionService      `json:"services"`
	DeviceGroup *SelectionDeviceGroup   `json:"device_group"`
	Import      *Import                 `json:"import"`
	ImportType  *ImportType             `json:"importType"`
	PathOptions map[string][]PathOption `json:"path_options,omitempty"`
}

type GenericEventSource struct {
	FilterType string `json:"filter_type"`
	FilterIds  string `json:"filter_ids"`
	Topic      string `json:"topic"`
}

type SelectionDevice struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type SelectionDeviceGroup struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type SelectionService struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (this ProcessFilterCriteria) ToFilterCriteria() (result FilterCriteria) {
	if this.FunctionId != nil {
		result.FunctionId = *this.FunctionId
	}
	if this.DeviceClassId != nil {
		result.DeviceClassId = *this.DeviceClassId
	}
	if this.AspectId != nil {
		result.AspectId = *this.AspectId
	}
	return
}
