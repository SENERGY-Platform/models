/*
 * Copyright 2024 InfAI (CC SES)
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

type Permissions struct {
	Read         bool `json:"read"`
	Write        bool `json:"write"`
	Execute      bool `json:"execute"`
	Administrate bool `json:"administrate"`
}

type PermissionFlag rune

const UnsetPermissionFlag PermissionFlag = 0 //to identify where default values must be used
const Read PermissionFlag = 'r'              //user may read the resource (metadata)  (e.g. read device name)
const Write PermissionFlag = 'w'             //user may write the resource (metadata)(e.g. rename device)
const Administrate PermissionFlag = 'a'      // user may delete resource; user may change resource rights (e.g. delete device)
const Execute PermissionFlag = 'x'           //user may use the resource (e.g. cmd to device; read device data; read database)
