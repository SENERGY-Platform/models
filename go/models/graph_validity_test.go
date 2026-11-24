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

// day returns the given day of 2025 as RFC3339 midnight UTC, e.g. day("06-30"); "" is nil.
func day(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, "2025-"+s+"T00:00:00Z")
	if err != nil {
		panic(err)
	}
	return &t
}

func dayValue(s string) time.Time {
	return *day(s)
}

func formatBound(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format("01-02")
}

func describeEdge(e Edge) string {
	attrs := []string{}
	for _, a := range e.Attributes {
		attrs = append(attrs, fmt.Sprintf("%v=%v[%v,%v)", a.Key, a.Value, formatBound(a.ValidFrom), formatBound(a.ValidUntil)))
	}
	return fmt.Sprintf("%v>%v %v [%v,%v) %v", e.FromNodeId, e.ToNodeId, e.Weight, formatBound(e.ValidFrom), formatBound(e.ValidUntil), strings.Join(attrs, ","))
}

// describeEdges lists the edges sorted by identity and start, without ids.
func describeEdges(g Graph) []string {
	edges := slices.Clone(g.Edges)
	slices.SortFunc(edges, func(a, b Edge) int {
		if c := strings.Compare(a.FromNodeId+">"+a.ToNodeId, b.FromNodeId+">"+b.ToNodeId); c != 0 {
			return c
		}
		return strings.Compare(formatBound(a.ValidFrom), formatBound(b.ValidFrom))
	})
	result := []string{}
	for _, e := range edges {
		result = append(result, strings.TrimSpace(describeEdge(e)))
	}
	return result
}

func describeNodes(g Graph) []string {
	result := []string{}
	for _, n := range g.Nodes {
		attrs := slices.Clone(n.Attributes)
		slices.SortFunc(attrs, func(a, b Attribute) int {
			return strings.Compare(a.Key+formatBound(a.ValidFrom), b.Key+formatBound(b.ValidFrom))
		})
		parts := []string{n.Id}
		if n.ValidFrom != nil || n.ValidUntil != nil {
			parts = append(parts, fmt.Sprintf("[%v,%v)", formatBound(n.ValidFrom), formatBound(n.ValidUntil)))
		}
		list := []string{}
		for _, a := range attrs {
			list = append(list, fmt.Sprintf("%v=%v[%v,%v)", a.Key, a.Value, formatBound(a.ValidFrom), formatBound(a.ValidUntil)))
		}
		if len(list) > 0 {
			parts = append(parts, strings.Join(list, ","))
		}
		result = append(result, strings.Join(parts, " "))
	}
	slices.Sort(result)
	return result
}

// meterGraph is root <- a, b and a meter m that feeds a with 90 and b with 10.
func meterGraph() Graph {
	return Graph{
		Id:    "g",
		Owner: "test",
		Nodes: []Node{
			{Id: "root", ResourceType: "test"},
			{Id: "a", ResourceType: "test"},
			{Id: "b", ResourceType: "test"},
			{Id: "m", ResourceType: "test"},
		},
		Edges: []Edge{
			{Id: "a>root", FromNodeId: "a", ToNodeId: "root", Weight: 100},
			{Id: "b>root", FromNodeId: "b", ToNodeId: "root", Weight: 100},
			{Id: "m>a", FromNodeId: "m", ToNodeId: "a", Weight: 90},
			{Id: "m>b", FromNodeId: "m", ToNodeId: "b", Weight: 10},
		},
	}
}

