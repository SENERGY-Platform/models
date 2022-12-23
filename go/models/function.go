package models

type Function struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	ConceptId   string `json:"concept_id"`
	RdfType     string `json:"rdf_type"`
}
