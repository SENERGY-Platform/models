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
	"time"
)

type Thing struct {
	Id               string            `json:"id"`
	Name             string            `json:"name"`
	Attributes       []Attribute       `json:"attributes"`
	ContentVariables []ContentVariable `json:"content_variables"`
	OwnerId          string            `json:"owner_id"`
	DeviceLinks      []ThingDeviceLink `json:"device_links"`
}

// ThingDeviceLink states that a device measures a thing during [From, Until).
// Until == nil means the link is open.
type ThingDeviceLink struct {
	Id       string     `json:"id"`
	DeviceId string     `json:"device_id"`
	From     time.Time  `json:"from"`
	Until    *time.Time `json:"until"`
}

// Overlaps checks the time ranges only: both are half-open intervals [From, Until),
// a nil Until is unbounded. Ranges that touch (a.Until == b.From) do not overlap.
// DeviceId is not compared; callers filter by device beforehand.
func (this ThingDeviceLink) Overlaps(other ThingDeviceLink) bool {
	thisEndsBeforeOtherStarts := this.Until != nil && !this.Until.After(other.From)
	otherEndsBeforeThisStarts := other.Until != nil && !other.Until.After(this.From)
	return !thisEndsBeforeOtherStarts && !otherEndsBeforeThisStarts
}

// Valid does not validate ContentVariables.
func (this Thing) Valid() error {
	if this.Id == "" {
		return errors.New("thing id is required")
	}
	if this.Name == "" {
		return errors.New("thing name is required")
	}
	if this.OwnerId == "" {
		return errors.New("thing owner_id is required")
	}
	linkIds := map[string]bool{}
	for i, link := range this.DeviceLinks {
		if link.Id == "" {
			return errors.New("missing device link id")
		}
		if linkIds[link.Id] {
			return fmt.Errorf("duplicate device link id: %v", link.Id)
		}
		linkIds[link.Id] = true
		if link.DeviceId == "" {
			return fmt.Errorf("missing device_id in device link %v", link.Id)
		}
		if link.From.IsZero() {
			return fmt.Errorf("missing from in device link %v", link.Id)
		}
		if link.Until != nil && !link.Until.After(link.From) {
			return fmt.Errorf("until must be after from in device link %v", link.Id)
		}
		for _, other := range this.DeviceLinks[:i] {
			if other.DeviceId == link.DeviceId && other.Overlaps(link) {
				return fmt.Errorf("device links %v and %v for device %v overlap", other.Id, link.Id, link.DeviceId)
			}
		}
	}
	return nil
}