func TestGraphProjection(t *testing.T) {
	g := meterGraph()
	g.Nodes[3].ValidFrom = day("03-01")
	g.Nodes[3].ValidUntil = day("09-01")
	g.Edges[2].ValidFrom = day("03-01")
	g.Edges[3].ValidFrom = day("03-01")
	g.Attributes = []Attribute{{Key: "k", Value: "v1", ValidUntil: day("06-01")}, {Key: "k", Value: "v2", ValidFrom: day("06-01")}}

	tests := []struct {
		name          string
		at            string
		expectedNodes []string
		expectedEdges []string
		expectedAttr  string
	}{
		{"before the start", "02-28", []string{"a", "b", "root"}, []string{"a>root 100 [,)", "b>root 100 [,)"}, "v1"},
		{"start is included", "03-01", []string{"a", "b", "m", "root"}, []string{"a>root 100 [,)", "b>root 100 [,)", "m>a 90 [03-01,)", "m>b 10 [03-01,)"}, "v1"},
		{"attribute bound is the next version", "06-01", []string{"a", "b", "m", "root"}, []string{"a>root 100 [,)", "b>root 100 [,)", "m>a 90 [03-01,)", "m>b 10 [03-01,)"}, "v2"},
		{"end is excluded", "09-01", []string{"a", "b", "root"}, []string{"a>root 100 [,)", "b>root 100 [,)", "m>a 90 [03-01,)", "m>b 10 [03-01,)"}, "v2"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			projection := g.At(dayValue(test.at))
			nodes := []string{}
			for _, n := range projection.Nodes {
				nodes = append(nodes, n.Id)
			}
			slices.Sort(nodes)
			if !slices.Equal(nodes, test.expectedNodes) {
				t.Errorf("\n%#v\n%#v", nodes, test.expectedNodes)
			}
			if edges := describeEdges(projection); !slices.Equal(edges, test.expectedEdges) {
				t.Errorf("\n%#v\n%#v", edges, test.expectedEdges)
			}
			if len(projection.Attributes) != 1 || projection.Attributes[0].Value != test.expectedAttr {
				t.Errorf("%#v", projection.Attributes)
			}
		})
	}

	t.Run("a graph without validities is its own projection", func(t *testing.T) {
		g := meterGraph()
		a, _ := json.Marshal(g)
		b, _ := json.Marshal(g.At(dayValue("06-01")))
		if string(a) != string(b) {
			t.Errorf("\n%s\n%s", a, b)
		}
	})

	t.Run("nil slices stay nil, so old readers see what they saw", func(t *testing.T) {
		g := Graph{Id: "x", Owner: "o"}
		b, _ := json.Marshal(g.At(dayValue("06-01")))
		a, _ := json.Marshal(g)
		if string(a) != string(b) {
			t.Errorf("\n%s\n%s", a, b)
		}
	})

	t.Run("json omits unset bounds and writes set ones", func(t *testing.T) {
		b, _ := json.Marshal(Edge{Id: "e", ValidFrom: day("03-01")})
		if !strings.Contains(string(b), `"valid_from":"2025-03-01T00:00:00Z"`) || strings.Contains(string(b), "valid_until") {
			t.Errorf("%s", b)
		}
		b, _ = json.Marshal(Attribute{Key: "k"})
		if strings.Contains(string(b), "valid_") {
			t.Errorf("%s", b)
		}
	})
}

func TestGraphProjectionWithoutNodes(t *testing.T) {
	g := meterGraph()
	g.Attributes = []Attribute{{Key: "k", Value: "early", ValidUntil: day("03-01")}, {Key: "k", Value: "late", ValidFrom: day("03-01")}}
	for i := range g.Nodes {
		g.Nodes[i].ValidFrom = day("03-01")
		g.Nodes[i].ValidUntil = day("09-01")
	}
	for i := range g.Edges {
		g.Edges[i].ValidFrom = day("03-01")
		g.Edges[i].ValidUntil = day("09-01")
	}

	for _, at := range []string{"02-01", "09-01"} {
		t.Run("at "+at+" no node is valid: the graph is there, without content", func(t *testing.T) {
			projection := g.At(dayValue(at))
			if projection.Id != "g" || projection.Owner != "test" || len(projection.Nodes) != 0 || len(projection.Edges) != 0 {
				t.Errorf("%#v", projection)
			}
			b, _ := json.Marshal(projection)
			if !strings.Contains(string(b), `"nodes":[]`) || !strings.Contains(string(b), `"edges":[]`) {
				t.Errorf("empty lists must be [] and not null: %s", b)
			}
		})
	}
	t.Run("its attributes are those valid then", func(t *testing.T) {
		projection := g.At(dayValue("02-01"))
		if len(projection.Attributes) != 1 || projection.Attributes[0].Value != "early" {
			t.Errorf("%#v", projection.Attributes)
		}
	})
	t.Run("a segment without nodes is valid", func(t *testing.T) {
		if err := g.Valid(); err != nil {
			t.Error(err)
		}
	})
	t.Run("a segment with nodes must still be valid", func(t *testing.T) {
		g := g
		g.Edges = slices.Clone(g.Edges)
		g.Edges[3].Weight = 20
		err := g.Valid()
		if err == nil || !strings.HasPrefix(err.Error(), "invalid at 2025-03-01T00:00:00Z: ") {
			t.Errorf("%v", err)
		}
	})
	t.Run("a graph that has no node at all is invalid as before", func(t *testing.T) {
		err := (&Graph{Id: "g", Owner: "o"}).Valid()
		if err == nil {
			t.Error("expected an error")
		}
	})
}

