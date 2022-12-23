package models

type Aspect struct {
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	SubAspects []Aspect `json:"sub_aspects"`
}

type AspectNode struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	RootId        string   `json:"root_id"`
	ParentId      string   `json:"parent_id"`
	ChildIds      []string `json:"child_ids"`
	AncestorIds   []string `json:"ancestor_ids"`
	DescendentIds []string `json:"descendent_ids"`
}
