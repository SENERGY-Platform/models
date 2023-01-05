/*
 * Copyright 2022 InfAI (CC SES)
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

type Concept struct {
	Id                   string               `json:"id"`
	Name                 string               `json:"name"`
	CharacteristicIds    []string             `json:"characteristic_ids"`
	BaseCharacteristicId string               `json:"base_characteristic_id"`
	Conversions          []ConverterExtension `json:"conversions"`
}

type ConverterExtension struct {
	From            string `json:"from"`
	To              string `json:"to"`
	Distance        int64  `json:"distance"`
	Formula         string `json:"formula"`
	PlaceholderName string `json:"placeholder_name"`
}

type ConceptWithCharacteristics struct {
	Id                   string               `json:"id"`
	Name                 string               `json:"name"`
	BaseCharacteristicId string               `json:"base_characteristic_id"`
	Characteristics      []Characteristic     `json:"characteristics"`
	Conversions          []ConverterExtension `json:"conversions"`
}
