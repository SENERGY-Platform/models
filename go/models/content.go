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

type Content struct {
	Id                string          `json:"id"`
	ContentVariable   ContentVariable `json:"content_variable"`
	Serialization     Serialization   `json:"serialization"`
	ProtocolSegmentId string          `json:"protocol_segment_id"`
}

type Serialization string

const (
	XML       Serialization = "xml"
	JSON      Serialization = "json"
	PlainText Serialization = "plain-text"
)

func (this Serialization) Valid() bool {
	switch this {
	case XML:
		return true
	case JSON:
		return true
	case PlainText:
		return true
	default:
		return false
	}
}
