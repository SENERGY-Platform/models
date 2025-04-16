package models

type ImportType struct {
	Id             string                `json:"id"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	Image          string                `json:"image"`
	DefaultRestart bool                  `json:"default_restart"`
	Configs        []ImportConfig        `json:"configs"`
	Output         ImportContentVariable `json:"output"`
	Owner          string                `json:"owner"`
}

type ImportContentVariable struct {
	Name                string                  `json:"name"`
	Type                Type                    `json:"type"`
	CharacteristicId    string                  `json:"characteristic_id"`
	SubContentVariables []ImportContentVariable `json:"sub_content_variables"`
	UseAsTag            bool                    `json:"use_as_tag"`
	FunctionId          string                  `json:"function_id,omitempty"`
	AspectId            string                  `json:"aspect_id,omitempty"`
}

type ImportTypeConfig struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Type         Type        `json:"type"`
	DefaultValue interface{} `json:"default_value"`
}

type Import struct {
	Id           string         `json:"id"`
	Name         string         `json:"name"`
	ImportTypeId string         `json:"import_type_id"`
	Image        string         `json:"image"`
	KafkaTopic   string         `json:"kafka_topic"`
	Configs      []ImportConfig `json:"configs"`
	Restart      *bool          `json:"restart"`
}

type ImportConfig struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type ImportTypeFilterCriteria struct {
	FunctionId string `json:"function_id"`
	AspectId   string `json:"aspect_id"`
}

func (this ImportTypeFilterCriteria) Short() string {
	return this.AspectId + "_" + this.FunctionId
}
