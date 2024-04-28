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
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAPI2FileToOpenAPI3File(t *testing.T) {

	tests := []struct {
		name         string
		dir          string
		inputFile    string
		expectedFile string
		allowCycles  bool
		wantErr      error
	}{
		{

			"petstore OAS2(JSON) to OAS3(JSON)",
			"petstore",
			"oas2.json",
			"oas3.json",
			false,
			nil,
		},
		{
			"petstore OAS2(JSON) to OAS3(YAML)",
			"petstore",
			"oas2.json",
			"oas3.yaml",
			false,
			nil,
		},
		{
			"petstore OAS2(YAML) to OAS3(YAML)",
			"petstore",
			"oas2.yaml",
			"oas3.yaml",
			false,
			nil,
		},
		{
			"petstore OAS2(YAML) to OAS3(JSON)",
			"petstore",
			"oas2.yaml",
			"oas3.json",
			false,
			nil,
		},
		{

			"petstore-refs OAS2(JSON) to OAS3(JSON)",
			"petstore-refs",
			"oas2.json",
			"oas3.json",
			false,
			nil,
		},
		{
			"petstore-cycle OAS2(JSON) to OAS3(JSON)",
			"petstore-cycle",
			"oas2.json",
			"oas3.json",
			true,
			nil,
		},
		{
			"petstore-cycle OAS2(JSON) to OAS3(JSON)",
			"petstore-cycle",
			"oas2.json",
			"oas3.json",
			false,
			errors.New("cyclic JSONRef at $.definitions.Widgets.properties.widgets.items.properties.subWidgets"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttDir := filepath.Join("testdata", "oas2", tt.dir)
			inputFile := filepath.Join(ttDir, tt.inputFile)
			outputFile := filepath.Join(ttDir, fmt.Sprint("out-", tt.expectedFile))
			expectedFile := filepath.Join(ttDir, fmt.Sprint(tt.expectedFile))

			var err error
			err = os.RemoveAll(outputFile)
			require.NoError(t, err)

			err = OAS2FileToOAS3File(inputFile, outputFile, tt.allowCycles)
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
