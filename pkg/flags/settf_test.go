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

package flags

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"

	// These are the imports from your original test file
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/apigee/apigee-go-gen/pkg/values"
	"github.com/stretchr/testify/require"
)

func TestSetTF_Set(t *testing.T) {
	tests := []struct {
		name string // Name of the test case
	}{
		{
			"list-tuple",
		},
		{
			"map-object",
		},
		{
			"simple-block-labeled",
		},
		{
			"empty",
		},
		{
			"primitives",
		},
		{
			"string-interpolation",
		},
		{
			"single-unlabeled-block",
		},
		{
			"multiple-unlabeled-blocks",
		},
		{
			"multiple-labeled-blocks",
		},
		{
			"mixed-content-and-nesting",
		},
		{
			"ordered-provisioner",
		},
		{
			"resource-lifecycle",
		},
		{
			"keycloak-full",
		},
		{
			"generic-logic",
		},
		{
			"block-duplication",
		},
		{
			"resource-and-data-blocks",
		},
		{
			"variable-block",
		},
		{
			"output-block",
		},
		{
			"locals-block",
		},
		{
			"module-block",
		},
		{
			"provider-block",
		},
		{
			"terraform-block",
		},
		{
			"module-with-providers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dir := filepath.Join("testdata", "settf", tt.name)
			inFile := filepath.Join(dir, "data.tf")
			outFile := filepath.Join(dir, fmt.Sprintf("out-%s", "data.tf.json"))
			expFile := filepath.Join(dir, "exp-data.tf.json")

			err := os.RemoveAll(outFile)
			require.NoError(t, err)

			data := values.Map{}
			v := NewSetTF(&data)

			// Create a temp file

			// Run the .Set() method
			err = v.Set(fmt.Sprintf("tf=%s", inFile))
			require.NoError(t, err)
			assert.Contains(t, data, "tf")

			var outFileJson []byte
			outFileJson, err = json.MarshalIndent(data["tf"], "", "  ")
			assert.NoError(t, err)

			err = os.WriteFile(outFile, outFileJson, os.ModePerm)
			assert.NoError(t, err)

			outBytes := utils.MustReadFileBytes(outFile)
			wantBytes := utils.MustReadFileBytes(expFile)
			require.JSONEq(t, string(wantBytes), string(outBytes))

		})
	}
}
