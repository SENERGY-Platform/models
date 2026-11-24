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

// GraphSegment is the projection of a graph on [From, Until), a range in which no validity bound lies.
// A nil From or Until is unbounded and is written as null.
type GraphSegment struct {
	From  *time.Time `json:"from"`
	Until *time.Time `json:"until"`
	Graph Graph      `json:"graph"`
}

// ValidityContains reports whether t lies in [from, until); nil bounds are unbounded.
func ValidityContains(from *time.Time, until *time.Time, t time.Time) bool {
	if from != nil && t.Before(*from) {
		return false
	}
	if until != nil && !t.Before(*until) {
		return false
	}
	return true
}

// ValidityOverlaps reports whether [aFrom, aUntil) and [bFrom, bUntil) share an instant.
// Ranges that touch (a ends where b starts) do not overlap. nil bounds are unbounded.
func ValidityOverlaps(aFrom *time.Time, aUntil *time.Time, bFrom *time.Time, bUntil *time.Time) bool {
	if aUntil != nil && bFrom != nil && !aUntil.After(*bFrom) {
		return false
	}
	if bUntil != nil && aFrom != nil && !bUntil.After(*aFrom) {
		return false
	}
	return true
}

// ValidityIntersects reports whether the given ranges share an instant. Each range is [from, until) with nil as unbounded.
func ValidityIntersects(ranges ...[2]*time.Time) bool {
	var from, until *time.Time
	for _, r := range ranges {
		if r[0] != nil && (from == nil || r[0].After(*from)) {
			from = r[0]
		}
		if r[1] != nil && (until == nil || r[1].Before(*until)) {
			until = r[1]
		}
	}
	return from == nil || until == nil || from.Before(*until)
}

func (this Node) ValidAt(t time.Time) bool {
	return ValidityContains(this.ValidFrom, this.ValidUntil, t)
}

func (this Edge) ValidAt(t time.Time) bool {
	return ValidityContains(this.ValidFrom, this.ValidUntil, t)
}

