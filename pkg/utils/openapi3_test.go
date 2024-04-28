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
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAPI3FileToOpenAPI2File(t *testing.T) {

	tests := []struct {
		name         string
		dir          string
		inputFile    string
		expectedFile string
		allowCycles  bool
		wantErr      error
	}{
		{

			"npr OAS3(JSON) to OAS2(JSON)",
			"npr",
			"oas3.json",
			"oas2.json",
			false,
			nil,
		},
		{
			"npr OAS3(JSON) to OAS2(YAML)",
			"npr",
			"oas3.json",
			"oas2.yaml",
			false,
			nil,
		},
		{
			"npr OAS3(YAML) to OAS2(YAML)",
			"npr",
			"oas3.yaml",
			"oas2.yaml",
			false,
			nil,
		},
		{
			"npr OAS3(YAML) to OAS2(JSON)",
			"npr",
			"oas3.yaml",
			"oas2.json",
			false,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttDir := filepath.Join("testdata", "oas3-to-oas2", tt.dir)
			inputFile := filepath.Join(ttDir, tt.inputFile)
			outputFile := filepath.Join(ttDir, fmt.Sprint("out-", tt.expectedFile))
			expectedFile := filepath.Join(ttDir, fmt.Sprint(tt.expectedFile))

			var err error
			err = os.RemoveAll(outputFile)
			require.NoError(t, err)

			err = OAS3FileToOAS2File(inputFile, outputFile, tt.allowCycles)
			if tt.wantErr != nil {
				require.EqualError(t, tt.wantErr, err.Error())
				return
			}

			require.NoError(t, err)

			outputBytes := MustReadFileBytes(outputFile)
			expectedBytes := MustReadFileBytes(expectedFile)

			if filepath.Ext(expectedFile) == ".json" {
				require.JSONEq(t, string(expectedBytes), string(outputBytes))
			} else if filepath.Ext(expectedFile) == ".yaml" {
				outputBytes = RemoveYAMLComments(outputBytes)
				expectedBytes = RemoveYAMLComments(expectedBytes)
				require.YAMLEq(t, string(expectedBytes), string(outputBytes))
			} else {
				t.Error("unknown output format in testcase")
			}

		})
	}
}
