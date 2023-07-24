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
	"go/ast"
	"go/token"
)

func GetEnums(f map[string]*ast.Package) []Enum {
	enums := []Enum{}
	for _, spec := range ListEnum(f) {
		e := Enum{Name: spec.Name.String()}
		e.Values = GetEnumValues(f, spec.Name.String())
		enums = append(enums, e)
	}
	return enums
}

func ListEnum(f map[string]*ast.Package) (result []*ast.TypeSpec) {
	for _, p := range f {
		for _, f := range p.Files {
			for _, decl := range f.Decls {
				gdecl, ok := decl.(*ast.GenDecl)
				if ok && gdecl.Tok == token.TYPE {
					for _, spec := range gdecl.Specs {
						tspec, ok := spec.(*ast.TypeSpec)
						if ok {
							subt, ok := tspec.Type.(*ast.Ident)
							if ok && subt.String() == "string" {
								result = append(result, tspec)
							}
						}
					}
				}
			}
		}
	}
	return result
}

func GetEnumValues(f map[string]*ast.Package, enumName string) (result []string) {
	for _, p := range f {
		for _, f := range p.Files {
			for _, decl := range f.Decls {
				gdecl, ok := decl.(*ast.GenDecl)
				if ok && gdecl.Tok == token.CONST {
					for _, spec := range gdecl.Specs {
						vspec, ok := spec.(*ast.ValueSpec)
						if ok && vspec.Type != nil {
							vspect, ok := vspec.Type.(*ast.Ident)
							if ok && vspect.Name == enumName {
								for _, v := range vspec.Values {
									lit, ok := v.(*ast.BasicLit)
									if ok {
										result = append(result, lit.Value)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return result
}
