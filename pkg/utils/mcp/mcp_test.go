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
	"errors"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"path"
	"path/filepath"
	"testing"
)

func TestOAS3ToMCPValues(t *testing.T) {
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
			name:    "weather",
			spec:    "oas3/weather/oas3.yaml",
			wantErr: nil,
		},
		{
			name:    "petstore-oas2",
			spec:    "oas2/petstore/oas2.yaml",
			wantErr: errors.New("input file '../testdata/specs/oas2/petstore/oas2.yaml' does not contain 'openapi' field"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inSpec := filepath.Join("..", "testdata", "specs", tt.spec)
			testDir := filepath.Join("..", "testdata", "mcp", tt.name)

			valuesMap, err := OAS3ToMCPValues(inSpec)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}
			require.NoError(t, err)

			outValuesFile := path.Join(testDir, "out-values.yaml")
			outValuesMapText, err := yaml.Marshal(valuesMap)
			require.NoError(t, err)
			err = utils.WriteOutputText(outValuesFile, outValuesMapText)
			require.NoError(t, err)

			expValuesFile := path.Join(testDir, "exp-values.yaml")
			exptValuesMapText, err := utils.ReadInputTextFile(expValuesFile)
			require.NoError(t, err)

			assert.YAMLEq(t, string(exptValuesMapText), string(outValuesMapText))
		})
	}
}

func TestGenerateAlternateOperationID(t *testing.T) {
	tests := []struct {
		name     string
		apiPath  string
		httpVerb string
		want     string
		wantErr  bool
	}{
		{
			name:     "Standard Path",
			apiPath:  "/users",
			httpVerb: "POST",
			want:     "post-users",
		},
		{
			name:     "Multi-Segment Path with Parameter",
			apiPath:  "/v1/projects/{projectId}",
			httpVerb: "PATCH",
			want:     "patch-v1_projects_projectid",
		},
		{
			name:     "Hyphens and Mixed Case in Path",
			apiPath:  "/Data-Service/status",
			httpVerb: "GET",
			want:     "get-data_service_status",
		},
		{
			name:     "Path Parameter with Underscore",
			apiPath:  "/items/{item_name}",
			httpVerb: "DELETE",
			want:     "delete-items_item_name",
		},
		{
			name:     "Root Path",
			apiPath:  "/",
			httpVerb: "OPTIONS",
			want:     "options",
		},
		{
			name:     "Complex Path (Multiple special chars)",
			apiPath:  "/a/b-c/{d_e}/f",
			httpVerb: "PUT",
			want:     "put-a_b_c_d_e_f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateAlternateOperationID(tt.apiPath, tt.httpVerb)

			// Check for expected result only if no error was expected
			if got != tt.want {
				t.Errorf("generateAlternateOperationID() = %v, want %v", got, tt.want)
			}
		})
	}
}
