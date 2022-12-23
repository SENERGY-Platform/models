package models

type Content struct {
	Id                string          `json:"id"`
	ContentVariable   ContentVariable `json:"content_variable"`
	Serialization     Serialization   `json:"serialization"`
	ProtocolSegmentId string          `json:"protocol_segment_id"`
}

type Serialization string

const (
	XML       Serialization = "xml"
	JSON      Serialization = "json"
	PlainText Serialization = "plain-text"
)

func (this Serialization) Valid() bool {
	switch this {
	case XML:
		return true
	case JSON:
		return true
	case PlainText:
		return true
	default:
		return false
	}
}
