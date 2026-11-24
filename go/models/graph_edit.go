/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package models

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

const (
	// GraphEditModeChange makes an edit hold from the given instant on: what it replaces ends there.
	GraphEditModeChange = "change"
	// GraphEditModeCorrection makes an edit hold over the whole span of what it modifies, as if it had always been so.
	GraphEditModeCorrection = "correction"
)

// Edit applies body, the edited projection of the graph at the instant at, to the stored document this.
// The result is the new stored document; it is not validated, call Valid on it.
//
// The body is compared with this.At(at) by identity: nodes by Id, edges by (FromNodeId, ToNodeId) and
// attributes by key within their owner (a graph, a node or an edge). What the body has more is added,
// what it has less is removed, what differs is changed. Only the versions that are valid at the instant
// take part, every other version is left as it is.
//
// In mode correction a changed version is modified in place over its whole span, an added element gets no bounds
// but the end where a later version of the same identity starts (and the start where an earlier one ended),
// and a removed version is deleted.
//
// In mode change a changed version is split at the instant: the old one ends there, the new one runs from there
// to the end of the old one. An added element is valid from the instant until the next version of its identity,
// or open. A removed version ends at the instant.
//
// With until, which is only allowed in mode change, the edit holds for [at, until) exactly. That holds for every
// identity the edit touches, whichever versions of it lie in the span: they are cut at at and at until, what lies
// in between is replaced by the edit (a version wholly inside the span is dropped), and a version that goes on
// beyond until is resumed there as its own version with its old content. An added element and a changed version
// are valid for [at, until) only. Identities the edit does not touch are not touched, whatever lies in the span.
// A node has one validity only, so a node that is valid beyond until can not be removed for the span.
//
// The valid_from and valid_until of the body are not read for elements that exist in the stored projection,
// and refused for elements that are added: the bounds are set by the edit.
// A node's resource_type and resource_id may not change.
func (this Graph) Edit(body Graph, at time.Time, until *time.Time, mode string) (result Graph, err error) {
	switch mode {
	case GraphEditModeChange, GraphEditModeCorrection:
	default:
		return result, fmt.Errorf("unknown edit mode %q, use %v or %v", mode, GraphEditModeChange, GraphEditModeCorrection)
	}
	if until != nil {
		if mode != GraphEditModeChange {
			return result, fmt.Errorf("until is only allowed in mode %v", GraphEditModeChange)
		}
		if !until.After(at) {
			return result, errors.New("until must be after at")
		}
	}
	err = body.validEditIdentities()
	if err != nil {
		return result, err
	}
	err = this.validEditNodeResources(body)
	if err != nil {
		return result, err
	}
	if mode == GraphEditModeCorrection && !this.hasValidityBounds() && !body.hasValidityBounds() {
		//nothing has a history yet: the corrected state is the body, as it was before graphs had validities
		result = body
		result.Id = this.Id
		return result, nil
	}
	return this.edit(body, at, until, mode)
}

func (this Graph) hasValidityBounds() bool {
	return len(this.ValidityBounds()) > 0
}

// validEditIdentities checks that the body can be compared by identity.
func (this Graph) validEditIdentities() error {
	nodeIds := map[string]bool{}
	for _, node := range this.Nodes {
		if node.Id == "" {
			return errors.New("missing node id")
		}
		if nodeIds[node.Id] {
			return fmt.Errorf("duplicate node id: %v", node.Id)
		}
		nodeIds[node.Id] = true
	}
	edgeIds := map[string]bool{}
	pairs := map[[2]string]bool{}
	for _, edge := range this.Edges {
		if edge.Id == "" {
			return fmt.Errorf("missing edge id between %v and %v", edge.FromNodeId, edge.ToNodeId)
		}
		if edgeIds[edge.Id] {
			return fmt.Errorf("duplicate edge id: %v", edge.Id)
		}
		edgeIds[edge.Id] = true
		pair := [2]string{edge.FromNodeId, edge.ToNodeId}
		if pairs[pair] {
			return fmt.Errorf("duplicate edge between nodes: %v -> %v", edge.FromNodeId, edge.ToNodeId)
		}
		pairs[pair] = true
	}
	return nil
}

func (this Graph) validEditNodeResources(body Graph) error {
	for _, bodyNode := range body.Nodes {
		i := slices.IndexFunc(this.Nodes, func(n Node) bool { return n.Id == bodyNode.Id })
		if i < 0 {
			continue
		}
		stored := this.Nodes[i]
		if stored.ResourceType != bodyNode.ResourceType || stored.ResourceId != bodyNode.ResourceId {
			return fmt.Errorf("resource of node %v may not change, a different resource is a different node", bodyNode.Id)
		}
	}
	return nil
}

