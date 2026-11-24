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

import "time"

type Attribute struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Origin string `json:"origin"`

	// ValidFrom and ValidUntil are only interpreted on the attributes of a graph, its nodes and its edges.
	// They bound the attribute to [ValidFrom, ValidUntil), nil means unbounded on that side.
	// Every other owner of attributes (devices, device types, things, ...) ignores them and does not write them.
	ValidFrom  *time.Time `json:"valid_from,omitempty" bson:",omitempty"`
	ValidUntil *time.Time `json:"valid_until,omitempty" bson:",omitempty"`
}
