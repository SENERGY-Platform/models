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
	"encoding/json"
	"strings"
	"testing"
)

// The two path option shapes answer two different endpoints and spell their fields
// differently. Moving them into this package put both spellings in one place for the first
// time, which is also the first chance to mix them up.
func TestPathOptionSpelling(t *testing.T) {
	t.Run("a selectables path option spells its fields camelCase", func(t *testing.T) {
		assertKeys(t, PathOption{}, "path", "characteristicId", "aspectNode", "functionId", "isVoid")
	})

	t.Run("a device-type-selectables path option spells its fields snake_case", func(t *testing.T) {
		assertKeys(t, ServicePathOption{}, "service_id", "path", "characteristic_id", "aspect_node", "function_id", "is_void", "is_controlling_function", "interaction")
	})

	t.Run("a configurable spells its fields snake_case in both shapes", func(t *testing.T) {
		assertKeys(t, Configurable{}, "path", "characteristic_id", "aspect_node", "function_id")
	})
}

// The aspect list is omitempty on every shape, so an answer naming a single aspect looks
// exactly as it did before the list existed. That is what lets a client and a whole-payload
// fixture survive the change untouched.
func TestAspectNodeListIsOmittedWhenEmpty(t *testing.T) {
	for name, value := range map[string]interface{}{
		"selectables path option":             PathOption{AspectNode: AspectNode{Id: "aid"}},
		"device-type-selectables path option": ServicePathOption{AspectNode: AspectNode{Id: "aid"}},
		"configurable":                        Configurable{AspectNode: AspectNode{Id: "aid"}},
	} {
		t.Run(name, func(t *testing.T) {
			encoded := encode(t, value)
			if strings.Contains(encoded, "aspect_nodes") || strings.Contains(encoded, "aspectNodes") {
				t.Error(encoded)
			}
		})
	}
}

func TestAspectNodeListIsCarriedWhenFilled(t *testing.T) {
	t.Run("a selectables path option carries aspectNodes", func(t *testing.T) {
		option := PathOption{AspectNodes: []AspectNode{{Id: "aid1"}, {Id: "aid2"}}}
		decoded := PathOption{}
		if err := json.Unmarshal([]byte(encode(t, option)), &decoded); err != nil {
			t.Fatal(err)
		}
		if len(decoded.AspectNodes) != 2 || decoded.AspectNodes[0].Id != "aid1" {
			t.Error(decoded.AspectNodes)
		}
	})

	t.Run("a device-type-selectables path option carries aspect_nodes", func(t *testing.T) {
		option := ServicePathOption{AspectNodes: []AspectNode{{Id: "aid1"}, {Id: "aid2"}}}
		decoded := ServicePathOption{}
		if err := json.Unmarshal([]byte(encode(t, option)), &decoded); err != nil {
			t.Fatal(err)
		}
		if len(decoded.AspectNodes) != 2 || decoded.AspectNodes[0].Id != "aid1" {
			t.Error(decoded.AspectNodes)
		}
	})
}

func assertKeys(t *testing.T, value interface{}, keys ...string) {
	t.Helper()
	decoded := map[string]interface{}{}
	if err := json.Unmarshal([]byte(encode(t, value)), &decoded); err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if _, ok := decoded[key]; !ok {
			t.Error("missing key", key, decoded)
		}
	}
}

func encode(t *testing.T, value interface{}) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}
