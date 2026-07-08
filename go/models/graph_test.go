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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestGraphValidation(t *testing.T) {

	t.Run("tree", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{
					Id:           "o",
					ResourceType: "test",
				},
				{
					Id:           "uo1",
					ResourceType: "test",
				},
				{
					Id:           "standort",
					ResourceType: "test",
				},
				{
					Id:           "uo1uo1",
					ResourceType: "test",
				},
			},
			Edges: []Edge{
				{
					Id:         "uo1->o",
					FromNodeId: "uo1",
					ToNodeId:   "o",
					Weight:     100,
				},
				{
					Id:         "standort->o",
					FromNodeId: "standort",
					ToNodeId:   "o",
					Weight:     100,
				},
				{
					Id:         "uo1uo1->uo1",
					FromNodeId: "uo1uo1",
					ToNodeId:   "uo1",
					Weight:     100,
				},
			},
			IdProvider: idProvider,
		}
		err := graph.Valid()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("orga", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{
					Id:           "orga",
					ResourceType: "test",
				},
				{
					Id:           "finanzbuchhaltung",
					ResourceType: "test",
				},
				{
					Id:           "it",
					ResourceType: "test",
				},
				{
					Id:           "standort1",
					ResourceType: "test",
				},
				{
					Id:           "standort2",
					ResourceType: "test",
				},
				{
					Id:           "zähler1",
					ResourceType: "test",
				},
				{
					Id:           "zähler2",
					ResourceType: "test",
				},
			},
			Edges: []Edge{
				{
					Id:         "finanzbuchhaltung->orga",
					FromNodeId: "finanzbuchhaltung",
					ToNodeId:   "orga",
					Weight:     100,
				},
				{
					Id:         "it->orga",
					FromNodeId: "it",
					ToNodeId:   "orga",
					Weight:     100,
				},
				{
					Id:         "standort1->finanzbuchhaltung",
					FromNodeId: "standort1",
					ToNodeId:   "finanzbuchhaltung",
					Weight:     90,
				},
				{
					Id:         "standort1->it",
					FromNodeId: "standort1",
					ToNodeId:   "it",
					Weight:     10,
				},
				{
					Id:         "standort2->it",
					FromNodeId: "standort2",
					ToNodeId:   "it",
					Weight:     100,
				},
				{
					Id:         "zähler1->standort1",
					FromNodeId: "zähler1",
					ToNodeId:   "standort1",
					Weight:     100,
				},
				{
					Id:         "zähler2->standort2",
					FromNodeId: "zähler2",
					ToNodeId:   "standort2",
					Weight:     100,
				},
			},
			IdProvider: idProvider,
		}
		err := graph.Valid()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("o no loop", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{
					Id:           "o",
					ResourceType: "test",
				},
				{
					Id:           "uo1",
					ResourceType: "test",
				},
				{
					Id:           "it",
					ResourceType: "test",
				},
				{
					Id:           "uo1uo1",
					ResourceType: "test",
				},
				{
					Id:           "z1",
					ResourceType: "test",
				},
				{
					Id:           "z2",
					ResourceType: "test",
				},
				{
					Id:           "z3",
					ResourceType: "test",
				},
				{
					Id:           "z4",
					ResourceType: "test",
				},
				{
					Id:           "z5",
					ResourceType: "test",
				},
				{
					Id:           "z6",
					ResourceType: "test",
				},
			},
			Edges: []Edge{
				{
					Id:         "uo1->o",
					FromNodeId: "uo1",
					ToNodeId:   "o",
					Weight:     100,
				},
				{
					Id:         "it->o",
					FromNodeId: "it",
					ToNodeId:   "o",
					Weight:     90,
				},
				{
					Id:         "it->uo1",
					FromNodeId: "it",
					ToNodeId:   "uo1",
					Weight:     10,
				},
				{
					Id:         "uo1uo1->uo1",
					FromNodeId: "uo1uo1",
					ToNodeId:   "uo1",
					Weight:     40,
				},
				{
					Id:         "uo1uo1->it",
					FromNodeId: "uo1uo1",
					ToNodeId:   "it",
					Weight:     30,
				},
				{
					Id:         "uo1uo1->o",
					FromNodeId: "uo1uo1",
					ToNodeId:   "o",
					Weight:     30,
				},
				{
					Id:         "z1->it",
					FromNodeId: "z1",
					ToNodeId:   "it",
					Weight:     80,
				},
				{
					Id:         "z1->uo1uo1",
					FromNodeId: "z1",
					ToNodeId:   "uo1uo1",
					Weight:     20,
				},
				{
					Id:         "z2->uo1uo1",
					FromNodeId: "z2",
					ToNodeId:   "uo1uo1",
					Weight:     100,
				},
				{
					Id:         "z3->z2",
					FromNodeId: "z3",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z4->z2",
					FromNodeId: "z4",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z5->z2",
					FromNodeId: "z5",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z6->z2",
					FromNodeId: "z6",
					ToNodeId:   "z2",
					Weight:     100,
				},
			},
			IdProvider: idProvider,
		}
		err := graph.Valid()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("o loop flip it o", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{
					Id:           "o",
					ResourceType: "test",
				},
				{
					Id:           "uo1",
					ResourceType: "test",
				},
				{
					Id:           "it",
					ResourceType: "test",
				},
				{
					Id:           "uo1uo1",
					ResourceType: "test",
				},
				{
					Id:           "z1",
					ResourceType: "test",
				},
				{
					Id:           "z2",
					ResourceType: "test",
				},
				{
					Id:           "z3",
					ResourceType: "test",
				},
				{
					Id:           "z4",
					ResourceType: "test",
				},
				{
					Id:           "z5",
					ResourceType: "test",
				},
				{
					Id:           "z6",
					ResourceType: "test",
				},
			},
			Edges: []Edge{
				{
					Id:         "uo1->o",
					FromNodeId: "uo1",
					ToNodeId:   "o",
					Weight:     100,
				},
				{
					Id:         "o->it",
					FromNodeId: "o",
					ToNodeId:   "it",
					Weight:     100,
				},
				{
					Id:         "it->uo1",
					FromNodeId: "it",
					ToNodeId:   "uo1",
					Weight:     100,
				},
				{
					Id:         "uo1uo1->uo1",
					FromNodeId: "uo1uo1",
					ToNodeId:   "uo1",
					Weight:     40,
				},
				{
					Id:         "uo1uo1->it",
					FromNodeId: "uo1uo1",
					ToNodeId:   "it",
					Weight:     30,
				},
				{
					Id:         "uo1uo1->o",
					FromNodeId: "uo1uo1",
					ToNodeId:   "o",
					Weight:     30,
				},
				{
					Id:         "z1->it",
					FromNodeId: "z1",
					ToNodeId:   "it",
					Weight:     80,
				},
				{
					Id:         "z1->uo1uo1",
					FromNodeId: "z1",
					ToNodeId:   "uo1uo1",
					Weight:     20,
				},
				{
					Id:         "z2->uo1uo1",
					FromNodeId: "z2",
					ToNodeId:   "uo1uo1",
					Weight:     100,
				},
				{
					Id:         "z3->z2",
					FromNodeId: "z3",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z4->z2",
					FromNodeId: "z4",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z5->z2",
					FromNodeId: "z5",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z6->z2",
					FromNodeId: "z6",
					ToNodeId:   "z2",
					Weight:     100,
				},
			},
			IdProvider: idProvider,
		}
		err := graph.Valid()
		if !errors.Is(err, ErrGraphLoop) {
			t.Errorf("expected error '%v', got '%v'", ErrGraphLoop, err)
		}
	})

	t.Run("o loop add it uo1uo1", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{
					Id:           "o",
					ResourceType: "test",
				},
				{
					Id:           "uo1",
					ResourceType: "test",
				},
				{
					Id:           "it",
					ResourceType: "test",
				},
				{
					Id:           "uo1uo1",
					ResourceType: "test",
				},
				{
					Id:           "z1",
					ResourceType: "test",
				},
				{
					Id:           "z2",
					ResourceType: "test",
				},
				{
					Id:           "z3",
					ResourceType: "test",
				},
				{
					Id:           "z4",
					ResourceType: "test",
				},
				{
					Id:           "z5",
					ResourceType: "test",
				},
				{
					Id:           "z6",
					ResourceType: "test",
				},
			},
			Edges: []Edge{
				{
					Id:         "uo1->o",
					FromNodeId: "uo1",
					ToNodeId:   "o",
					Weight:     100,
				},
				{
					Id:         "it->o",
					FromNodeId: "it",
					ToNodeId:   "o",
					Weight:     80,
				},
				{
					Id:         "it->uo1",
					FromNodeId: "it",
					ToNodeId:   "uo1",
					Weight:     10,
				},
				{
					Id:         "uo1uo1->uo1",
					FromNodeId: "uo1uo1",
					ToNodeId:   "uo1",
					Weight:     40,
				},
				{
					Id:         "uo1uo1->it",
					FromNodeId: "uo1uo1",
					ToNodeId:   "it",
					Weight:     30,
				},
				{
					Id:         "it->uo1uo1",
					FromNodeId: "it",
					ToNodeId:   "uo1uo1",
					Weight:     10,
				},
				{
					Id:         "uo1uo1->o",
					FromNodeId: "uo1uo1",
					ToNodeId:   "o",
					Weight:     30,
				},
				{
					Id:         "z1->it",
					FromNodeId: "z1",
					ToNodeId:   "it",
					Weight:     80,
				},
				{
					Id:         "z1->uo1uo1",
					FromNodeId: "z1",
					ToNodeId:   "uo1uo1",
					Weight:     20,
				},
				{
					Id:         "z2->uo1uo1",
					FromNodeId: "z2",
					ToNodeId:   "uo1uo1",
					Weight:     100,
				},
				{
					Id:         "z3->z2",
					FromNodeId: "z3",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z4->z2",
					FromNodeId: "z4",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z5->z2",
					FromNodeId: "z5",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z6->z2",
					FromNodeId: "z6",
					ToNodeId:   "z2",
					Weight:     100,
				},
			},
			IdProvider: idProvider,
		}
		err := graph.Valid()
		if !errors.Is(err, ErrGraphLoop) {
			t.Errorf("expected error '%v', got '%v'", ErrGraphLoop, err)
		}
	})

	t.Run("o 0 weight", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{
					Id:           "o",
					ResourceType: "test",
				},
				{
					Id:           "uo1",
					ResourceType: "test",
				},
				{
					Id:           "it",
					ResourceType: "test",
				},
				{
					Id:           "uo1uo1",
					ResourceType: "test",
				},
				{
					Id:           "z1",
					ResourceType: "test",
				},
				{
					Id:           "z2",
					ResourceType: "test",
				},
				{
					Id:           "z3",
					ResourceType: "test",
				},
				{
					Id:           "z4",
					ResourceType: "test",
				},
				{
					Id:           "z5",
					ResourceType: "test",
				},
				{
					Id:           "z6",
					ResourceType: "test",
				},
			},
			Edges: []Edge{
				{
					Id:         "uo1->o",
					FromNodeId: "uo1",
					ToNodeId:   "o",
					Weight:     100,
				},
				{
					Id:         "it->o",
					FromNodeId: "it",
					ToNodeId:   "o",
					Weight:     100,
				},
				{
					Id:         "it->uo1",
					FromNodeId: "it",
					ToNodeId:   "uo1",
					Weight:     0,
				},
				{
					Id:         "uo1uo1->uo1",
					FromNodeId: "uo1uo1",
					ToNodeId:   "uo1",
					Weight:     40,
				},
				{
					Id:         "uo1uo1->it",
					FromNodeId: "uo1uo1",
					ToNodeId:   "it",
					Weight:     30,
				},
				{
					Id:         "uo1uo1->o",
					FromNodeId: "uo1uo1",
					ToNodeId:   "o",
					Weight:     30,
				},
				{
					Id:         "z1->it",
					FromNodeId: "z1",
					ToNodeId:   "it",
					Weight:     80,
				},
				{
					Id:         "z1->uo1uo1",
					FromNodeId: "z1",
					ToNodeId:   "uo1uo1",
					Weight:     20,
				},
				{
					Id:         "z2->uo1uo1",
					FromNodeId: "z2",
					ToNodeId:   "uo1uo1",
					Weight:     100,
				},
				{
					Id:         "z3->z2",
					FromNodeId: "z3",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z4->z2",
					FromNodeId: "z4",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z5->z2",
					FromNodeId: "z5",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z6->z2",
					FromNodeId: "z6",
					ToNodeId:   "z2",
					Weight:     100,
				},
			},
			IdProvider: idProvider,
		}
		err := graph.Valid()
		if err == nil {
			t.Error("expected invalid edge weight")
		}
	})

	t.Run("o missing output", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{
					Id:           "o",
					ResourceType: "test",
				},
				{
					Id:           "uo1",
					ResourceType: "test",
				},
				{
					Id:           "it",
					ResourceType: "test",
				},
				{
					Id:           "uo1uo1",
					ResourceType: "test",
				},
				{
					Id:           "z1",
					ResourceType: "test",
				},
				{
					Id:           "z2",
					ResourceType: "test",
				},
				{
					Id:           "z3",
					ResourceType: "test",
				},
				{
					Id:           "z4",
					ResourceType: "test",
				},
				{
					Id:           "z5",
					ResourceType: "test",
				},
				{
					Id:           "z6",
					ResourceType: "test",
				},
			},
			Edges: []Edge{
				{
					Id:         "uo1->o",
					FromNodeId: "uo1",
					ToNodeId:   "o",
					Weight:     100,
				},
				{
					Id:         "it->o",
					FromNodeId: "it",
					ToNodeId:   "o",
					Weight:     80,
				},
				{
					Id:         "it->uo1",
					FromNodeId: "it",
					ToNodeId:   "uo1",
					Weight:     10,
				},
				{
					Id:         "uo1uo1->uo1",
					FromNodeId: "uo1uo1",
					ToNodeId:   "uo1",
					Weight:     40,
				},
				{
					Id:         "uo1uo1->it",
					FromNodeId: "uo1uo1",
					ToNodeId:   "it",
					Weight:     30,
				},
				{
					Id:         "uo1uo1->o",
					FromNodeId: "uo1uo1",
					ToNodeId:   "o",
					Weight:     30,
				},
				{
					Id:         "z1->it",
					FromNodeId: "z1",
					ToNodeId:   "it",
					Weight:     80,
				},
				{
					Id:         "z1->uo1uo1",
					FromNodeId: "z1",
					ToNodeId:   "uo1uo1",
					Weight:     20,
				},
				{
					Id:         "z2->uo1uo1",
					FromNodeId: "z2",
					ToNodeId:   "uo1uo1",
					Weight:     100,
				},
				{
					Id:         "z3->z2",
					FromNodeId: "z3",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z4->z2",
					FromNodeId: "z4",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z5->z2",
					FromNodeId: "z5",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z6->z2",
					FromNodeId: "z6",
					ToNodeId:   "z2",
					Weight:     100,
				},
			},
			IdProvider: idProvider,
		}
		err := graph.Valid()
		if err == nil {
			t.Error("expected invalid sum of outgoing edge weights")
		}
	})

	t.Run("o unexpected output", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{
					Id:           "o",
					ResourceType: "test",
				},
				{
					Id:           "uo1",
					ResourceType: "test",
				},
				{
					Id:           "it",
					ResourceType: "test",
				},
				{
					Id:           "uo1uo1",
					ResourceType: "test",
				},
				{
					Id:           "z1",
					ResourceType: "test",
				},
				{
					Id:           "z2",
					ResourceType: "test",
				},
				{
					Id:           "z3",
					ResourceType: "test",
				},
				{
					Id:           "z4",
					ResourceType: "test",
				},
				{
					Id:           "z5",
					ResourceType: "test",
				},
				{
					Id:           "z6",
					ResourceType: "test",
				},
			},
			Edges: []Edge{
				{
					Id:         "uo1->o",
					FromNodeId: "uo1",
					ToNodeId:   "o",
					Weight:     100,
				},
				{
					Id:         "it->o",
					FromNodeId: "it",
					ToNodeId:   "o",
					Weight:     90,
				},
				{
					Id:         "it->uo1",
					FromNodeId: "it",
					ToNodeId:   "uo1",
					Weight:     20,
				},
				{
					Id:         "uo1uo1->uo1",
					FromNodeId: "uo1uo1",
					ToNodeId:   "uo1",
					Weight:     40,
				},
				{
					Id:         "uo1uo1->it",
					FromNodeId: "uo1uo1",
					ToNodeId:   "it",
					Weight:     30,
				},
				{
					Id:         "uo1uo1->o",
					FromNodeId: "uo1uo1",
					ToNodeId:   "o",
					Weight:     30,
				},
				{
					Id:         "z1->it",
					FromNodeId: "z1",
					ToNodeId:   "it",
					Weight:     80,
				},
				{
					Id:         "z1->uo1uo1",
					FromNodeId: "z1",
					ToNodeId:   "uo1uo1",
					Weight:     20,
				},
				{
					Id:         "z2->uo1uo1",
					FromNodeId: "z2",
					ToNodeId:   "uo1uo1",
					Weight:     100,
				},
				{
					Id:         "z3->z2",
					FromNodeId: "z3",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z4->z2",
					FromNodeId: "z4",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z5->z2",
					FromNodeId: "z5",
					ToNodeId:   "z2",
					Weight:     100,
				},
				{
					Id:         "z6->z2",
					FromNodeId: "z6",
					ToNodeId:   "z2",
					Weight:     100,
				},
			},
			IdProvider: idProvider,
		}
		err := graph.Valid()
		if err == nil {
			t.Error("expected invalid sum of outgoing edge weights")
		}
	})

	t.Run("to many end nodes", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{Id: "1"},
				{Id: "2"},
				{Id: "3"},
				{Id: "4"},
			},
			Edges: []Edge{
				{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 80},
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 20},
				{Id: "2->3", FromNodeId: "2", ToNodeId: "3", Weight: 50},
				{Id: "2->4", FromNodeId: "2", ToNodeId: "4", Weight: 50},
			},
			IdProvider: idProvider,
		}
		err := graph.Valid()
		t.Log(err)
		if err == nil || !strings.Contains(err.Error(), "the graph must have exactly one node that has no outputs") {
			t.Error("expect multiple end node error")
		}
	})
	t.Run("end node has resource_type device", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{Id: "1"},
				{Id: "2"},
				{Id: "3"},
				{Id: "4"},
				{Id: "5", ResourceType: GraphResourceTypeDevice},
			},
			Edges: []Edge{
				{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 80},
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 15},
				{Id: "2->3", FromNodeId: "2", ToNodeId: "3", Weight: 50},
				{Id: "2->4", FromNodeId: "2", ToNodeId: "4", Weight: 50},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
			},
			IdProvider: idProvider,
		}
		err := graph.Valid()
		t.Log(err)
		if err == nil || strings.Contains(err.Error(), "the graph must have exactly one node that has no outputs") {
			t.Error("expect multiple end node error")
		}
	})
}

