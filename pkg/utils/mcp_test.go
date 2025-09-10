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
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"path"
	"path/filepath"
	"testing"
)

func TestOAS3ToMCPValues(t *testing.T) {
	type args struct {
		spec string
	}
	tests := []struct {
		name    string
		spec    string
		wantErr error
	}{
		{
			name:    "petstore",
			spec:    "oas3/petstore/oas3.yaml",
			wantErr: nil,
		},
		{
			name:    "npr",
			spec:    "oas3/npr/oas3.yaml",
			wantErr: errors.New("Operation at $.paths./v2/ratings.POST is missing 'requestBody' property"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inSpec := filepath.Join("testdata", "specs", tt.spec)
			testDir := filepath.Join("testdata", "mcp", tt.name)

			valuesMap, err := OAS3ToMCPValues(inSpec)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}
			require.NoError(t, err)

			outValuesFile := path.Join(testDir, "out-values.yaml")
			outValuesMapText, err := yaml.Marshal(valuesMap)
			require.NoError(t, err)
			err = WriteOutputText(outValuesFile, outValuesMapText)
			require.NoError(t, err)

			expValuesFile := path.Join(testDir, "exp-values.yaml")
			exptValuesMapText, err := ReadInputTextFile(expValuesFile)
			require.NoError(t, err)

			assert.YAMLEq(t, string(exptValuesMapText), string(outValuesMapText))
		})
	}
}
