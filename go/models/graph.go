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

	"github.com/google/uuid"
)

const GraphEdgeAttrSystemChanged = "system_changed"

const GraphResourceTypeDevice GraphResourceType = "device"

var ErrGraphLoop = fmt.Errorf("graph contains cycle/loop")

var DefaultGraphIdProvider = func() string {
	return uuid.NewString()
}

type Graph struct {
	Id         string        `json:"id"`
	Attributes []Attribute   `json:"attributes"`
	Nodes      []Node        `json:"nodes"`
	Edges      []Edge        `json:"edges"`
	IdProvider func() string `json:"-"`
}

type Node struct {
	Id           string            `json:"id"`
	Attributes   []Attribute       `json:"attributes"`
	ResourceType GraphResourceType `json:"resource_type"`
	ResourceId   string            `json:"resource_id"`
}

type Edge struct {
	Id         string      `json:"id"`
	FromNodeId string      `json:"from_node_id"`
	ToNodeId   string      `json:"to_node_id"`
	Weight     int         `json:"weight"`
	Attributes []Attribute `json:"attributes"`
}

type GraphResourceType = string

func (this *Graph) createId() string {
	if this.IdProvider != nil {
		return this.IdProvider()
	}
	return DefaultGraphIdProvider()
}

func (this *Graph) Valid() error {
	nodeIds := map[string]bool{}
	resourceIds := map[string]bool{}
	outgoingWeightSums := map[string]int{}

	for _, node := range this.Nodes {
		if node.Id == "" {
			return errors.New("missing node id")
		}
		if nodeIds[node.Id] {
			return fmt.Errorf("duplicate node id: %v", node.Id)
		}
		nodeIds[node.Id] = true

		if node.ResourceId != "" {
			if resourceIds[node.ResourceId] {
				return fmt.Errorf("duplicate resource id: %v", node.ResourceId)
			}
			resourceIds[node.ResourceId] = true
		}

		outgoingWeightSums[node.Id] = 0
	}

	edgeIds := map[string]bool{}
	edgeDirections := map[string]map[string]string{} //[from][to] = id

	for _, edge := range this.Edges {
		if edge.Id == "" {
			return fmt.Errorf("missing edge id between %v and %v", edge.FromNodeId, edge.ToNodeId)
		}
		if edgeIds[edge.Id] {
			return fmt.Errorf("duplicate edge id: %v", edge.Id)
		}
		if edge.Weight <= 0 || edge.Weight > 100 {
			return fmt.Errorf("invalid edge weight: %v in %v", edge.Weight, edge.Id)
		}
		edgeIds[edge.Id] = true

		if !nodeIds[edge.FromNodeId] {
			return fmt.Errorf("unknown from node id: %v", edge.FromNodeId)
		}
		if !nodeIds[edge.ToNodeId] {
			return fmt.Errorf("unknown to node id: %v", edge.ToNodeId)
		}

		if edgeDirections[edge.FromNodeId] != nil && edgeDirections[edge.FromNodeId][edge.ToNodeId] != "" && edgeDirections[edge.FromNodeId][edge.ToNodeId] != edge.Id {
			return fmt.Errorf("duplicate edge between nodes: %v -> %v (%v vs %v)", edge.FromNodeId, edge.ToNodeId, edgeDirections[edge.FromNodeId][edge.ToNodeId], edge.Id)
		}
		if edgeDirections[edge.FromNodeId] == nil {
			edgeDirections[edge.FromNodeId] = map[string]string{}
		}
		edgeDirections[edge.FromNodeId][edge.ToNodeId] = edge.Id

		outgoingWeightSums[edge.FromNodeId] += edge.Weight
	}

	if this.ContainsLoop() {
		return ErrGraphLoop
	}

	endNodes := []Node{}
	for _, node := range this.Nodes {
		if len(edgeDirections[node.Id]) == 0 {
			endNodes = append(endNodes, node)
		}
	}
	if len(endNodes) == 0 {
		return errors.New("the graph must have exactly one node that has no outputs, none where found")
	}
	if len(endNodes) > 1 {
		ids := []string{}
		for _, node := range endNodes {
			ids = append(ids, node.Id)
		}
		return fmt.Errorf("the graph must have exactly one node that has no outputs, %v where found (%+v)", len(ids), ids)
	}

	if endNodes[0].ResourceType == GraphResourceTypeDevice {
		return fmt.Errorf("the end node may not have the resource_type '%v'", GraphResourceTypeDevice)
	}

	for nodeId, sum := range outgoingWeightSums {
		if sum != 0 && sum != 100 {
			return fmt.Errorf("sum of outgoing edge weights for node %v must be 100, got %v", nodeId, sum)
		}
	}

	return nil
}

