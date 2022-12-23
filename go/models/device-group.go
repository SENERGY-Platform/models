package models

type DeviceGroup struct {
	Id            string                      `json:"id"`
	Name          string                      `json:"name"`
	Image         string                      `json:"image"`
	Criteria      []DeviceGroupFilterCriteria `json:"criteria"`
	DeviceIds     []string                    `json:"device_ids"`
	CriteriaShort []string                    `json:"criteria_short,omitempty"`
	Attributes    []Attribute                 `json:"attributes"`
}

func (this *DeviceGroup) SetShortCriteria() {
	this.CriteriaShort = []string{}
	for _, criteria := range this.Criteria {
		this.CriteriaShort = append(this.CriteriaShort, criteria.Short())
	}
}

type DeviceGroupFilterCriteria struct {
	Interaction   Interaction `json:"interaction"`
	FunctionId    string      `json:"function_id"`
	AspectId      string      `json:"aspect_id"`
	DeviceClassId string      `json:"device_class_id"`
}

func (this DeviceGroupFilterCriteria) Short() string {
	return this.FunctionId + "_" + this.AspectId + "_" + this.DeviceClassId + "_" + string(this.Interaction)
}
