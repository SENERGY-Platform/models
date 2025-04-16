package models

type Selectable struct {
	Device             *DeviceWithDisplayName  `json:"device"`
	Services           []Service               `json:"services"`
	DeviceGroup        *DeviceGroup            `json:"device_group,omitempty"`
	Import             *Import                 `json:"import,omitempty"`
	ImportType         *ImportType             `json:"importType,omitempty"`
	ServicePathOptions map[string][]PathOption `json:"servicePathOptions,omitempty"`
}

type DeviceWithDisplayName struct {
	Device
	DisplayName string `json:"display_name"`
}

type FilterCriteriaAndSet []FilterCriteria

type FilterCriteriaOrSet []FilterCriteria

type FilterCriteria struct {
	FunctionId    string `json:"function_id"`
	DeviceClassId string `json:"device_class_id"`
	AspectId      string `json:"aspect_id"`
}

type FilterCriteriaWithInteraction struct {
	FilterCriteria
	Interaction Interaction `json:"interaction,omitempty"`
}

type BulkRequestElement struct {
	Id                string               `json:"id"`
	FilterInteraction *Interaction         `json:"filter_interaction"`
	FilterProtocols   []string             `json:"filter_protocols"`
	Criteria          FilterCriteriaAndSet `json:"criteria"`
	IncludeGroups     bool                 `json:"include_groups"`
	IncludeImports    bool                 `json:"include_imports"`
}

type BulkRequest []BulkRequestElement

type BulkRequestElementV2 struct {
	Id                       string                          `json:"id"`
	Criteria                 []FilterCriteriaWithInteraction `json:"criteria"`
	IncludeGroups            bool                            `json:"include_groups"`
	IncludeImports           bool                            `json:"include_imports"`
	IncludeDevices           bool                            `json:"include_devices"`
	IncludeIdModifiedDevices bool                            `json:"include_id_modified_devices"`
	LocalDevices             []string                        `json:"local_devices"`
}

type BulkRequestV2 []BulkRequestElementV2

type BulkResult []BulkResultElement

type BulkResultElement struct {
	Id          string       `json:"id"`
	Selectables []Selectable `json:"selectables"`
}

type PathOption struct {
	Path             string         `json:"path"`
	CharacteristicId string         `json:"characteristicId"`
	AspectNode       AspectNode     `json:"aspectNode"`
	FunctionId       string         `json:"functionId"`
	IsVoid           bool           `json:"isVoid"`
	Value            interface{}    `json:"value,omitempty"`
	Type             string         `json:"type,omitempty"`
	Configurables    []Configurable `json:"configurables,omitempty"`
}

type Configurable struct {
	Path             string      `json:"path"`
	CharacteristicId string      `json:"characteristic_id"`
	AspectNode       AspectNode  `json:"aspect_node"`
	FunctionId       string      `json:"function_id"`
	Value            interface{} `json:"value,omitempty"`
	Type             string      `json:"type,omitempty"`
}
