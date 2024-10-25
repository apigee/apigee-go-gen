//  Copyright 2024 Google LLC
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

package utils

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestXMLText2YAMLText(t *testing.T) {
	tests := []struct {
		dir string
	}{
		{
			"simple_nested",
		},
		{
			"scalar_without_attrs",
		},
		{
			"element_with_attr",
		},
		{
			"scalar_with_attrs",
		},
		{
			"sequence_parent_without_attrs",
		},
		{
			"sequence_parent_with_attrs",
		},
		{
			"sequence_without_parent",
		},
		{
			"sequence_without_parent_with_attrs",
		},
		{
			"unique_children_with_attrs_parent_without_attrs",
		},
		{
			"unique_children_without_attrs_parent_without_attrs",
		},
		{
			"repeated_children_without_attrs_parent_without_attrs",
		},
		{
			"complex_raise_fault_policy",
		},
		{
			"flow_callout_policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			dir := filepath.Join("testdata", "snippets", tt.dir)
			inFile := filepath.Join(dir, "data.xml")
			inBytes := MustReadFileBytes(inFile)

			wantFile := filepath.Join(dir, "data.yaml")
			wantBytes := MustReadFileBytes(wantFile)

			gotBytes, err := XMLText2YAMLText(bytes.NewReader(inBytes))
			assert.NoError(t, err)
			wantBytes = RemoveYAMLComments(wantBytes)
			assert.Equal(t, string(wantBytes), string(gotBytes))
		})
	}
}