func TestGraphNodeDelete(t *testing.T) {
	t.Run("with simple merge", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{Id: "1"},
				{Id: "2"},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []Edge{
				{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 80},
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 20},
				{Id: "2->3", FromNodeId: "2", ToNodeId: "3", Weight: 50},
				{Id: "2->4", FromNodeId: "2", ToNodeId: "4", Weight: 50},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
			},
			IdProvider: idProvider,
		}
		graph.DeleteNode("2")
		err := graph.Valid()
		if err != nil {
			t.Error(err)
		}
		if len(graph.Nodes) != 4 {
			t.Errorf("expected 4 nodes, got %v", len(graph.Nodes))
		}
		if len(graph.Edges) != 4 {
			t.Errorf("expected 4 edges, got %v", len(graph.Edges))
		}

		expectedResultGraph := Graph{
			Id: "a",
			Nodes: []Node{
				{Id: "1"},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []Edge{
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 60, Attributes: []Attribute{{Key: GraphEdgeAttrSystemChanged, Value: "true"}}},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
				{Id: "id-2", FromNodeId: "1", ToNodeId: "4", Weight: 40, Attributes: []Attribute{{Key: GraphEdgeAttrSystemChanged, Value: "true"}}},
			},
			IdProvider: idProvider,
		}

		graph = normalize(graph)
		expectedResultGraph = normalize(expectedResultGraph)
		if !reflect.DeepEqual(graph, expectedResultGraph) {
			t.Errorf("\ne=%#v\na=%#v\n", expectedResultGraph, graph)
		}
	})
	t.Run("with weight correction", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{Id: "1"},
				{Id: "2"},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []Edge{
				{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 85},
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 15},
				{Id: "2->3", FromNodeId: "2", ToNodeId: "3", Weight: 50},
				{Id: "2->4", FromNodeId: "2", ToNodeId: "4", Weight: 50},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
			},
			IdProvider: idProvider,
		}
		graph.DeleteNode("2")
		err := graph.Valid()
		if err != nil {
			t.Error(err)
		}
		if len(graph.Nodes) != 4 {
			t.Errorf("expected 4 nodes, got %v", len(graph.Nodes))
		}
		if len(graph.Edges) != 4 {
			t.Errorf("expected 4 edges, got %v", len(graph.Edges))
		}

		expectedResultGraph := Graph{
			Id: "a",
			Nodes: []Node{
				{Id: "1"},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []Edge{
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 58, Attributes: []Attribute{{Key: GraphEdgeAttrSystemChanged, Value: "true"}}},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
				{Id: "id-2", FromNodeId: "1", ToNodeId: "4", Weight: 42, Attributes: []Attribute{{Key: GraphEdgeAttrSystemChanged, Value: "true"}}},
			},
			IdProvider: idProvider,
		}

		graph = normalize(graph)
		expectedResultGraph = normalize(expectedResultGraph)
		if !reflect.DeepEqual(graph, expectedResultGraph) {
			t.Errorf("\ne=%#v\na=%#v\n", expectedResultGraph, graph)
		}
	})

	t.Run("multiple inputs and outputs", func(t *testing.T) {
		idProviderValue := 0
		idProvider := func() string {
			idProviderValue++
			return fmt.Sprintf("id-%v", idProviderValue)
		}
		graph := Graph{
			Id: "a",
			Nodes: []Node{
				{Id: "1"},
				{Id: "2"},
				{Id: "s2"},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []Edge{
				{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 85},
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 15},
				{Id: "s2->2", FromNodeId: "s2", ToNodeId: "2", Weight: 100},
				{Id: "2->3", FromNodeId: "2", ToNodeId: "3", Weight: 30},
				{Id: "2->4", FromNodeId: "2", ToNodeId: "4", Weight: 70},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
			},
			IdProvider: idProvider,
		}
		graph.DeleteNode("2")
		err := graph.Valid()
		if err != nil {
			t.Error(err)
		}

		expectedResultGraph := Graph{
			Id: "a",
			Nodes: []Node{
				{Id: "1"},
				{Id: "s2"},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []Edge{
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 41, Attributes: []Attribute{{Key: GraphEdgeAttrSystemChanged, Value: "true"}}},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
				{Id: "id-2", FromNodeId: "1", ToNodeId: "4", Weight: 59, Attributes: []Attribute{{Key: GraphEdgeAttrSystemChanged, Value: "true"}}},
				{Id: "id-3", FromNodeId: "s2", ToNodeId: "3", Weight: 30, Attributes: []Attribute{{Key: GraphEdgeAttrSystemChanged, Value: "true"}}},
				{Id: "id-4", FromNodeId: "s2", ToNodeId: "4", Weight: 70, Attributes: []Attribute{{Key: GraphEdgeAttrSystemChanged, Value: "true"}}},
			},
			IdProvider: idProvider,
		}

		graph = normalize(graph)
		expectedResultGraph = normalize(expectedResultGraph)
		if !reflect.DeepEqual(graph, expectedResultGraph) {
			t.Errorf("\ne=%#v\na=%#v\n", expectedResultGraph, graph)
		}
	})
}

func normalize[T any](e T) (r T) {
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &r)
	return r
}

//TODO: exactly one node with only input edges and no output edges