type graphEdit struct {
	at       time.Time
	until    *time.Time
	mode     string
	original Graph
	newId    func() string
}

func (this Graph) edit(body Graph, at time.Time, until *time.Time, mode string) (result Graph, err error) {
	e := graphEdit{at: at, until: until, mode: mode, original: this, newId: this.createId}
	result = this
	result.Owner = body.Owner
	if result.Owner == "" {
		result.Owner = this.Owner
	}
	result.Attributes, err = e.attributes(this.Attributes, body.Attributes)
	if err != nil {
		return result, err
	}
	var addedNodes, addedEdges map[string]bool
	result.Nodes, addedNodes, err = e.nodes(body.Nodes)
	if err != nil {
		return result, err
	}
	result.Edges, addedEdges, err = e.edges(body.Edges)
	if err != nil {
		return result, err
	}
	err = clampAdded(result.Nodes, result.Edges, addedNodes, addedEdges)
	if err != nil {
		return result, err
	}
	return result, nil
}

// clampAdded cuts the validity of the added edges to the intersection of the validities of their two nodes,
// and the validity of an added node to the span of its added outgoing edges: an edge that is valid where its node
// is not, or a node that is valid where it has no output, would make the graph invalid for the segments in between.
// The nodes and edges are modified in place.
func clampAdded(nodes []Node, edges []Edge, addedNodes map[string]bool, addedEdges map[string]bool) error {
	if len(addedEdges) == 0 {
		return nil
	}
	nodeIndex := map[string]int{}
	for i, node := range nodes {
		nodeIndex[node.Id] = i
	}
	clamp := func(from **time.Time, until **time.Time, otherFrom *time.Time, otherUntil *time.Time) {
		if otherFrom != nil && (*from == nil || otherFrom.After(**from)) {
			*from = otherFrom
		}
		if otherUntil != nil && (*until == nil || otherUntil.Before(**until)) {
			*until = otherUntil
		}
	}
	clampEdge := func(edge *Edge, onlyExisting bool) {
		for _, id := range []string{edge.FromNodeId, edge.ToNodeId} {
			i, ok := nodeIndex[id]
			if !ok || (onlyExisting && addedNodes[id]) {
				continue
			}
			clamp(&edge.ValidFrom, &edge.ValidUntil, nodes[i].ValidFrom, nodes[i].ValidUntil)
		}
	}
	for i := range edges {
		if addedEdges[edges[i].Id] {
			clampEdge(&edges[i], true)
		}
	}
	//an added node is valid as long as one of its added outputs is
	hulls := map[string][2]*time.Time{}
	open := map[string][2]bool{}
	for _, edge := range edges {
		if !addedEdges[edge.Id] || !addedNodes[edge.FromNodeId] {
			continue
		}
		hull, ok := hulls[edge.FromNodeId]
		o := open[edge.FromNodeId]
		if !ok {
			hull = [2]*time.Time{edge.ValidFrom, edge.ValidUntil}
			o = [2]bool{edge.ValidFrom == nil, edge.ValidUntil == nil}
		} else {
			if edge.ValidFrom == nil {
				o[0] = true
			} else if hull[0] != nil && edge.ValidFrom.Before(*hull[0]) {
				hull[0] = edge.ValidFrom
			}
			if edge.ValidUntil == nil {
				o[1] = true
			} else if hull[1] != nil && edge.ValidUntil.After(*hull[1]) {
				hull[1] = edge.ValidUntil
			}
		}
		hulls[edge.FromNodeId], open[edge.FromNodeId] = hull, o
	}
	for id, hull := range hulls {
		i := nodeIndex[id]
		o := open[id]
		if o[0] {
			hull[0] = nil
		}
		if o[1] {
			hull[1] = nil
		}
		clamp(&nodes[i].ValidFrom, &nodes[i].ValidUntil, hull[0], hull[1])
	}
	for i := range edges {
		if !addedEdges[edges[i].Id] {
			continue
		}
		clampEdge(&edges[i], false)
		if edges[i].ValidFrom != nil && edges[i].ValidUntil != nil && !edges[i].ValidUntil.After(*edges[i].ValidFrom) {
			return fmt.Errorf("edge %v -> %v is added but its nodes are not valid at the same time during the edit", edges[i].FromNodeId, edges[i].ToNodeId)
		}
	}
	return nil
}

