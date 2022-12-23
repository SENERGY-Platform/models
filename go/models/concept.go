package models

type Concept struct {
	Id                   string               `json:"id"`
	Name                 string               `json:"name"`
	CharacteristicIds    []string             `json:"characteristic_ids"`
	BaseCharacteristicId string               `json:"base_characteristic_id"`
	Conversions          []ConverterExtension `json:"conversions"`
}

type ConverterExtension struct {
	From            string `json:"from"`
	To              string `json:"to"`
	Distance        int64  `json:"distance"`
	Formula         string `json:"formula"`
	PlaceholderName string `json:"placeholder_name"`
}

type ConceptWithCharacteristics struct {
	Id                   string               `json:"id"`
	Name                 string               `json:"name"`
	BaseCharacteristicId string               `json:"base_characteristic_id"`
	Characteristics      []Characteristic     `json:"characteristics"`
	Conversions          []ConverterExtension `json:"conversions"`
}
