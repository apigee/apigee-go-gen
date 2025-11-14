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

package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

// TestRemoveSchemaExtensions implements table-driven tests for the schema extension removal logic.
func TestRemoveSchemaExtensions(t *testing.T) {

	tests := []struct {
		dir string
	}{
		{
			"deeply-nested",
		},
		{
			"old-definitions",
		},
		{
			"properties-level",
		},
		{
			"ref-sibling",
		},
		{
			"schema-level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			dir := filepath.Join("testdata", "remove-oas-schema-extensions", tt.dir)
			inFile := filepath.Join(dir, "oas.yaml")
			outFile := filepath.Join(dir, "out-oas.yaml")
			expFile := filepath.Join(dir, "exp-oas.yaml")

			err := RemoveSchemaExtensions(inFile, outFile)
			require.NoError(t, err)
			expYAML := MustReadFileBytes(expFile)
			outYAML := MustReadFileBytes(outFile)

			expYAML = RemoveYAMLComments(expYAML)
			outYAML = RemoveYAMLComments(outYAML)
			require.YAMLEq(t, string(expYAML), string(outYAML))
			assert.NoError(t, err)
		})
	}
}
