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

func GetStructs(f map[string]*ast.Package) []Struct {
	structs := []Struct{}
	for _, spec := range ListStructTypeSpecs(f) {
		structs = append(structs, Struct{Name: spec.Name.String(), Fields: GetStructFields(spec)})
	}
	return structs
}

func GetStructFields(spec *ast.TypeSpec) (result []Field) {
	stype, ok := spec.Type.(*ast.StructType)
	if !ok {
		return nil
	}
	if stype.Fields == nil {
		return nil
	}
	for _, field := range stype.Fields.List {
		fieldTypeName, elementTypeName, many := GetTypeInfo(field.Type)
		if field.Names == nil {
			//embedded types
			ft, ok := field.Type.(*ast.Ident)
			if ok {
				ftspec, ok := ft.Obj.Decl.(*ast.TypeSpec)
				if ok {
					result = append(result, GetStructFields(ftspec)...)
				}
			}
		}
		for _, name := range field.Names {
			result = append(result, Field{
				Name:        name.String(),
				Type:        fieldTypeName,
				ElementType: elementTypeName,
				Many:        many,
			})
		}
	}
	return result
}

func GetTypeName(t ast.Expr) string {
	result, _, _ := GetTypeInfo(t)
	return result
}

func GetTypeInfo(t ast.Expr) (fullName string, elementName string, many bool) {
	switch ft := t.(type) {
	case *ast.Ident:
		if ft.Obj != nil {
			ot, ok := ft.Obj.Type.(*ast.Ident)
			if ok && GetTypeName(ot) == "string" {
				return "string", "string", false
			}
		}
		elementName = ft.String()
		return elementName, elementName, false
	case *ast.ArrayType:
		fullSub, elemName, _ := GetTypeInfo(ft.Elt)
		return "[]" + fullSub, elemName, true
	case *ast.MapType:
		valueTypeName, elemName, _ := GetTypeInfo(ft.Value)
		keyTypeName, _, _ := GetTypeInfo(ft.Key)
		return "map[" + keyTypeName + "]" + valueTypeName, elemName, true
	case *ast.InterfaceType:
		if len(ft.Methods.List) == 0 {
			return "any", "any", false
		} else {
			return "<Interface>", "<Interface>", false
		}
	default:
		return "", "", false
	}
}

func ListStructTypeSpecs(f map[string]*ast.Package) (result []*ast.TypeSpec) {
	for _, p := range f {
		for _, f := range p.Files {
			for _, decl := range f.Decls {
				gdecl, ok := decl.(*ast.GenDecl)
				if ok && gdecl.Tok == token.TYPE {
					for _, spec := range gdecl.Specs {
						tspec, ok := spec.(*ast.TypeSpec)
						if ok {
							_, ok := tspec.Type.(*ast.StructType)
							if ok {
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
