//  Copyright 2025 Google LLC
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package mcp

import (
	"testing"
)

func TestJsonPointerToJSONPath(t *testing.T) {
	tests := []struct {
		name         string
		refPath      string
		wantJSONPath string
		wantErr      bool
	}{
		{
			name:         "Standard Reference",
			refPath:      "#/components/schemas/User",
			wantJSONPath: "$['components']['schemas']['User']",
			wantErr:      false,
		},
		{
			name:         "Reference with Dots in Segment",
			refPath:      "#/components/schemas/io.k8s.api.core.v1.Pod",
			wantJSONPath: "$['components']['schemas']['io.k8s.api.core.v1.Pod']",
			wantErr:      false,
		},
		{
			name:         "Reference with JSON Pointer Tilde Escaping (~0)",
			refPath:      "#/paths/~01example", // "~0" should become "~"
			wantJSONPath: "$['paths']['~1example']",
			wantErr:      false,
		},
		{
			name:         "Reference with JSON Pointer Slash Escaping (~1)",
			refPath:      "#/paths/slash~1example", // "~1" should become "/"
			wantJSONPath: "$['paths']['slash/example']",
			wantErr:      false,
		},
		{
			name:         "Reference with both Escapings",
			refPath:      "#/a~1b/c~0d", // "~1" -> "/", "~0" -> "~"
			wantJSONPath: "$['a/b']['c~d']",
			wantErr:      false,
		},
		{
			name:         "Root Reference Only",
			refPath:      "#",
			wantJSONPath: "$",
			wantErr:      false,
		},
		{
			name:         "Invalid: Missing Hash Prefix",
			refPath:      "/components/schemas/User",
			wantJSONPath: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONPointerToJSONPath(tt.refPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("jsonPointerToJSONPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.wantJSONPath {
				t.Errorf("jsonPointerToJSONPath() got = %v, want %v", got, tt.wantJSONPath)
			}
		})
	}
}
