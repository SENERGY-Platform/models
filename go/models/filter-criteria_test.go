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

import "testing"

func TestFilterCriteriaShort(t *testing.T) {
	t.Run("device-group deprecated aspect id", func(t *testing.T) {
		criteria := DeviceGroupFilterCriteria{
			Interaction:   EVENT,
			FunctionId:    "fid",
			AspectId:      "aid",
			DeviceClassId: "dcid",
		}
		if criteria.Short() != "fid_aid_dcid_event" {
			t.Error(criteria.Short())
		}
	})

	t.Run("device-group aspect id as alias for single element list", func(t *testing.T) {
		withAspectId := DeviceGroupFilterCriteria{FunctionId: "fid", AspectId: "aid"}
		withAspectIds := DeviceGroupFilterCriteria{FunctionId: "fid", AspectIds: []string{"aid"}}
		if withAspectId.Short() != withAspectIds.Short() {
			t.Error(withAspectId.Short(), withAspectIds.Short())
		}
	})

	t.Run("device-group aspect ids", func(t *testing.T) {
		criteria := DeviceGroupFilterCriteria{
			Interaction:   EVENT,
			FunctionId:    "fid",
			AspectIds:     []string{"aid2", "aid1"},
			DeviceClassId: "dcid",
		}
		if criteria.Short() != "fid_aid1,aid2_dcid_event" {
			t.Error(criteria.Short())
		}
	})

	t.Run("device-group aspect ids order is irrelevant", func(t *testing.T) {
		asc := DeviceGroupFilterCriteria{AspectIds: []string{"aid1", "aid2"}}
		desc := DeviceGroupFilterCriteria{AspectIds: []string{"aid2", "aid1"}}
		if asc.Short() != desc.Short() {
			t.Error(asc.Short(), desc.Short())
		}
	})

	t.Run("device-group aspect ids are not reordered in place", func(t *testing.T) {
		aspectIds := []string{"aid2", "aid1"}
		criteria := DeviceGroupFilterCriteria{AspectIds: aspectIds}
		criteria.Short()
		if aspectIds[0] != "aid2" || aspectIds[1] != "aid1" {
			t.Error(aspectIds)
		}
	})

	t.Run("import-type deprecated aspect id", func(t *testing.T) {
		criteria := ImportTypeFilterCriteria{FunctionId: "fid", AspectId: "aid"}
		if criteria.Short() != "aid_fid" {
			t.Error(criteria.Short())
		}
	})

	t.Run("import-type aspect ids", func(t *testing.T) {
		criteria := ImportTypeFilterCriteria{FunctionId: "fid", AspectIds: []string{"aid2", "aid1"}}
		if criteria.Short() != "aid1,aid2_fid" {
			t.Error(criteria.Short())
		}
	})

	t.Run("no aspect", func(t *testing.T) {
		criteria := ImportTypeFilterCriteria{FunctionId: "fid"}
		if criteria.Short() != "_fid" {
			t.Error(criteria.Short())
		}
	})
}
