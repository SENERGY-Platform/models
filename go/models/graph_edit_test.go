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
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"
)

func useCountingIds(t *testing.T) {
	old := GraphIdProvider
	n := 0
	GraphIdProvider = func() string {
		n++
		return fmt.Sprintf("new-%v", n)
	}
	t.Cleanup(func() { GraphIdProvider = old })
}

// parentGraph is a device x with one parent pa, next to the parents pb and pc, all below root.
func parentGraph() Graph {
	return Graph{
		Id:    "g",
		Owner: "test",
		Nodes: []Node{
			{Id: "root", ResourceType: "test"},
			{Id: "pa", ResourceType: "test"},
			{Id: "pb", ResourceType: "test"},
			{Id: "pc", ResourceType: "test"},
			{Id: "x", ResourceType: "device", ResourceId: "dx"},
		},
		Edges: []Edge{
			{Id: "pa>root", FromNodeId: "pa", ToNodeId: "root", Weight: 100},
			{Id: "pb>root", FromNodeId: "pb", ToNodeId: "root", Weight: 100},
			{Id: "pc>root", FromNodeId: "pc", ToNodeId: "root", Weight: 100},
			{Id: "x>pa", FromNodeId: "x", ToNodeId: "pa", Weight: 100},
		},
	}
}

func withoutEdges(g Graph, ids ...string) Graph {
	g.Edges = slices.DeleteFunc(slices.Clone(g.Edges), func(e Edge) bool { return slices.Contains(ids, e.Id) })
	return g
}

