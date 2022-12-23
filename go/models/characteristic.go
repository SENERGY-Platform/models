package models

type Characteristic struct {
	Id                 string           `json:"id"`
	Name               string           `json:"name"`
	DisplayUnit        string           `json:"display_unit"`
	Type               Type             `json:"type"`
	MinValue           interface{}      `json:"min_value,omitempty"`
	MaxValue           interface{}      `json:"max_value,omitempty"`
	AllowedValues      []interface{}    `json:"allowed_values"`
	Value              interface{}      `json:"value,omitempty"`
	SubCharacteristics []Characteristic `json:"sub_characteristics"`
}