func TestGraphSegments(t *testing.T) {
	g := meterGraph()
	g.Nodes[3].ValidFrom = day("03-01")
	g.Nodes[3].ValidUntil = day("09-01")
	g.Edges[2].ValidFrom = day("03-01")
	g.Edges[2].ValidUntil = day("09-01")
	g.Edges[3].ValidFrom = day("03-01")
	g.Edges[3].ValidUntil = day("09-01")
	g.Attributes = []Attribute{{Key: "k", Value: "v", ValidFrom: day("06-01")}}

	tests := []struct {
		name     string
		start    string      //"" is open
		end      string      //"" is open
		expected [][3]string //from, until, number of nodes; "" is open
		err      bool
	}{
		{"cut at every bound inside, attribute bounds included", "01-01", "12-31", [][3]string{{"01-01", "03-01", "3"}, {"03-01", "06-01", "4"}, {"06-01", "09-01", "4"}, {"09-01", "12-31", "3"}}, false},
		{"a bound at the start or end does not cut", "03-01", "09-01", [][3]string{{"03-01", "06-01", "4"}, {"06-01", "09-01", "4"}}, false},
		{"inside one piece", "04-01", "05-01", [][3]string{{"04-01", "05-01", "4"}}, false},
		{"end excluded: a piece ending where the node ends", "08-01", "09-01", [][3]string{{"08-01", "09-01", "4"}}, false},
		{"without start the first piece is unbounded", "", "07-01", [][3]string{{"", "03-01", "3"}, {"03-01", "06-01", "4"}, {"06-01", "07-01", "4"}}, false},
		{"without end the last piece is unbounded", "08-01", "", [][3]string{{"08-01", "09-01", "4"}, {"09-01", "", "3"}}, false},
		{"without both it is the whole history", "", "", [][3]string{{"", "03-01", "3"}, {"03-01", "06-01", "4"}, {"06-01", "09-01", "4"}, {"09-01", "", "3"}}, false},
		{"empty range", "05-01", "05-01", nil, true},
		{"reversed range", "05-02", "05-01", nil, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			segments, err := g.Segments(day(test.start), day(test.end))
			if (err != nil) != test.err {
				t.Fatal(err)
			}
			actual := [][3]string{}
			for _, s := range segments {
				actual = append(actual, [3]string{formatBound(s.From), formatBound(s.Until), fmt.Sprint(len(s.Graph.Nodes))})
			}
			if test.expected == nil {
				test.expected = [][3]string{}
			}
			if !slices.Equal(actual, test.expected) {
				t.Errorf("\n%v\n%v", actual, test.expected)
			}
		})
	}

	t.Run("an open side is written as null", func(t *testing.T) {
		segments, _ := g.Segments(nil, nil)
		b, _ := json.Marshal(segments[0])
		if !strings.HasPrefix(string(b), `{"from":null,"until":"2025-03-01T00:00:00Z","graph":`) {
			t.Errorf("%s", b)
		}
	})

	t.Run("a graph without bounds is one piece", func(t *testing.T) {
		segments, err := meterGraph().Segments(nil, nil)
		if err != nil || len(segments) != 1 || segments[0].From != nil || segments[0].Until != nil || len(segments[0].Graph.Nodes) != 4 {
			t.Errorf("%#v %v", segments, err)
		}
	})

	t.Run("each piece carries the projection at its start", func(t *testing.T) {
		segments, _ := g.Segments(day("01-01"), day("12-31"))
		if len(segments[0].Graph.Attributes) != 0 || len(segments[2].Graph.Attributes) != 1 {
			t.Errorf("%#v", segments)
		}
	})
}

