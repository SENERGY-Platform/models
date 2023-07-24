/*
 * Copyright (c) 2023 InfAI (CC SES)
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

package pkg

import "fmt"

func SetRelations(structs []Struct, enums []Enum) (result []Struct) {
	result = SetCompositions(structs, enums)
	result = SetInferredRelations(result)
	return result
}

func SetCompositions(structs []Struct, enums []Enum) (result []Struct) {
	knownStructs := map[string]bool{}
	for _, s := range structs {
		knownStructs[s.Name] = true
	}
	knownEnums := map[string]bool{}
	for _, e := range enums {
		knownEnums[e.Name] = true
	}
	indexFieldTypeToStruct := map[string][]string{}
	for _, s := range structs {
		for _, f := range s.Fields {
			indexFieldTypeToStruct[f.ElementType] = append(indexFieldTypeToStruct[f.ElementType], s.Name)
		}
	}
	for _, s := range structs {
		for _, f := range s.Fields {
			if knownStructs[f.ElementType] {
				s.Compositions = append(s.Compositions, getCompositeRel(f, indexFieldTypeToStruct))
			}
			if knownEnums[f.ElementType] {
				s.Compositions = append(s.Compositions, getEnumRel(f))
			}
		}
		s.Compositions = mergeDirectedRelations(s.Compositions)
		result = append(result, s)
	}
	return result
}

func getCompositeRel(f Field, indexFieldTypeToStruct map[string][]string) (result Rel) {
	result.StructName = f.ElementType
	result.TargetCardinality = f.Card
	result.SourceCardinality = ExactOne
	if len(indexFieldTypeToStruct[f.ElementType]) > 1 {
		result.SourceCardinality = MaxOne
	}
	return result
}

func getEnumRel(f Field) (result Rel) {
	result.StructName = f.ElementType
	result.TargetCardinality = f.Card
	result.SourceCardinality = Many
	return result
}

func mergeDirectedRelations(composition []Rel) (result []Rel) {
	set := map[string][]Rel{}
	for _, r := range composition {
		set[r.StructName] = append(set[r.StructName], r)
	}
	for _, s := range set {
		element := Rel{}
		for _, r := range s {
			element.StructName = r.StructName
			element.SourceCardinality = r.SourceCardinality //source cardinality should be the same for every element

			//we have only 2 possible states: Card and MaxOne
			//and many overwrites MaxOne
			if element.TargetCardinality == "" {
				element.TargetCardinality = r.TargetCardinality
			}
			if r.TargetCardinality == Many {
				element.TargetCardinality = Many
			}
		}
		result = append(result, element)
	}
	return result
}

func SetInferredRelations(structs []Struct) (result []Struct) {
	index := map[string][]Struct{}
	for _, s := range structs {
		for _, f := range s.Fields {
			index[s.Name+f.Name] = append(index[s.Name+f.Name], s)
			index[s.Name+f.Name+"s"] = append(index[s.Name+f.Name+"s"], s) //for list fields
		}
	}
	for _, s := range structs {
		for _, f := range s.Fields {
			if list, found := index[f.Name]; found {
				if len(list) > 1 {
					fmt.Printf("WARNING: ambiguous inferred relation for %v, %v => %#v", s.Name, f.Name, list)
				}
				card := f.Card
				if card == ExactOne {
					//inferred relations (strings) may be empty
					card = MaxOne
				}
				s.InferredRelations = append(s.InferredRelations, Rel{
					StructName:        list[0].Name,
					SourceCardinality: Many,
					TargetCardinality: card,
				})
			}
		}
		s.InferredRelations = mergeDirectedRelations(s.InferredRelations)
		result = append(result, s)
	}
	return result
}