func boundedAttributes(attributes []Attribute) bool {
	return slices.ContainsFunc(attributes, func(a Attribute) bool { return boundedEdit(a.ValidFrom, a.ValidUntil) })
}

func (this graphEdit) change() bool {
	return this.mode == GraphEditModeChange
}

// resumes reports whether a version that ends at versionUntil (nil: open) is resumed at the end of a bounded edit.
func (this graphEdit) resumes(versionUntil *time.Time) bool {
	return this.until != nil && (versionUntil == nil || this.until.Before(*versionUntil))
}

// spanEnd is the end of the edited part of a version that ends at versionUntil:
// a bounded edit holds until its own end, whatever the old version had.
func (this graphEdit) spanEnd(versionUntil *time.Time) *time.Time {
	if this.until != nil {
		return this.until
	}
	return versionUntil
}

func (this graphEdit) startsAtInstant(from *time.Time) bool {
	return from != nil && from.Equal(this.at)
}

func boundedEdit(from *time.Time, until *time.Time) bool {
	return from != nil || until != nil
}

// addedBounds returns the validity of an element that is added: identityFrom and identityUntil are the validities of
// the other versions of its identity, so the element is bounded by the neighbours.
func (this graphEdit) addedBounds(others [][2]*time.Time) (from *time.Time, until *time.Time) {
	var next, previous *time.Time
	for _, other := range others {
		if other[0] != nil && other[0].After(this.at) && (next == nil || other[0].Before(*next)) {
			next = other[0]
		}
		if other[1] != nil && !other[1].After(this.at) && (previous == nil || other[1].After(*previous)) {
			previous = other[1]
		}
	}
	until = next
	if this.change() {
		from = timePtr(this.at)
		if this.until != nil {
			//a bounded edit replaces whatever lies in its span
			until = this.until
		}
		return from, until
	}
	return previous, until
}

// attributes compares the attributes of one owner that are valid at the instant with the wanted ones.
func (this graphEdit) attributes(current []Attribute, want []Attribute) (result []Attribute, err error) {
	keys := []string{}
	live := map[string][]int{}
	for i, attribute := range current {
		if !attribute.ValidAt(this.at) {
			continue
		}
		if _, ok := live[attribute.Key]; !ok {
			keys = append(keys, attribute.Key)
		}
		live[attribute.Key] = append(live[attribute.Key], i)
	}
	wanted := map[string][]Attribute{}
	for _, attribute := range want {
		if _, ok := live[attribute.Key]; !ok {
			if _, ok := wanted[attribute.Key]; !ok {
				keys = append(keys, attribute.Key)
			}
		}
		wanted[attribute.Key] = append(wanted[attribute.Key], attribute)
	}

	replacements := map[int][]Attribute{}
	added := []Attribute{}
	for _, key := range keys {
		lives := slices.Clone(live[key])
		wants := slices.Clone(wanted[key])

		//unchanged ones first, so a legacy duplicate key is compared by content
		for wi := 0; wi < len(wants); {
			li := slices.IndexFunc(lives, func(i int) bool {
				return current[i].Value == wants[wi].Value && current[i].Origin == wants[wi].Origin
			})
			if li < 0 {
				wi++
				continue
			}
			lives = slices.Delete(lives, li, li+1)
			wants = slices.Delete(wants, wi, wi+1)
		}
		if len(lives) == 0 && len(wants) == 0 {
			continue
		}
		for _, w := range wants {
			if boundedEdit(w.ValidFrom, w.ValidUntil) {
				return nil, fmt.Errorf("attribute %v is added with valid_from or valid_until, the bounds are set by the edit", w.Key)
			}
		}

		if this.until != nil {
			//bounded edit: the versions of the key in the span are replaced by the wanted ones
			for i, a := range current {
				if a.Key == key && ValidityOverlaps(a.ValidFrom, a.ValidUntil, timePtr(this.at), this.until) {
					replacements[i] = this.cutAttribute(a)
				}
			}
			for _, w := range wanted[key] {
				added = append(added, Attribute{Key: w.Key, Value: w.Value, Origin: w.Origin, ValidFrom: timePtr(this.at), ValidUntil: this.until})
			}
			continue
		}

		for len(lives) > 0 && len(wants) > 0 {
			replacements[lives[0]] = this.changedAttribute(current[lives[0]], wants[0])
			lives = lives[1:]
			wants = wants[1:]
		}
		for _, i := range lives {
			replacements[i] = this.removedAttribute(current[i])
		}
		for _, w := range wants {
			others := [][2]*time.Time{}
			for _, c := range current {
				if c.Key == key {
					others = append(others, [2]*time.Time{c.ValidFrom, c.ValidUntil})
				}
			}
			from, until := this.addedBounds(others)
			added = append(added, Attribute{Key: w.Key, Value: w.Value, Origin: w.Origin, ValidFrom: from, ValidUntil: until})
		}
	}

	if current != nil || len(added) > 0 {
		result = make([]Attribute, 0, len(current)+len(added))
	}
	for i, attribute := range current {
		if replacement, ok := replacements[i]; ok {
			result = append(result, replacement...)
		} else {
			result = append(result, attribute)
		}
	}
	return append(result, added...), nil
}

