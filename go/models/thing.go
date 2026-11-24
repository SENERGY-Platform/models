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

const (
	ThingKindConsumer = "consumer"
	ThingKindMeter    = "meter"
)

type Thing struct {
	Id               string            `json:"id"`
	Kind             string            `json:"kind,omitempty"` // empty means ThingKindConsumer
	Name             string            `json:"name"`
	Attributes       []Attribute       `json:"attributes"`
	ContentVariables []ContentVariable `json:"content_variables"`
	OwnerId          string            `json:"owner_id"`
	DeviceLinks      []ThingDeviceLink `json:"device_links"`
}

// ThingDeviceLink states that a device measures a thing during [From, Until).
// Until == nil means the link is open.
// Readings and Factor are only allowed on the links of a meter thing.
type ThingDeviceLink struct {
	Id       string                   `json:"id"`
	DeviceId string                   `json:"device_id"`
	From     time.Time                `json:"from"`
	Until    *time.Time               `json:"until"`
	Readings []ThingDeviceLinkReading `json:"readings,omitempty"`
	Factor   *float64                 `json:"factor,omitempty"` // transformer ratio, nil means 1
}

// ThingDeviceLinkReading is the counter reading of one column of the linked device
// at installation (FromValue) and removal (UntilValue, nil while the link is open).
// Column is the dot path as in the timescale columns.
// Both values are in the unit the column is delivered in after the platform's characteristic conversion,
// i.e. kWh for energy and m³ for volume, the unit read off the meter display, not the raw device unit (e.g. Wh).
type ThingDeviceLinkReading struct {
	ServiceId  string   `json:"service_id"`
	Column     string   `json:"column"`
	FromValue  float64  `json:"from_value"`
	UntilValue *float64 `json:"until_value,omitempty"`
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
	if this.Kind != "" && this.Kind != ThingKindConsumer && this.Kind != ThingKindMeter {
		return fmt.Errorf("unknown thing kind: %v", this.Kind)
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
		if err := this.validLinkMeterFields(link); err != nil {
			return err
		}
		for _, other := range this.DeviceLinks[:i] {
			if other.DeviceId == link.DeviceId && other.Overlaps(link) {
				return fmt.Errorf("device links %v and %v for device %v overlap", other.Id, link.Id, link.DeviceId)
			}
		}
	}
	return nil
}

// validLinkMeterFields checks Readings and Factor of a link: both are only allowed for meters.
func (this Thing) validLinkMeterFields(link ThingDeviceLink) error {
	if link.Factor != nil && !(*link.Factor > 0) {
		return fmt.Errorf("factor must be greater than 0 in device link %v", link.Id)
	}
	if this.Kind != ThingKindMeter && (len(link.Readings) > 0 || link.Factor != nil) {
		return fmt.Errorf("readings and factor in device link %v are only allowed for things of kind %v", link.Id, ThingKindMeter)
	}
	type readingKey struct{ serviceId, column string }
	keys := map[readingKey]bool{}
	for _, reading := range link.Readings {
		if reading.ServiceId == "" {
			return fmt.Errorf("missing service_id in reading of device link %v", link.Id)
		}
		if reading.Column == "" {
			return fmt.Errorf("missing column in reading of device link %v", link.Id)
		}
		key := readingKey{reading.ServiceId, reading.Column}
		if keys[key] {
			return fmt.Errorf("duplicate reading %v %v in device link %v", reading.ServiceId, reading.Column, link.Id)
		}
		keys[key] = true
		if reading.UntilValue != nil && link.Until == nil {
			return fmt.Errorf("until_value in reading %v %v needs until in device link %v", reading.ServiceId, reading.Column, link.Id)
		}
	}
	return nil
}
