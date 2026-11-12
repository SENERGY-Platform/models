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
	"strings"
	"testing"
	"time"
)

func tm(day int) time.Time {
	return time.Date(2026, 11, day, 0, 0, 0, 0, time.UTC)
}

func tp(day int) *time.Time {
	t := tm(day)
	return &t
}

func TestThingDeviceLinkOverlaps(t *testing.T) {
	tests := []struct {
		name string
		a, b ThingDeviceLink
		want bool
	}{
		{"overlapping", ThingDeviceLink{From: tm(1), Until: tp(10)}, ThingDeviceLink{From: tm(5), Until: tp(15)}, true},
		{"touching a before b", ThingDeviceLink{From: tm(1), Until: tp(5)}, ThingDeviceLink{From: tm(5), Until: tp(10)}, false},
		{"touching b before a", ThingDeviceLink{From: tm(5), Until: tp(10)}, ThingDeviceLink{From: tm(1), Until: tp(5)}, false},
		{"both open", ThingDeviceLink{From: tm(1)}, ThingDeviceLink{From: tm(5)}, true},
		{"a open", ThingDeviceLink{From: tm(1)}, ThingDeviceLink{From: tm(5), Until: tp(10)}, true},
		{"b open", ThingDeviceLink{From: tm(5), Until: tp(10)}, ThingDeviceLink{From: tm(1)}, true},
		{"open after closed", ThingDeviceLink{From: tm(1), Until: tp(5)}, ThingDeviceLink{From: tm(7)}, false},
		{"contained", ThingDeviceLink{From: tm(1), Until: tp(20)}, ThingDeviceLink{From: tm(5), Until: tp(10)}, true},
		{"contained reversed", ThingDeviceLink{From: tm(5), Until: tp(10)}, ThingDeviceLink{From: tm(1), Until: tp(20)}, true},
		{"disjoint", ThingDeviceLink{From: tm(1), Until: tp(3)}, ThingDeviceLink{From: tm(5), Until: tp(10)}, false},
		{"disjoint reversed", ThingDeviceLink{From: tm(5), Until: tp(10)}, ThingDeviceLink{From: tm(1), Until: tp(3)}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.Overlaps(tc.b); got != tc.want {
				t.Errorf("a.Overlaps(b) = %v, want %v", got, tc.want)
			}
			if got := tc.b.Overlaps(tc.a); got != tc.want {
				t.Errorf("b.Overlaps(a) = %v, want %v", got, tc.want)
			}
		})
	}
}

func validThing() Thing {
	return Thing{
		Id:      "urn:infai:ses:thing:1",
		Name:    "oven",
		OwnerId: "owner",
		DeviceLinks: []ThingDeviceLink{
			{Id: "l1", DeviceId: "d1", From: tm(1), Until: tp(5)},
		},
	}
}

func TestThingValid(t *testing.T) {
	tests := []struct {
		name    string
		mod     func(*Thing)
		wantErr string
	}{
		{"valid", func(th *Thing) {}, ""},
		{"valid without links", func(th *Thing) { th.DeviceLinks = nil }, ""},
		{"missing id", func(th *Thing) { th.Id = "" }, "thing id"},
		{"missing name", func(th *Thing) { th.Name = "" }, "thing name"},
		{"missing owner", func(th *Thing) { th.OwnerId = "" }, "owner_id"},
		{"missing link id", func(th *Thing) { th.DeviceLinks[0].Id = "" }, "missing device link id"},
		{"missing device id", func(th *Thing) { th.DeviceLinks[0].DeviceId = "" }, "missing device_id"},
		{"zero from", func(th *Thing) { th.DeviceLinks[0].From = time.Time{} }, "missing from"},
		{"until equals from", func(th *Thing) { th.DeviceLinks[0].Until = tp(1) }, "until must be after from"},
		{"until before from", func(th *Thing) { th.DeviceLinks[0].Until = tp(0) }, "until must be after from"},
		{"duplicate link id", func(th *Thing) {
			th.DeviceLinks = append(th.DeviceLinks, ThingDeviceLink{Id: "l1", DeviceId: "d2", From: tm(1)})
		}, "duplicate device link id"},
		{"overlap same device", func(th *Thing) {
			th.DeviceLinks = append(th.DeviceLinks, ThingDeviceLink{Id: "l2", DeviceId: "d1", From: tm(3)})
		}, "overlap"},
		{"overlap different devices allowed", func(th *Thing) {
			th.DeviceLinks = append(th.DeviceLinks, ThingDeviceLink{Id: "l2", DeviceId: "d2", From: tm(3)})
		}, ""},
		{"touching same device allowed", func(th *Thing) {
			th.DeviceLinks = append(th.DeviceLinks, ThingDeviceLink{Id: "l2", DeviceId: "d1", From: tm(5)})
		}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			th := validThing()
			tc.mod(&th)
			err := th.Valid()
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestThingGenerateId(t *testing.T) {
	t.Run("prefix", func(t *testing.T) {
		th := Thing{}
		th.GenerateId()
		if !strings.HasPrefix(th.Id, THING_PREFIX) || len(th.Id) <= len(THING_PREFIX) {
			t.Errorf("unexpected id %q", th.Id)
		}
	})
	t.Run("existing id kept", func(t *testing.T) {
		th := Thing{Id: "keep"}
		th.GenerateId()
		if th.Id != "keep" {
			t.Errorf("id changed to %q", th.Id)
		}
	})
	t.Run("content variable ids", func(t *testing.T) {
		th := Thing{ContentVariables: []ContentVariable{
			{Name: "a", SubContentVariables: []ContentVariable{
				{Name: "b", SubContentVariables: []ContentVariable{{Name: "c"}}},
				{Id: "keep-sub", Name: "d"},
			}},
			{Id: "keep-top", Name: "e"},
		}}
		th.GenerateId()
		cv := th.ContentVariables
		for _, id := range []string{cv[0].Id, cv[0].SubContentVariables[0].Id, cv[0].SubContentVariables[0].SubContentVariables[0].Id} {
			if !strings.HasPrefix(id, URN_PREFIX+"content-variable:") {
				t.Errorf("unexpected content variable id %q", id)
			}
		}
		if cv[0].SubContentVariables[1].Id != "keep-sub" || cv[1].Id != "keep-top" {
			t.Errorf("existing content variable ids changed: %+v", cv)
		}
	})
}