func (this Attribute) ValidAt(t time.Time) bool {
	return ValidityContains(this.ValidFrom, this.ValidUntil, t)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func attributesAt(attributes []Attribute, t time.Time) []Attribute {
	if attributes == nil {
		return nil
	}
	result := make([]Attribute, 0, len(attributes))
	for _, attribute := range attributes {
		if attribute.ValidAt(t) {
			result = append(result, attribute)
		}
	}
	return result
}

// At is the projection of the graph at t: the nodes, edges and attributes whose validity contains t.
// The result is an ordinary graph; the validity fields of the kept elements are left as stored.
// It shares nothing with the receiver that could be modified by appending.
func (this Graph) At(t time.Time) Graph {
	result := this
	result.Attributes = attributesAt(this.Attributes, t)
	if this.Nodes != nil {
		result.Nodes = make([]Node, 0, len(this.Nodes))
		for _, node := range this.Nodes {
			if node.ValidAt(t) {
				node.Attributes = attributesAt(node.Attributes, t)
				result.Nodes = append(result.Nodes, node)
			}
		}
	}
	if this.Edges != nil {
		result.Edges = make([]Edge, 0, len(this.Edges))
		for _, edge := range this.Edges {
			if edge.ValidAt(t) {
				edge.Attributes = attributesAt(edge.Attributes, t)
				result.Edges = append(result.Edges, edge)
			}
		}
	}
	return result
}

// ValidityBounds returns every instant at which the projection may change: all valid_from and valid_until
// of nodes, edges and their attributes and of the graph attributes, sorted and without duplicates.
func (this Graph) ValidityBounds() []time.Time {
	bounds := []time.Time{}
	add := func(from *time.Time, until *time.Time) {
		if from != nil {
			bounds = append(bounds, *from)
		}
		if until != nil {
			bounds = append(bounds, *until)
		}
	}
	for _, attribute := range this.Attributes {
		add(attribute.ValidFrom, attribute.ValidUntil)
	}
	for _, node := range this.Nodes {
		add(node.ValidFrom, node.ValidUntil)
		for _, attribute := range node.Attributes {
			add(attribute.ValidFrom, attribute.ValidUntil)
		}
	}
	for _, edge := range this.Edges {
		add(edge.ValidFrom, edge.ValidUntil)
		for _, attribute := range edge.Attributes {
			add(attribute.ValidFrom, attribute.ValidUntil)
		}
	}
	slices.SortFunc(bounds, func(a, b time.Time) int { return a.Compare(b) })
	return slices.CompactFunc(bounds, func(a, b time.Time) bool { return a.Equal(b) })
}

// Segments cuts [start, end) at every validity bound inside it and returns each piece with its projection.
// A nil start or end leaves the range open on that side, so without both the result is the whole history.
// Pieces are not merged, even where two neighbours have the same projection.
func (this Graph) Segments(start *time.Time, end *time.Time) (result []GraphSegment, err error) {
	if start != nil && end != nil && !end.After(*start) {
		return nil, errors.New("end must be after start")
	}
	points := []*time.Time{start}
	for _, bound := range this.ValidityBounds() {
		if (start == nil || bound.After(*start)) && (end == nil || bound.Before(*end)) {
			points = append(points, timePtr(bound))
		}
	}
	points = append(points, end)
	result = make([]GraphSegment, 0, len(points)-1)
	for i := 0; i < len(points)-1; i++ {
		from, until := points[i], points[i+1]
		var instant time.Time
		switch {
		case from != nil:
			instant = *from
		case until != nil:
			instant = until.Add(-time.Nanosecond) //no bound lies before the first one
		}
		result = append(result, GraphSegment{From: from, Until: until, Graph: this.At(instant)})
	}
	return result, nil
}

// Valid checks the graph.
//
// For a graph without validities that is the check of a plain graph. A graph with validities is checked for each
// segment between two bounds of the document: the projection of every segment must be a valid plain graph
// (weights of 0 or 100 per node, exactly one end node that is no device or thing, no loop, unique resource ids,
// no duplicate edge, edges only between nodes that are valid at the same time).
// A projection without any node is valid, the graph does not exist then.
// On top of that valid_until must be after valid_from, ids are unique over all versions, versions of one edge
// (identified by its two nodes) do not overlap, and neither do versions of one attribute key of one element.
// The error of a segment names the first invalid instant.
func (this *Graph) Valid() error {
	if this.Owner == "" {
		return errors.New("graph owner is required")
	}
	err := this.validValidityRanges()
	if err != nil {
		return err
	}
	bounds := this.ValidityBounds()
	if len(bounds) == 0 {
		return this.validProjection()
	}
	err = this.validUniqueIds()
	if err != nil {
		return err
	}
	err = this.validEdgeVersions()
	if err != nil {
		return err
	}
	err = this.validAttributeVersions()
	if err != nil {
		return err
	}

	//before the first bound, then from each bound on
	instants := []time.Time{bounds[0].Add(-time.Nanosecond)}
	instants = append(instants, bounds...)
	for i, instant := range instants {
		projection := this.At(instant)
		if len(projection.Nodes) == 0 && len(this.Nodes) > 0 {
			//the graph does not exist at this instant
			continue
		}
		err = projection.validProjection()
		if err != nil {
			if i == 0 {
				return fmt.Errorf("invalid before %v: %w", bounds[0].UTC().Format(time.RFC3339Nano), err)
			}
			return fmt.Errorf("invalid at %v: %w", instant.UTC().Format(time.RFC3339Nano), err)
		}
	}
	return nil
}

func validRange(from *time.Time, until *time.Time, owner string) error {
	if from != nil && until != nil && !until.After(*from) {
		return fmt.Errorf("valid_until must be after valid_from in %v", owner)
	}
	return nil
}

func validAttributeRanges(attributes []Attribute, owner string) error {
	for _, attribute := range attributes {
		err := validRange(attribute.ValidFrom, attribute.ValidUntil, fmt.Sprintf("attribute %v of %v", attribute.Key, owner))
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Graph) validValidityRanges() error {
	err := validAttributeRanges(this.Attributes, "graph "+this.Id)
	if err != nil {
		return err
	}
	for _, node := range this.Nodes {
		err = validRange(node.ValidFrom, node.ValidUntil, "node "+node.Id)
		if err != nil {
			return err
		}
		err = validAttributeRanges(node.Attributes, "node "+node.Id)
		if err != nil {
			return err
		}
	}
	for _, edge := range this.Edges {
		err = validRange(edge.ValidFrom, edge.ValidUntil, "edge "+edge.Id)
		if err != nil {
			return err
		}
		err = validAttributeRanges(edge.Attributes, "edge "+edge.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

// validUniqueIds checks the ids over all versions; the projections only see the ids that are valid at one instant.
func (this *Graph) validUniqueIds() error {
	nodeIds := map[string]bool{}
	for _, node := range this.Nodes {
		if nodeIds[node.Id] {
			return fmt.Errorf("duplicate node id: %v", node.Id)
		}
		nodeIds[node.Id] = true
	}
	edgeIds := map[string]bool{}
	for _, edge := range this.Edges {
		if edgeIds[edge.Id] {
			return fmt.Errorf("duplicate edge id: %v", edge.Id)
		}
		edgeIds[edge.Id] = true
	}
	return nil
}

// validEdgeVersions rejects versions of one edge, identified by its two nodes, that overlap.
func (this *Graph) validEdgeVersions() error {
	for i, a := range this.Edges {
		for _, b := range this.Edges[:i] {
			if a.FromNodeId == b.FromNodeId && a.ToNodeId == b.ToNodeId && ValidityOverlaps(a.ValidFrom, a.ValidUntil, b.ValidFrom, b.ValidUntil) {
				return fmt.Errorf("versions %v and %v of the edge %v -> %v overlap", b.Id, a.Id, a.FromNodeId, a.ToNodeId)
			}
		}
	}
	return nil
}

// validAttributeVersions rejects attributes of one element with the same key whose validities overlap.
// Two attributes with the same key and no bounds at all are no versions of each other, they are duplicates as they
// always were allowed, so the rule only applies if one of the two has a bound.
func (this *Graph) validAttributeVersions() error {
	err := validAttributeVersionList(this.Attributes, "graph "+this.Id)
	if err != nil {
		return err
	}
	for _, node := range this.Nodes {
		err = validAttributeVersionList(node.Attributes, "node "+node.Id)
		if err != nil {
			return err
		}
	}
	for _, edge := range this.Edges {
		err = validAttributeVersionList(edge.Attributes, "edge "+edge.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func validAttributeVersionList(attributes []Attribute, owner string) error {
	for i, a := range attributes {
		for _, b := range attributes[:i] {
			if a.Key != b.Key {
				continue
			}
			bounded := a.ValidFrom != nil || a.ValidUntil != nil || b.ValidFrom != nil || b.ValidUntil != nil
			if bounded && ValidityOverlaps(a.ValidFrom, a.ValidUntil, b.ValidFrom, b.ValidUntil) {
				return fmt.Errorf("versions of attribute %v of %v overlap", a.Key, owner)
			}
		}
	}
	return nil
}