// endedAttribute is the part of a live attribute version before the instant, nil if it is empty.
func (this graphEdit) endedAttribute(a Attribute) []Attribute {
	if this.startsAtInstant(a.ValidFrom) {
		return nil
	}
	a.ValidUntil = timePtr(this.at)
	return []Attribute{a}
}

func (this graphEdit) resumedAttribute(a Attribute) []Attribute {
	if !this.resumes(a.ValidUntil) {
		return nil
	}
	a.ValidFrom = timePtr(*this.until)
	return []Attribute{a}
}

// cutAttribute is what is left of an attribute version outside [at, until): the part before at and the part from until on.
func (this graphEdit) cutAttribute(a Attribute) (result []Attribute) {
	if a.ValidFrom == nil || a.ValidFrom.Before(this.at) {
		before := a
		before.ValidUntil = timePtr(this.at)
		if a.ValidUntil != nil && a.ValidUntil.Before(this.at) {
			before.ValidUntil = a.ValidUntil
		}
		result = append(result, before)
	}
	if a.ValidUntil == nil || a.ValidUntil.After(*this.until) {
		after := a
		after.ValidFrom = timePtr(*this.until)
		if a.ValidFrom != nil && a.ValidFrom.After(*this.until) {
			after.ValidFrom = a.ValidFrom
		}
		result = append(result, after)
	}
	return result
}

func (this graphEdit) removedAttribute(a Attribute) []Attribute {
	if !this.change() {
		return []Attribute{}
	}
	result := this.endedAttribute(a)
	return append(result, this.resumedAttribute(a)...)
}

func (this graphEdit) changedAttribute(a Attribute, w Attribute) []Attribute {
	changed := a
	changed.Value = w.Value
	changed.Origin = w.Origin
	if !this.change() {
		return []Attribute{changed}
	}
	result := this.endedAttribute(a)
	changed.ValidFrom = timePtr(this.at)
	changed.ValidUntil = this.spanEnd(a.ValidUntil)
	result = append(result, changed)
	return append(result, this.resumedAttribute(a)...)
}

// clipAttributes returns the attributes cut to [from, until): attributes that do not reach into the range
// are left out, the bounds of the others are narrowed to the range.
func clipAttributes(attributes []Attribute, from *time.Time, until *time.Time) []Attribute {
	if attributes == nil {
		return nil
	}
	result := make([]Attribute, 0, len(attributes))
	for _, a := range attributes {
		if !ValidityOverlaps(a.ValidFrom, a.ValidUntil, from, until) {
			continue
		}
		if from != nil && (a.ValidFrom == nil || a.ValidFrom.Before(*from)) {
			a.ValidFrom = from
		}
		if until != nil && (a.ValidUntil == nil || a.ValidUntil.After(*until)) {
			a.ValidUntil = until
		}
		result = append(result, a)
	}
	return result
}

