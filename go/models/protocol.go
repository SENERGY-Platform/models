package models

type Protocol struct {
	Id               string            `json:"id"`
	Name             string            `json:"name"`
	Handler          string            `json:"handler"`
	ProtocolSegments []ProtocolSegment `json:"protocol_segments"`
	Constraints      []string          `json:"constraints"`
}

type ProtocolSegment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