func TestGraphValidWithValidities(t *testing.T) {
	// m feeds a and b with 90/10 until 06-30 and with 70/30 from then on
	split := func(mutate func(g *Graph)) Graph {
		g := meterGraph()
		g.Edges[2].ValidUntil = day("06-30")
		g.Edges[3].ValidUntil = day("06-30")
		g.Edges = append(g.Edges,
			Edge{Id: "m>a-2", FromNodeId: "m", ToNodeId: "a", Weight: 70, ValidFrom: day("06-30")},
			Edge{Id: "m>b-2", FromNodeId: "m", ToNodeId: "b", Weight: 30, ValidFrom: day("06-30")},
		)
		if mutate != nil {
			mutate(&g)
		}
		return g
	}

	tests := []struct {
		name     string
		graph    Graph
		expected string //substring of the error, empty means valid
	}{
		{"two versions per edge with 90/10 then 70/30", split(nil), ""},
		{"the version that ends where the next starts does not overlap", split(func(g *Graph) { g.Edges[4].ValidFrom = day("06-30") }), ""},
		{"a gap: from 06-30 the node only has 70", split(func(g *Graph) { g.Edges = slices.Delete(g.Edges, 5, 6) }), "invalid at 2025-06-30T00:00:00Z: sum of outgoing edge weights for node m must be 100, got 70"},
		{"a node that has 90 only", split(func(g *Graph) {
			g.Edges = slices.Delete(g.Edges, 3, 4)
			g.Edges = slices.DeleteFunc(g.Edges, func(e Edge) bool { return e.Id == "m>b-2" })
		}), "got 90"},
		{"a gap before the first version names the segment before the first bound", split(func(g *Graph) { g.Edges[4].ValidFrom = day("07-15") }), "invalid at 2025-06-30T00:00:00Z"},
		{"versions of one edge overlap", split(func(g *Graph) { g.Edges[4].ValidFrom = day("06-01") }), "versions m>a and m>a-2 of the edge m -> a overlap"},
		{"a second weight over the old one for the same pair", split(func(g *Graph) { g.Edges[2].ValidUntil = nil }), "of the edge m -> a overlap"},
		{"valid_until equal to valid_from", split(func(g *Graph) { g.Edges[4].ValidUntil = day("06-30") }), "valid_until must be after valid_from in edge m>a-2"},
		{"valid_until before valid_from on a node", split(func(g *Graph) { g.Nodes[0].ValidFrom = day("06-30"); g.Nodes[0].ValidUntil = day("06-01") }), "valid_until must be after valid_from in node root"},
		{"edge ids are unique over all versions", split(func(g *Graph) { g.Edges[4].Id = "m>a" }), "duplicate edge id: m>a"},
		{"node ids are unique over all versions", split(func(g *Graph) {
			g.Nodes = append(g.Nodes, Node{Id: "m", ResourceType: "test", ValidFrom: day("07-01")})
		}), "duplicate node id: m"},
		{"an edge to a node that does not exist yet", split(func(g *Graph) { g.Nodes[1].ValidFrom = day("07-01") }), "unknown from node id: a"},
		{"two end nodes in one segment only", split(func(g *Graph) { g.Edges[1].ValidFrom = day("07-01") }), "exactly one node that has no outputs"},
		{"the end node may not be a device", split(func(g *Graph) { g.Nodes[0].ResourceType = GraphResourceTypeDevice }), "invalid before 2025-06-30T00:00:00Z: the end node may not have the resource_type"},
		{"a loop in one segment only", split(func(g *Graph) {
			g.Edges = append(g.Edges, Edge{Id: "root>m", FromNodeId: "root", ToNodeId: "m", Weight: 100, ValidFrom: day("08-01")})
		}), "invalid at 2025-08-01T00:00:00Z"},
		{"the same resource twice at once", split(func(g *Graph) {
			g.Nodes[1].ResourceId = "r1"
			g.Nodes[2].ResourceId = "r1"
		}), "duplicate resource id: r1"},
		{"the same resource twice one after the other", split(func(g *Graph) {
			g.Nodes[1].ResourceId = "r1"
			g.Nodes[1].ValidUntil = day("07-01")
			g.Nodes = append(g.Nodes, Node{Id: "a2", ResourceType: "test", ResourceId: "r1", ValidFrom: day("07-01")})
			g.Edges[0].ValidUntil = day("07-01")
			g.Edges = append(g.Edges, Edge{Id: "a2>root", FromNodeId: "a2", ToNodeId: "root", Weight: 100, ValidFrom: day("07-01")})
			g.Edges[2].ValidUntil = day("06-30")
			g.Edges[4].ValidUntil = day("07-01")
			g.Edges = append(g.Edges, Edge{Id: "m>a2", FromNodeId: "m", ToNodeId: "a2", Weight: 70, ValidFrom: day("07-01")})
		}), ""},
		{"attribute versions that do not overlap", split(func(g *Graph) {
			g.Attributes = []Attribute{{Key: "k", Value: "1", ValidUntil: day("06-01")}, {Key: "k", Value: "2", ValidFrom: day("06-01")}}
		}), ""},
		{"attribute versions that overlap", split(func(g *Graph) {
			g.Nodes[0].Attributes = []Attribute{{Key: "k", Value: "1", ValidUntil: day("06-02")}, {Key: "k", Value: "2", ValidFrom: day("06-01")}}
		}), "versions of attribute k of node root overlap"},
		{"an attribute version that overlaps an unbounded one", split(func(g *Graph) {
			g.Edges[0].Attributes = []Attribute{{Key: "k", Value: "1"}, {Key: "k", Value: "2", ValidFrom: day("06-01")}}
		}), "versions of attribute k of edge a>root overlap"},
		{"duplicate keys without any bound stay as legal as they were", split(func(g *Graph) {
			g.Attributes = []Attribute{{Key: "k", Value: "1"}, {Key: "k", Value: "2"}}
		}), ""},
		{"attribute valid_until before valid_from", split(func(g *Graph) {
			g.Attributes = []Attribute{{Key: "k", Value: "1", ValidFrom: day("06-02"), ValidUntil: day("06-01")}}
		}), "valid_until must be after valid_from in attribute k of graph g"},
		{"a graph that does not exist before its first node is valid there", func() Graph {
			g := meterGraph()
			for i := range g.Nodes {
				g.Nodes[i].ValidFrom = day("03-01")
			}
			for i := range g.Edges {
				g.Edges[i].ValidFrom = day("03-01")
			}
			return g
		}(), ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.graph.Valid()
			if test.expected == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q", test.expected)
			}
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("\n%v\n%v", err.Error(), test.expected)
			}
		})
	}

	t.Run("a graph without validities is checked as before", func(t *testing.T) {
		g := meterGraph()
		g.Edges[3].Weight = 20
		err := g.Valid()
		if err == nil || err.Error() != "sum of outgoing edge weights for node m must be 100, got 110" {
			t.Errorf("%v", err)
		}
	})

	t.Run("the segment error names an instant, not a segment index", func(t *testing.T) {
		g := split(func(g *Graph) { g.Edges = slices.Delete(g.Edges, 5, 6) })
		err := g.Valid()
		if err == nil || !strings.HasPrefix(err.Error(), "invalid at 2025-06-30T00:00:00Z: ") {
			t.Errorf("%v", err)
		}
	})
}