func (this graphEdit) nodes(body []Node) (result []Node, addedIds map[string]bool, err error) {
	original := this.original.Nodes
	replacements := map[int][]Node{}
	added := []Node{}
	addedIds = map[string]bool{}
	bodyIds := map[string]bool{}
	for _, bodyNode := range body {
		bodyIds[bodyNode.Id] = true
		i := slices.IndexFunc(original, func(n Node) bool { return n.Id == bodyNode.Id })
		if i < 0 {
			if boundedEdit(bodyNode.ValidFrom, bodyNode.ValidUntil) {
				return nil, nil, fmt.Errorf("node %v is added with valid_from or valid_until, the bounds are set by the edit", bodyNode.Id)
			}
			if boundedAttributes(bodyNode.Attributes) {
				return nil, nil, fmt.Errorf("node %v is added with attributes that have valid_from or valid_until, the bounds are set by the edit", bodyNode.Id)
			}
			node := Node{Id: bodyNode.Id, ResourceType: bodyNode.ResourceType, ResourceId: bodyNode.ResourceId}
			if this.change() {
				node.ValidFrom = timePtr(this.at)
				node.ValidUntil = this.until
			}
			node.Attributes = slices.Clone(bodyNode.Attributes)
			added = append(added, node)
			addedIds[node.Id] = true
			continue
		}
		stored := original[i]
		if !stored.ValidAt(this.at) {
			return nil, nil, fmt.Errorf("node %v exists but is not valid at %v, the id of a node can not be used again", stored.Id, this.at.UTC().Format(time.RFC3339Nano))
		}
		stored.Attributes, err = this.attributes(stored.Attributes, bodyNode.Attributes)
		if err != nil {
			return nil, nil, fmt.Errorf("node %v: %w", stored.Id, err)
		}
		replacements[i] = []Node{stored}
	}
	for i, stored := range original {
		if !stored.ValidAt(this.at) || bodyIds[stored.Id] {
			continue
		}
		if !this.change() {
			replacements[i] = []Node{}
			continue
		}
		if this.resumes(stored.ValidUntil) {
			return nil, nil, fmt.Errorf("node %v can not be removed for a bounded span, it has one validity; remove it from the instant on", stored.Id)
		}
		if this.startsAtInstant(stored.ValidFrom) {
			replacements[i] = []Node{}
			continue
		}
		stored.ValidUntil = timePtr(this.at)
		replacements[i] = []Node{stored}
	}
	if original != nil || len(added) > 0 {
		result = make([]Node, 0, len(original)+len(added))
	}
	for i, node := range original {
		if replacement, ok := replacements[i]; ok {
			result = append(result, replacement...)
		} else {
			result = append(result, node)
		}
	}
	return append(result, added...), addedIds, nil
}

func (this graphEdit) edges(body []Edge) (result []Edge, addedIds map[string]bool, err error) {
	original := this.original.Edges
	replacements := map[int][]Edge{}
	added := []Edge{}
	addedIds = map[string]bool{}
	liveIndex := map[[2]string]int{}
	for i, stored := range original {
		if stored.ValidAt(this.at) {
			pair := [2]string{stored.FromNodeId, stored.ToNodeId}
			if _, ok := liveIndex[pair]; !ok {
				liveIndex[pair] = i
			}
		}
	}
	bodyPairs := map[[2]string]bool{}
	touched := map[[2]string]bool{}
	for _, bodyEdge := range body {
		pair := [2]string{bodyEdge.FromNodeId, bodyEdge.ToNodeId}
		bodyPairs[pair] = true
		i, ok := liveIndex[pair]
		if !ok {
			if boundedEdit(bodyEdge.ValidFrom, bodyEdge.ValidUntil) {
				return nil, nil, fmt.Errorf("edge %v -> %v is added with valid_from or valid_until, the bounds are set by the edit", bodyEdge.FromNodeId, bodyEdge.ToNodeId)
			}
			if boundedAttributes(bodyEdge.Attributes) {
				return nil, nil, fmt.Errorf("edge %v -> %v is added with attributes that have valid_from or valid_until, the bounds are set by the edit", bodyEdge.FromNodeId, bodyEdge.ToNodeId)
			}
			others := [][2]*time.Time{}
			for _, o := range original {
				if o.FromNodeId == bodyEdge.FromNodeId && o.ToNodeId == bodyEdge.ToNodeId {
					others = append(others, [2]*time.Time{o.ValidFrom, o.ValidUntil})
				}
			}
			from, until := this.addedBounds(others)
			touched[pair] = true
			id := bodyEdge.Id
			if this.edgeIdUsed(id) {
				id = this.newId()
			}
			addedIds[id] = true
			added = append(added, Edge{
				Id:         id,
				FromNodeId: bodyEdge.FromNodeId,
				ToNodeId:   bodyEdge.ToNodeId,
				Weight:     bodyEdge.Weight,
				Attributes: slices.Clone(bodyEdge.Attributes),
				ValidFrom:  from,
				ValidUntil: until,
			})
			continue
		}
		replacements[i], err = this.changedEdge(original[i], bodyEdge)
		if err != nil {
			return nil, nil, fmt.Errorf("edge %v -> %v: %w", bodyEdge.FromNodeId, bodyEdge.ToNodeId, err)
		}
		if original[i].Weight != bodyEdge.Weight {
			touched[pair] = true
		}
	}
	for pair, i := range liveIndex {
		if !bodyPairs[pair] {
			replacements[i] = this.removedEdge(original[i])
			touched[pair] = true
		}
	}
	if this.until != nil {
		//a bounded edit replaces whatever the touched identities have in its span
		for j, other := range original {
			pair := [2]string{other.FromNodeId, other.ToNodeId}
			if _, replaced := replacements[j]; replaced || !touched[pair] {
				continue
			}
			if !ValidityOverlaps(other.ValidFrom, other.ValidUntil, timePtr(this.at), this.until) {
				continue
			}
			if other.ValidUntil != nil && !other.ValidUntil.After(*this.until) {
				replacements[j] = []Edge{}
				continue
			}
			other.ValidFrom = timePtr(*this.until)
			other.Attributes = clipAttributes(other.Attributes, other.ValidFrom, other.ValidUntil)
			replacements[j] = []Edge{other}
		}
	}
	if original != nil || len(added) > 0 {
		result = make([]Edge, 0, len(original)+len(added))
	}
	for i, edge := range original {
		if replacement, ok := replacements[i]; ok {
			result = append(result, replacement...)
		} else {
			result = append(result, edge)
		}
	}
	return append(result, added...), addedIds, nil
}

