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

import (
	"fmt"
	"strings"
)

func Generate(structs []Struct, enums []Enum) (result string) {
	header := "@startuml\n!pragma layout elk\n\n"
	footer := "\n\n@enduml"

	enumsUml := EnumsUml(enums)
	structUml := StructsUml(structs)

	return header + strings.Join(enumsUml, "\n\n") + "\n\n" + strings.Join(structUml, "\n") + footer
}

func EnumsUml(enums []Enum) (result []string) {
	for _, e := range enums {
		result = append(result, EnumUml(e))
	}
	return
}

func EnumUml(enum Enum) string {
	return fmt.Sprintf(`enum %v {
		%v
	}`, enum.Name, strings.Join(enum.Values, "\n"))
}

func StructsUml(structs []Struct) (result []string) {
	for _, s := range structs {
		result = append(result, StructUml(s))
	}
	return
}

func StructUml(s Struct) string {
	return fmt.Sprintf(`
class %v {
%v
}
%v
%v`,
		s.Name,
		strings.Join(FieldsUml(s.Fields), "\n"),
		strings.Join(CompositionsUml(s.Name, s.Compositions), "\n"),
		strings.Join(InferredRelationsUml(s.Name, s.InferredRelations), "\n"),
	)
}

func CompositionsUml(source string, compositions []Rel) (result []string) {
	for _, c := range compositions {
		result = append(result, CompositionUml(source, c))
	}
	return
}

func CompositionUml(source string, rel Rel) string {
	return fmt.Sprintf(`%v "%v" o- "%v" %v`, source, rel.SourceCardinality, rel.TargetCardinality, rel.StructName)
}

func InferredRelationsUml(source string, relations []Rel) (result []string) {
	for _, rel := range relations {
		result = append(result, InferredRelationUml(source, rel))
	}
	return
}

func InferredRelationUml(source string, rel Rel) string {
	return fmt.Sprintf(`%v "%v" - "%v" %v`, source, rel.SourceCardinality, rel.TargetCardinality, rel.StructName)
}

func FieldsUml(fields []Field) (result []string) {
	for _, f := range fields {
		result = append(result, FieldUml(f))
	}
	return
}

func FieldUml(field Field) string {
	return fmt.Sprintf("%v %v", field.Name, field.Type)
}
