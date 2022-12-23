package models

type ContentVariable struct {
	Id                   string            `json:"id"`
	Name                 string            `json:"name"`
	IsVoid               bool              `json:"is_void"`
	Type                 Type              `json:"type"`
	SubContentVariables  []ContentVariable `json:"sub_content_variables"`
	CharacteristicId     string            `json:"characteristic_id"`
	Value                interface{}       `json:"value"`
	SerializationOptions []string          `json:"serialization_options"`
	UnitReference        string            `json:"unit_reference,omitempty"`
	FunctionId           string            `json:"function_id,omitempty"`
	AspectId             string            `json:"aspect_id,omitempty"`
}


type Type string

const (
	String  Type = "https://schema.org/Text"
	Integer Type = "https://schema.org/Integer"
	Float   Type = "https://schema.org/Float"
	Boolean Type = "https://schema.org/Boolean"

	List      Type = "https://schema.org/ItemList"
	Structure Type = "https://schema.org/StructuredValue"
)