// resumedEdge is the old content of a version that is resumed at the end of a bounded edit, under a new id.
func (this graphEdit) resumedEdge(old Edge) []Edge {
	if !this.resumes(old.ValidUntil) {
		return nil
	}
	resumed := old
	resumed.Id = this.newId()
	resumed.ValidFrom = timePtr(*this.until)
	resumed.Attributes = clipAttributes(old.Attributes, resumed.ValidFrom, resumed.ValidUntil)
	return []Edge{resumed}
}

func (this graphEdit) removedEdge(old Edge) []Edge {
	if !this.change() {
		return []Edge{}
	}
	result := []Edge{}
	if !this.startsAtInstant(old.ValidFrom) {
		ended := old
		ended.ValidUntil = timePtr(this.at)
		ended.Attributes = clipAttributes(old.Attributes, nil, ended.ValidUntil)
		result = append(result, ended)
	}
	return append(result, this.resumedEdge(old)...)
}

func (this graphEdit) changedEdge(old Edge, want Edge) (result []Edge, err error) {
	if !this.change() {
		edge := old
		edge.Weight = want.Weight
		edge.Attributes, err = this.attributes(old.Attributes, want.Attributes)
		if err != nil {
			return nil, err
		}
		if want.Id != old.Id && this.identityVersions(old) == 1 && !this.edgeIdUsed(want.Id) {
			//a writer that makes up its ids keeps doing so, as long as nothing else carries the id
			edge.Id = want.Id
		}
		return []Edge{edge}, nil
	}

	weightChanged := old.Weight != want.Weight
	if !weightChanged {
		//only the attributes can differ: they have their own versions
		edge := old
		edge.Attributes, err = this.attributes(old.Attributes, want.Attributes)
		if err != nil {
			return nil, err
		}
		return []Edge{edge}, nil
	}

	changed := old
	changed.Weight = want.Weight
	changed.ValidFrom = timePtr(this.at)
	changed.ValidUntil = this.spanEnd(old.ValidUntil)
	changed.Attributes = clipAttributes(old.Attributes, changed.ValidFrom, changed.ValidUntil)
	changed.Attributes, err = this.attributes(changed.Attributes, want.Attributes)
	if err != nil {
		return nil, err
	}
	result = []Edge{}
	if this.startsAtInstant(old.ValidFrom) {
		//the old version would be empty: the new one takes its place and its id
		changed.Id = old.Id
		changed.ValidFrom = old.ValidFrom
	} else {
		ended := old
		ended.ValidUntil = timePtr(this.at)
		ended.Attributes = clipAttributes(old.Attributes, nil, ended.ValidUntil)
		result = append(result, ended)
		changed.Id = this.newId()
	}
	result = append(result, changed)
	return append(result, this.resumedEdge(old)...), nil
}

func (this graphEdit) identityVersions(edge Edge) (count int) {
	for _, other := range this.original.Edges {
		if other.FromNodeId == edge.FromNodeId && other.ToNodeId == edge.ToNodeId {
			count++
		}
	}
	return count
}

func (this graphEdit) edgeIdUsed(id string) bool {
	return slices.ContainsFunc(this.original.Edges, func(edge Edge) bool { return edge.Id == id })
}