func TestGraphEdit(t *testing.T) {
	useCountingIds(t)

	// x -> pa 100 and pa/pb/pc -> root, but pa may also be fed by a meter m that splits 90/10 onto a and b
	meter := meterGraph()

	// the same, split on 06-30 into 90/10 before and 70/30 from then on
	splitMeter := meterGraph()
	splitMeter.Edges[2].ValidUntil = day("06-30")
	splitMeter.Edges[3].ValidUntil = day("06-30")
	splitMeter.Edges = append(splitMeter.Edges,
		Edge{Id: "m>a-2", FromNodeId: "m", ToNodeId: "a", Weight: 70, ValidFrom: day("06-30")},
		Edge{Id: "m>b-2", FromNodeId: "m", ToNodeId: "b", Weight: 30, ValidFrom: day("06-30")},
	)

	weights := func(g Graph, a int, b int) Graph {
		g.Edges = slices.Clone(g.Edges)
		for i, e := range g.Edges {
			if e.FromNodeId == "m" && e.ToNodeId == "a" {
				g.Edges[i].Weight = a
			}
			if e.FromNodeId == "m" && e.ToNodeId == "b" {
				g.Edges[i].Weight = b
			}
		}
		return g
	}
	// body of an edit: the projection at the instant, changed by mutate
	body := func(stored Graph, at string, mutate func(g *Graph)) Graph {
		g := stored.At(dayValue(at))
		g.Nodes = slices.Clone(g.Nodes)
		g.Edges = slices.Clone(g.Edges)
		if mutate != nil {
			mutate(&g)
		}
		return g
	}
	setNodeAttr := func(g *Graph, node string, a Attribute) {
		for i := range g.Nodes {
			if g.Nodes[i].Id == node {
				g.Nodes[i].Attributes = append(slices.Clone(g.Nodes[i].Attributes), a)
			}
		}
	}

	type expected struct {
		nodes []string
		edges []string
	}
	tests := []struct {
		name     string
		stored   Graph
		body     func(stored Graph) Graph
		at       string
		until    string
		mode     string
		expected expected
		err      string
	}{
		{
			name:   "change of the weights splits both edges at the instant",
			stored: meter, at: "06-30", mode: GraphEditModeChange,
			body: func(s Graph) Graph { return weights(body(s, "06-30", nil), 70, 30) },
			expected: expected{edges: []string{
				"a>root 100 [,)", "b>root 100 [,)",
				"m>a 90 [,06-30)", "m>a 70 [06-30,)", "m>b 10 [,06-30)", "m>b 30 [06-30,)",
			}},
		},
		{
			name:   "correction of the weights modifies them in place over the whole span",
			stored: meter, at: "06-30", mode: GraphEditModeCorrection,
			body: func(s Graph) Graph { return weights(body(s, "06-30", nil), 70, 30) },
			expected: expected{edges: []string{
				"a>root 100 [,)", "b>root 100 [,)", "m>a 70 [,)", "m>b 30 [,)",
			}},
		},
		{
			name:   "change of a version that ends later runs to the end of the old one",
			stored: splitMeter, at: "03-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph { return weights(body(s, "03-01", nil), 80, 20) },
			expected: expected{edges: []string{
				"a>root 100 [,)", "b>root 100 [,)",
				"m>a 90 [,03-01)", "m>a 80 [03-01,06-30)", "m>a 70 [06-30,)",
				"m>b 10 [,03-01)", "m>b 20 [03-01,06-30)", "m>b 30 [06-30,)",
			}},
		},
		{
			name:   "a change on the day a version starts replaces it, no empty version is left",
			stored: splitMeter, at: "06-30", mode: GraphEditModeChange,
			body: func(s Graph) Graph { return weights(body(s, "06-30", nil), 60, 40) },
			expected: expected{edges: []string{
				"a>root 100 [,)", "b>root 100 [,)",
				"m>a 90 [,06-30)", "m>a 60 [06-30,)", "m>b 10 [,06-30)", "m>b 40 [06-30,)",
			}},
		},
		{
			name:   "correction of an earlier version leaves the later one alone",
			stored: splitMeter, at: "03-01", mode: GraphEditModeCorrection,
			body: func(s Graph) Graph { return weights(body(s, "03-01", nil), 80, 20) },
			expected: expected{edges: []string{
				"a>root 100 [,)", "b>root 100 [,)",
				"m>a 80 [,06-30)", "m>a 70 [06-30,)", "m>b 20 [,06-30)", "m>b 30 [06-30,)",
			}},
		},
		{
			name:   "change adds a node and an edge from the instant on",
			stored: parentGraph(), at: "03-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) {
					g.Nodes = append(g.Nodes, Node{Id: "y", ResourceType: "device", ResourceId: "dy"})
					g.Edges = append(g.Edges, Edge{Id: "y>pb", FromNodeId: "y", ToNodeId: "pb", Weight: 100})
				})
			},
			expected: expected{
				nodes: []string{"pa", "pb", "pc", "root", "x", "y [03-01,)"},
				edges: []string{"pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,)", "y>pb 100 [03-01,)"},
			},
		},
		{
			name:   "correction adds them without bounds",
			stored: withHistory(parentGraph()), at: "03-01", mode: GraphEditModeCorrection,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) {
					g.Nodes = append(g.Nodes, Node{Id: "y", ResourceType: "device", ResourceId: "dy"})
					g.Edges = append(g.Edges, Edge{Id: "y>pb", FromNodeId: "y", ToNodeId: "pb", Weight: 100})
				})
			},
			expected: expected{
				nodes: []string{"old [,01-01)", "pa", "pb", "pc", "root", "x", "y"},
				edges: []string{"old>pa 100 [,01-01)", "pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,09-01)", "x>pc 100 [09-01,)", "y>pb 100 [,)"},
			},
		},
		{
			name:   "an added edge ends where a later version of its identity starts",
			stored: withHistory(parentGraph()), at: "03-01", mode: GraphEditModeCorrection,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) {
					g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.Id == "x>pa" })
					g.Edges = append(g.Edges, Edge{Id: "x>pc", FromNodeId: "x", ToNodeId: "pc", Weight: 100})
				})
			},
			expected: expected{
				nodes: []string{"old [,01-01)", "pa", "pb", "pc", "root", "x"},
				edges: []string{"old>pa 100 [,01-01)", "pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pc 100 [,09-01)", "x>pc 100 [09-01,)"},
			},
		},
		{
			name:   "an added edge starts where an earlier version of its identity ended",
			stored: withHistory(parentGraph()), at: "10-01", mode: GraphEditModeCorrection,
			body: func(s Graph) Graph {
				return body(s, "10-01", func(g *Graph) {
					g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.FromNodeId == "x" })
					g.Edges = append(g.Edges, Edge{Id: "x>pa", FromNodeId: "x", ToNodeId: "pa", Weight: 100})
				})
			},
			expected: expected{
				nodes: []string{"old [,01-01)", "pa", "pb", "pc", "root", "x"},
				edges: []string{"old>pa 100 [,01-01)", "pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,09-01)", "x>pa 100 [09-01,)"},
			},
		},
		{
			name:   "change adds an edge until the next version of its identity",
			stored: withHistory(parentGraph()), at: "03-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) {
					g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.Id == "x>pa" })
					g.Edges = append(g.Edges, Edge{Id: "x>pc", FromNodeId: "x", ToNodeId: "pc", Weight: 100})
				})
			},
			expected: expected{
				nodes: []string{"old [,01-01)", "pa", "pb", "pc", "root", "x"},
				edges: []string{"old>pa 100 [,01-01)", "pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,03-01)", "x>pc 100 [03-01,09-01)", "x>pc 100 [09-01,)"},
			},
		},
		{
			name:   "change removes a node and its edges from the instant on",
			stored: meter, at: "05-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				g := body(s, "05-01", func(g *Graph) {
					g.Nodes = slices.DeleteFunc(g.Nodes, func(n Node) bool { return n.Id == "b" })
					g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.ToNodeId == "b" || e.FromNodeId == "b" })
				})
				return weights(g, 100, 0)
			},
			expected: expected{
				nodes: []string{"a", "b [,05-01)", "m", "root"},
				edges: []string{"a>root 100 [,)", "b>root 100 [,05-01)", "m>a 90 [,05-01)", "m>a 100 [05-01,)", "m>b 10 [,05-01)"},
			},
		},
		{
			name:   "correction removes them as if they never were",
			stored: meter, at: "05-01", mode: GraphEditModeCorrection,
			body: func(s Graph) Graph {
				g := body(s, "05-01", func(g *Graph) {
					g.Nodes = slices.DeleteFunc(g.Nodes, func(n Node) bool { return n.Id == "b" })
					g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.ToNodeId == "b" || e.FromNodeId == "b" })
				})
				return weights(g, 100, 0)
			},
			expected: expected{
				nodes: []string{"a", "m", "root"},
				edges: []string{"a>root 100 [,)", "m>a 100 [,)"},
			},
		},
		{
			name:   "a change on the day an edge starts replaces it, no empty version is left",
			stored: withHistory(parentGraph()), at: "09-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "09-01", func(g *Graph) {
					g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.Id == "x>pc" })
					g.Edges = append(g.Edges, Edge{Id: "x>pb", FromNodeId: "x", ToNodeId: "pb", Weight: 100})
				})
			},
			expected: expected{
				nodes: []string{"old [,01-01)", "pa", "pb", "pc", "root", "x"},
				edges: []string{"old>pa 100 [,01-01)", "pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,09-01)", "x>pb 100 [09-01,)"},
			},
		},
		{
			name:   "a backdated change of an attribute holds until the next version",
			stored: attributeVersions(), at: "01-15", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "01-15", func(g *Graph) { g.Nodes[1].Attributes = []Attribute{{Key: "x", Value: "c"}} })
			},
			expected: expected{
				nodes: []string{"pa x=a[,01-15),x=c[01-15,03-01),x=b[03-01,)", "pb", "pc", "root", "x"},
				edges: []string{"pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,)"},
			},
		},
		{
			name:   "a backdated correction of an attribute changes the version in place",
			stored: attributeVersions(), at: "01-15", mode: GraphEditModeCorrection,
			body: func(s Graph) Graph {
				return body(s, "01-15", func(g *Graph) { g.Nodes[1].Attributes = []Attribute{{Key: "x", Value: "c"}} })
			},
			expected: expected{
				nodes: []string{"pa x=c[,03-01),x=b[03-01,)", "pb", "pc", "root", "x"},
				edges: []string{"pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,)"},
			},
		},
		{
			name:   "change adds an attribute from the instant on, removes one at the instant",
			stored: attributeVersions(), at: "04-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "04-01", func(g *Graph) {
					g.Nodes[1].Attributes = nil
					setNodeAttr(g, "pb", Attribute{Key: "y", Value: "1"})
				})
			},
			expected: expected{
				nodes: []string{"pa x=a[,03-01),x=b[03-01,04-01)", "pb y=1[04-01,)", "pc", "root", "x"},
				edges: []string{"pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,)"},
			},
		},
		{
			name:   "correction removes and adds attributes without bounds",
			stored: attributeVersions(), at: "04-01", mode: GraphEditModeCorrection,
			body: func(s Graph) Graph {
				return body(s, "04-01", func(g *Graph) {
					g.Nodes[1].Attributes = nil
					setNodeAttr(g, "pb", Attribute{Key: "y", Value: "1"})
				})
			},
			expected: expected{
				nodes: []string{"pa x=a[,03-01)", "pb y=1[,)", "pc", "root", "x"},
				edges: []string{"pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,)"},
			},
		},
		{
			name:   "an unchanged body changes nothing",
			stored: splitMeter, at: "07-15", mode: GraphEditModeChange,
			body: func(s Graph) Graph { return s.At(dayValue("07-15")) },
			expected: expected{
				nodes: []string{"a", "b", "m", "root"},
				edges: []string{"a>root 100 [,)", "b>root 100 [,)", "m>a 90 [,06-30)", "m>a 70 [06-30,)", "m>b 10 [,06-30)", "m>b 30 [06-30,)"},
			},
		},
		{
			name:   "the old graph from before the change of the weights is still there after a correction now",
			stored: splitMeter, at: "07-15", mode: GraphEditModeCorrection,
			body: func(s Graph) Graph {
				return body(s, "07-15", func(g *Graph) { setNodeAttr(g, "m", Attribute{Key: "label", Value: "main"}) })
			},
			expected: expected{
				nodes: []string{"a", "b", "m label=main[,)", "root"},
				edges: []string{"a>root 100 [,)", "b>root 100 [,)", "m>a 90 [,06-30)", "m>a 70 [06-30,)", "m>b 10 [,06-30)", "m>b 30 [06-30,)"},
			},
		},
		{
			name:   "bounded change: a month that was wrongly assigned becomes three versions",
			stored: parentGraph(), at: "01-01", until: "04-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "01-01", func(g *Graph) {
					g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.Id == "x>pa" })
					g.Edges = append(g.Edges, Edge{Id: "x>pb", FromNodeId: "x", ToNodeId: "pb", Weight: 100})
				})
			},
			expected: expected{
				nodes: []string{"pa", "pb", "pc", "root", "x"},
				edges: []string{"pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,01-01)", "x>pa 100 [04-01,)", "x>pb 100 [01-01,04-01)"},
			},
		},
		{
			name:   "bounded change over an existing boundary replaces what the touched identities have in the span",
			stored: splitMeter, at: "06-01", until: "08-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph { return weights(body(s, "06-01", nil), 50, 50) },
			expected: expected{
				nodes: []string{"a", "b", "m", "root"},
				edges: []string{
					"a>root 100 [,)", "b>root 100 [,)",
					"m>a 90 [,06-01)", "m>a 50 [06-01,08-01)", "m>a 70 [08-01,)",
					"m>b 10 [,06-01)", "m>b 50 [06-01,08-01)", "m>b 30 [08-01,)",
				},
			},
		},
		{
			name:   "bounded change does not touch an identity the body does not mention, so a conflict is refused",
			stored: withHistory(parentGraph()), at: "07-01", until: "10-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "07-01", func(g *Graph) {
					g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.Id == "x>pa" })
					g.Edges = append(g.Edges, Edge{Id: "x>pb", FromNodeId: "x", ToNodeId: "pb", Weight: 100})
				})
			},
			err: "invalid at 2025-09-01T00:00:00Z: sum of outgoing edge weights for node x must be 100, got 200",
		},
		{
			name:   "bounded change of an attribute over a boundary",
			stored: attributeVersions(), at: "02-01", until: "04-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "02-01", func(g *Graph) { g.Nodes[1].Attributes = []Attribute{{Key: "x", Value: "c"}} })
			},
			expected: expected{
				nodes: []string{"pa x=a[,02-01),x=c[02-01,04-01),x=b[04-01,)", "pb", "pc", "root", "x"},
				edges: []string{"pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,)"},
			},
		},
		{
			name:   "bounded change that ends inside a later version resumes it from there",
			stored: attributeVersions(), at: "01-15", until: "05-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "01-15", func(g *Graph) { g.Nodes[1].Attributes = nil })
			},
			expected: expected{
				nodes: []string{"pa x=a[,01-15),x=b[05-01,)", "pb", "pc", "root", "x"},
				edges: []string{"pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,)"},
			},
		},
		{
			name:   "bounded change adds a node for the span only",
			stored: parentGraph(), at: "03-01", until: "06-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) {
					g.Nodes = append(g.Nodes, Node{Id: "y", ResourceType: "device", ResourceId: "dy"})
					g.Edges = append(g.Edges, Edge{Id: "y>pb", FromNodeId: "y", ToNodeId: "pb", Weight: 100})
				})
			},
			expected: expected{
				nodes: []string{"pa", "pb", "pc", "root", "x", "y [03-01,06-01)"},
				edges: []string{"pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,)", "y>pb 100 [03-01,06-01)"},
			},
		},
		{
			name:   "bounded change can not remove a node that is valid beyond the span",
			stored: parentGraph(), at: "03-01", until: "06-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) {
					g.Nodes = slices.DeleteFunc(g.Nodes, func(n Node) bool { return n.Id == "x" })
					g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.FromNodeId == "x" })
				})
			},
			err: "node x can not be removed for a bounded span",
		},
		{
			name: "bounded change may remove a node that ends inside the span",
			stored: func() Graph {
				g := parentGraph()
				g.Nodes[4].ValidUntil = day("05-01")
				g.Edges[3].ValidUntil = day("05-01")
				return g
			}(),
			at: "03-01", until: "06-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) {
					g.Nodes = slices.DeleteFunc(g.Nodes, func(n Node) bool { return n.Id == "x" })
					g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.FromNodeId == "x" })
				})
			},
			expected: expected{
				nodes: []string{"pa", "pb", "pc", "root", "x [,03-01)"},
				edges: []string{"pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,03-01)"},
			},
		},
		{
			name:   "until is only allowed for a change",
			stored: parentGraph(), at: "03-01", until: "06-01", mode: GraphEditModeCorrection,
			body: func(s Graph) Graph { return s },
			err:  "until is only allowed in mode change",
		},
		{
			name:   "until must be after at",
			stored: parentGraph(), at: "03-01", until: "03-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph { return s },
			err:  "until must be after at",
		},
		{
			name:   "unknown mode",
			stored: parentGraph(), at: "03-01", mode: "replace",
			body: func(s Graph) Graph { return s },
			err:  `unknown edit mode "replace"`,
		},
		{
			name:   "the resource of a node may not change",
			stored: parentGraph(), at: "03-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) { g.Nodes[4].ResourceId = "other" })
			},
			err: "resource of node x may not change",
		},
		{
			name:   "the resource of a node may not change in a correction either",
			stored: parentGraph(), at: "03-01", mode: GraphEditModeCorrection,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) { g.Nodes[4].ResourceId = "other" })
			},
			err: "resource of node x may not change",
		},
		{
			name:   "an added element may not bring bounds",
			stored: parentGraph(), at: "03-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) {
					g.Nodes = append(g.Nodes, Node{Id: "y", ResourceType: "device", ResourceId: "dy", ValidFrom: day("01-01")})
				})
			},
			err: "node y is added with valid_from or valid_until",
		},
		{
			name:   "an added edge may not bring bounds",
			stored: parentGraph(), at: "03-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) {
					g.Edges = append(g.Edges, Edge{Id: "x>pb", FromNodeId: "x", ToNodeId: "pb", Weight: 100, ValidUntil: day("05-01")})
				})
			},
			err: "is added with valid_from or valid_until",
		},
		{
			name:   "an added attribute may not bring bounds",
			stored: parentGraph(), at: "03-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) { setNodeAttr(g, "pb", Attribute{Key: "k", Value: "v", ValidUntil: day("05-01")}) })
			},
			err: "attribute k is added with valid_from or valid_until",
		},
		{
			name:   "the id of an ended node can not be used again",
			stored: withHistory(parentGraph()), at: "10-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "10-01", func(g *Graph) {
					g.Nodes = append(g.Nodes, Node{Id: "old", ResourceType: "device", ResourceId: "dold"})
				})
			},
			err: "node old exists but is not valid at",
		},
		{
			name:   "a body with two edges between the same nodes",
			stored: parentGraph(), at: "03-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) {
					g.Edges = append(g.Edges, Edge{Id: "x>pa-2", FromNodeId: "x", ToNodeId: "pa", Weight: 100})
				})
			},
			err: "duplicate edge between nodes: x -> pa",
		},
		{
			name:   "a body edge without id",
			stored: parentGraph(), at: "03-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "03-01", func(g *Graph) { g.Edges[0].Id = "" })
			},
			err: "missing edge id between pa and root",
		},
		{
			name:   "a body edge that takes the id of an ended version gets a new id",
			stored: withHistory(parentGraph()), at: "10-01", mode: GraphEditModeChange,
			body: func(s Graph) Graph {
				return body(s, "10-01", func(g *Graph) {
					g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.FromNodeId == "x" })
					g.Edges = append(g.Edges, Edge{Id: "x>pa-ended", FromNodeId: "x", ToNodeId: "pb", Weight: 100})
				})
			},
			expected: expected{
				nodes: []string{"old [,01-01)", "pa", "pb", "pc", "root", "x"},
				edges: []string{"old>pa 100 [,01-01)", "pa>root 100 [,)", "pb>root 100 [,)", "pc>root 100 [,)", "x>pa 100 [,09-01)", "x>pb 100 [10-01,)", "x>pc 100 [09-01,10-01)"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stored := test.stored
			before, _ := json.Marshal(stored)
			var until *time.Time
			if test.until != "" {
				until = day(test.until)
			}
			result, err := stored.Edit(test.body(stored), dayValue(test.at), until, test.mode)
			after, _ := json.Marshal(stored)
			if string(before) != string(after) {
				t.Errorf("the stored document was modified\n%s\n%s", before, after)
			}
			if err == nil {
				err = result.Valid()
			}
			if test.err != "" {
				if err == nil || !strings.Contains(err.Error(), test.err) {
					t.Fatalf("\n%v\n%v", err, test.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if edges := describeEdges(result); !slices.Equal(edges, test.expected.edges) {
				t.Errorf("edges\n%#v\n%#v", edges, test.expected.edges)
			}
			if test.expected.nodes != nil {
				if nodes := describeNodes(result); !slices.Equal(nodes, test.expected.nodes) {
					t.Errorf("nodes\n%#v\n%#v", nodes, test.expected.nodes)
				}
			}
		})
	}
}

// withHistory returns the parent graph with a past: x was below the node old until 09-01 (old is an ended device)
// and x is below pc from 09-01 on, while the version below pa runs until 09-01.
func withHistory(g Graph) Graph {
	g.Nodes = append(g.Nodes, Node{Id: "old", ResourceType: "device", ResourceId: "dold", ValidUntil: day("01-01")})
	g.Edges[3].ValidUntil = day("09-01")
	g.Edges = append(g.Edges,
		Edge{Id: "x>pc", FromNodeId: "x", ToNodeId: "pc", Weight: 100, ValidFrom: day("09-01")},
		Edge{Id: "x>pa-ended", FromNodeId: "old", ToNodeId: "pa", Weight: 100, ValidUntil: day("01-01")},
	)
	g.Nodes[3].ValidFrom = nil
	return g
}

// attributeVersions is the parent graph with two versions of the attribute x on pa.
func attributeVersions() Graph {
	g := parentGraph()
	g.Nodes[1].Attributes = []Attribute{{Key: "x", Value: "a", ValidUntil: day("03-01")}, {Key: "x", Value: "b", ValidFrom: day("03-01")}}
	return g
}

func TestGraphEditRoundTrip(t *testing.T) {
	useCountingIds(t)
	splitMeter := meterGraph()
	splitMeter.Edges[2].ValidUntil = day("06-30")
	splitMeter.Edges[3].ValidUntil = day("06-30")
	splitMeter.Edges = append(splitMeter.Edges,
		Edge{Id: "m>a-2", FromNodeId: "m", ToNodeId: "a", Weight: 70, ValidFrom: day("06-30")},
		Edge{Id: "m>b-2", FromNodeId: "m", ToNodeId: "b", Weight: 30, ValidFrom: day("06-30")},
	)
	splitMeter.Nodes[3].ValidFrom = day("02-01")
	splitMeter.Nodes[3].Attributes = []Attribute{{Key: "k", Value: "1", ValidUntil: day("04-01")}, {Key: "k", Value: "2", ValidFrom: day("04-01")}}

	graphs := map[string]Graph{"unversioned": meterGraph(), "versioned": splitMeter, "history": withHistory(parentGraph())}
	for name, stored := range graphs {
		for _, at := range []string{"01-01", "02-01", "04-01", "06-30", "08-01", "12-31"} {
			for _, mode := range []string{GraphEditModeChange, GraphEditModeCorrection} {
				t.Run(fmt.Sprintf("%v %v %v: reading the graph at an instant and writing it back changes nothing", name, at, mode), func(t *testing.T) {
					result, err := stored.Edit(stored.At(dayValue(at)), dayValue(at), nil, mode)
					if err != nil {
						t.Fatal(err)
					}
					a, _ := json.Marshal(stored)
					b, _ := json.Marshal(result)
					if string(a) != string(b) {
						t.Errorf("\n%s\n%s", a, b)
					}
				})
			}
		}
	}

	t.Run("an unversioned graph is replaced by the body in a correction, as before", func(t *testing.T) {
		stored := meterGraph()
		body := meterGraph()
		body.Edges[0].Id = "other-id"
		body.Edges = body.Edges[:3]
		body.Edges[2].Weight = 100
		body.Nodes = slices.Delete(body.Nodes, 2, 3)
		body.Edges = slices.DeleteFunc(body.Edges, func(e Edge) bool { return e.FromNodeId == "b" })
		result, err := stored.Edit(body, dayValue("06-01"), nil, GraphEditModeCorrection)
		if err != nil {
			t.Fatal(err)
		}
		a, _ := json.Marshal(body)
		b, _ := json.Marshal(result)
		if string(a) != string(b) {
			t.Errorf("\n%s\n%s", a, b)
		}
	})

	t.Run("the general path gives the same graph as the shortcut for an unversioned graph", func(t *testing.T) {
		stored := meterGraph()
		body := meterGraph()
		body.Edges[3].Weight = 20
		body.Edges[2].Weight = 80
		body.Nodes = append(body.Nodes, Node{Id: "c", ResourceType: "test"})
		body.Edges = append(body.Edges, Edge{Id: "c>root", FromNodeId: "c", ToNodeId: "root", Weight: 100})
		body.Nodes[1].Attributes = []Attribute{{Key: "k", Value: "v"}}
		short, err := stored.Edit(body, dayValue("06-01"), nil, GraphEditModeCorrection)
		if err != nil {
			t.Fatal(err)
		}
		general, err := stored.edit(body, dayValue("06-01"), nil, GraphEditModeCorrection)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(describeEdges(short), describeEdges(general)) || !slices.Equal(describeNodes(short), describeNodes(general)) {
			t.Errorf("\n%v\n%v\n%v\n%v", describeEdges(short), describeEdges(general), describeNodes(short), describeNodes(general))
		}
	})

	t.Run("a correction keeps the ids a writer makes up while the edge has a single version", func(t *testing.T) {
		stored := splitMeter
		body := stored.At(dayValue("07-01"))
		body.Edges = slices.Clone(body.Edges)
		for i := range body.Edges {
			body.Edges[i].Id = "fresh-" + body.Edges[i].Id
		}
		result, err := stored.Edit(body, dayValue("07-01"), nil, GraphEditModeCorrection)
		if err != nil {
			t.Fatal(err)
		}
		ids := []string{}
		for _, e := range result.Edges {
			ids = append(ids, e.Id)
		}
		slices.Sort(ids)
		//a>root and b>root have one version, m>a and m>b have two: their ids stay
		expected := []string{"fresh-a>root", "fresh-b>root", "m>a", "m>a-2", "m>b", "m>b-2"}
		if !slices.Equal(ids, expected) {
			t.Errorf("\n%v\n%v", ids, expected)
		}
	})
}

func TestGraphEditAddedEdgesAreClampedToTheirNodes(t *testing.T) {
	useCountingIds(t)
	T, U, now := "03-01", "09-01", "06-01"
	// P is the end node, A is the old device, B the new one from T
	history := func(bUntil string) Graph {
		return Graph{Id: "g", Owner: "o",
			Nodes: []Node{
				{Id: "P", ResourceType: "location", ResourceId: "l"},
				{Id: "A", ResourceType: "device", ResourceId: "da", ValidUntil: day(T)},
				{Id: "B", ResourceType: "device", ResourceId: "db", ValidFrom: day(T), ValidUntil: day(bUntil)},
			},
			Edges: []Edge{
				{Id: "e1", FromNodeId: "A", ToNodeId: "P", Weight: 100, ValidUntil: day(T)},
				{Id: "e2", FromNodeId: "B", ToNodeId: "P", Weight: 100, ValidFrom: day(T), ValidUntil: day(bUntil)},
			},
		}
	}
	tests := []struct {
		name     string
		bUntil   string
		mode     string
		until    string
		child    string // the node the new device hangs below
		expected []string
		nodes    []string
	}{
		{"graph-provider case: correction without at, B is valid from T", "", GraphEditModeCorrection, "", "B",
			[]string{"A>P 100 [,03-01)", "B>P 100 [03-01,)", "D>B 100 [03-01,)"}, []string{"A [,03-01)", "B [03-01,)", "D [03-01,)", "P"}},
		{"change: B valid from T until U", U, GraphEditModeChange, "", "B",
			[]string{"A>P 100 [,03-01)", "B>P 100 [03-01,09-01)", "D>B 100 [06-01,09-01)"}, []string{"A [,03-01)", "B [03-01,09-01)", "D [06-01,09-01)", "P"}},
		{"correction: B valid from T until U", U, GraphEditModeCorrection, "", "B",
			[]string{"A>P 100 [,03-01)", "B>P 100 [03-01,09-01)", "D>B 100 [03-01,09-01)"}, []string{"A [,03-01)", "B [03-01,09-01)", "D [03-01,09-01)", "P"}},
		{"bounded change inside the validity of B", "", GraphEditModeChange, "07-01", "B",
			[]string{"A>P 100 [,03-01)", "B>P 100 [03-01,)", "D>B 100 [06-01,07-01)"}, []string{"A [,03-01)", "B [03-01,)", "D [06-01,07-01)", "P"}},
		{"bounded change reaching beyond the validity of B is cut to it", U, GraphEditModeChange, "12-01", "B",
			[]string{"A>P 100 [,03-01)", "B>P 100 [03-01,09-01)", "D>B 100 [06-01,09-01)"}, []string{"A [,03-01)", "B [03-01,09-01)", "D [06-01,09-01)", "P"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stored := history(test.bUntil)
			if err := stored.Valid(); err != nil {
				t.Fatal(err)
			}
			body := stored.At(dayValue(now))
			body.Nodes = append(slices.Clone(body.Nodes), Node{Id: "D", ResourceType: "device", ResourceId: "dd"})
			body.Edges = append(slices.Clone(body.Edges), Edge{Id: "e3", FromNodeId: "D", ToNodeId: test.child, Weight: 100})
			result, err := stored.Edit(body, dayValue(now), day(test.until), test.mode)
			if err != nil {
				t.Fatal(err)
			}
			if err = result.Valid(); err != nil {
				t.Fatal(err)
			}
			if edges := describeEdges(result); !slices.Equal(edges, test.expected) {
				t.Errorf("\n%#v\n%#v", edges, test.expected)
			}
			if nodes := describeNodes(result); !slices.Equal(nodes, test.nodes) {
				t.Errorf("\n%#v\n%#v", nodes, test.nodes)
			}
		})
	}

	t.Run("an edge between nodes that are never valid together is refused", func(t *testing.T) {
		stored := history("")
		body := stored.At(dayValue(now))
		body.Edges = append(slices.Clone(body.Edges), Edge{Id: "e3", FromNodeId: "B", ToNodeId: "A", Weight: 100})
		body.Nodes = slices.Clone(body.Nodes)
		_, err := stored.Edit(body, dayValue(now), nil, GraphEditModeChange)
		if err == nil {
			t.Error("expected an error")
		}
	})
}

func TestGraphEditAddedElementsRefuseBoundedAttributes(t *testing.T) {
	stored := parentGraph()
	bounded := []Attribute{{Key: "k", Value: "v", ValidUntil: day("05-01")}}
	tests := map[string]func(g *Graph){
		"node": func(g *Graph) {
			g.Nodes = append(g.Nodes, Node{Id: "y", ResourceType: "device", ResourceId: "dy", Attributes: bounded})
			g.Edges = append(g.Edges, Edge{Id: "y>pb", FromNodeId: "y", ToNodeId: "pb", Weight: 100})
		},
		"edge": func(g *Graph) {
			g.Nodes = append(g.Nodes, Node{Id: "y", ResourceType: "device", ResourceId: "dy"})
			g.Edges = append(g.Edges, Edge{Id: "y>pb", FromNodeId: "y", ToNodeId: "pb", Weight: 100, Attributes: bounded})
		},
		"graph": func(g *Graph) { g.Attributes = append(g.Attributes, bounded...) },
	}
	for name, mutate := range tests {
		for _, mode := range []string{GraphEditModeChange, GraphEditModeCorrection} {
			t.Run(name+" "+mode, func(t *testing.T) {
				body := stored.At(dayValue("03-01"))
				body.Nodes, body.Edges = slices.Clone(body.Nodes), slices.Clone(body.Edges)
				mutate(&body)
				_, err := stored.Edit(body, dayValue("03-01"), nil, mode)
				if err == nil || !strings.Contains(err.Error(), "valid_from or valid_until") {
					t.Errorf("%v", err)
				}
			})
		}
	}
}

func TestGraphEnsureValidEdgeWeightsChangesTheEdgeItFound(t *testing.T) {
	// the edge to fix is not the first edge of the graph and its attribute is not at the edge's index
	g := Graph{
		Edges: []Edge{
			{Id: "e0", FromNodeId: "z", ToNodeId: "p", Weight: 100},
			{Id: "e1", FromNodeId: "m", ToNodeId: "p", Weight: 60, Attributes: []Attribute{{Key: "other", Value: "x"}, {Key: GraphEdgeAttrSystemChanged, Value: "false"}}},
			{Id: "e2", FromNodeId: "m", ToNodeId: "q", Weight: 30},
		},
	}
	g.ensureValidEdgeWeights()
	if g.Edges[1].Weight != 70 || g.Edges[1].Attributes[1].Value != "true" || g.Edges[1].Attributes[0].Value != "x" {
		t.Errorf("%#v", g.Edges[1])
	}
	h := Graph{Edges: []Edge{
		{Id: "e0", FromNodeId: "m", ToNodeId: "p", Weight: 60, Attributes: []Attribute{{Key: GraphEdgeAttrSystemChanged, Value: "false"}}},
	}}
	h.ensureValidEdgeWeights() // used to index the attributes by the edge index and could panic
	if h.Edges[0].Weight != 100 {
		t.Errorf("%#v", h.Edges[0])
	}
}
