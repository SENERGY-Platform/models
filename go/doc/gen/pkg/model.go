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

type Enum struct {
	Name   string
	Values []string
}

type Struct struct {
	Name   string
	Fields []Field

	//relations
	Compositions      []Rel
	InferredRelations []Rel
}

type Rel struct {
	StructName        string
	SourceCardinality Cardinality
	TargetCardinality Cardinality
}

type Cardinality = string

const MaxOne Cardinality = "0..1"
const ExactOne Cardinality = "1"
const AtLeastOne Cardinality = "1..*"
const Many Cardinality = "0..*"

type Field struct {
	Name        string
	Type        string
	ElementType string //if type is an array or map
	Card        Cardinality
}