func (this *Graph) ContainsLoop() bool {
	edges := map[string][]string{}
	for _, edge := range this.Edges {
		edges[edge.FromNodeId] = append(edges[edge.FromNodeId], edge.ToNodeId)
	}

	knownNodeWithoutLoop := map[string]bool{}
	visitedNodesInCurrentNodesCheck := map[string]bool{}

	var check func(nodeId string) bool
	check = func(nodeId string) bool {
		if knownNodeWithoutLoop[nodeId] {
			return false
		}
		if visitedNodesInCurrentNodesCheck[nodeId] {
			return true
		}
		visitedNodesInCurrentNodesCheck[nodeId] = true
		for _, to := range edges[nodeId] {
			if check(to) {
				return true
			}
		}

		knownNodeWithoutLoop[nodeId] = true
		return false
	}

	for _, node := range this.Nodes {
		visitedNodesInCurrentNodesCheck = map[string]bool{}
		if check(node.Id) {
			return true
		}
	}

	return false
}

func (this *Graph) DeleteNode(nodeId string) {
	this.Nodes = slices.DeleteFunc(this.Nodes, func(node Node) bool {
		return node.Id == nodeId
	})
	outgoingEdges := []Edge{}
	incomingEdges := []Edge{}
	for _, edge := range this.Edges {
		if edge.FromNodeId == nodeId {
			outgoingEdges = append(outgoingEdges, edge)
		}
		if edge.ToNodeId == nodeId {
			incomingEdges = append(incomingEdges, edge)
		}
	}
	for _, edge := range incomingEdges {
		this.rerouteEdge(edge, outgoingEdges)
	}
	for _, edge := range outgoingEdges {
		this.Edges = slices.DeleteFunc(this.Edges, func(e Edge) bool {
			return e.Id == edge.Id
		})
	}
}

func (this *Graph) rerouteEdge(edge Edge, copyDestinationFrom []Edge) {
	newEdges := []Edge{}
	for _, destination := range copyDestinationFrom {
		newWeigt := float64(edge.Weight) * (float64(destination.Weight) / 100) //assumption: sum of all weights of copyDestinationFrom is 100 (ref validation outgoingWeightSums)
		newEdge := Edge{
			Id:         this.createId(),
			FromNodeId: edge.FromNodeId,
			ToNodeId:   destination.ToNodeId,
			Weight:     int(newWeigt),
			Attributes: []Attribute{
				{
					Key:   GraphEdgeAttrSystemChanged,
					Value: "true",
				},
			},
		}
		this.Edges = append(this.Edges, newEdge)
		newEdges = append(newEdges, newEdge)
	}
	this.Edges = slices.DeleteFunc(this.Edges, func(e Edge) bool {
		return e.Id == edge.Id
	})
	for _, newEdge := range newEdges {
		this.mergeDuplicateEdge(newEdge)
	}
	this.ensureValidEdgeWeights() //in case of rounding problems
}

func (this *Graph) mergeDuplicateEdge(edge Edge) {
	for i, e := range this.Edges {
		if e.FromNodeId == edge.FromNodeId && e.ToNodeId == edge.ToNodeId && e.Id != edge.Id {
			e.Weight += edge.Weight
			e.Attributes = []Attribute{{Key: GraphEdgeAttrSystemChanged, Value: "true"}}
			this.Edges[i] = e
			this.Edges = slices.DeleteFunc(this.Edges, func(e2 Edge) bool {
				return e2.Id == edge.Id
			})
			return
		}
	}
}

func (this *Graph) ensureValidEdgeWeights() {
	outgoingWeightSums := map[string]int{}
	for _, edge := range this.Edges {
		outgoingWeightSums[edge.FromNodeId] += edge.Weight
	}
	for nodeId, sum := range outgoingWeightSums {
		if sum != 100 && sum != 0 {
			diff := 100 - sum
			for i, edge := range this.Edges {
				if edge.FromNodeId == nodeId {
					edge.Weight += diff
					attrIndex := slices.IndexFunc(edge.Attributes, func(a Attribute) bool {
						return a.Key == GraphEdgeAttrSystemChanged
					})
					if attrIndex == -1 {
						edge.Attributes = append(edge.Attributes, Attribute{Key: GraphEdgeAttrSystemChanged, Value: "true"})
					} else {
						edge.Attributes[i].Value = "true"
					}
					this.Edges[i] = edge
					break
				}
			}
		}
	}
}
