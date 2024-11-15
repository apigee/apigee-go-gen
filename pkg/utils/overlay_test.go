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
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestOASOverlay(t *testing.T) {
	type args struct {
	}

	tests := []struct {
		dir      string
		specFile string
		wantErr  error
	}{
		{
			"structured",
			"petstore/oas3.yaml",
			nil,
		},
		{
			"targeted",
			"petstore/oas3.yaml",
			nil,
		},
		{
			"wildcard",
			"petstore/oas3.yaml",
			nil,
		},
		{
			"array",
			"petstore/oas3.yaml",
			nil,
		},
		{
			"bad-update-target",
			"petstore/oas3.yaml",
			errors.New("invalid array index [?@.name=='status' && @.in=='query'] before position 58: non-integer array index at testdata/oas-overlay/bad-update-target/overlay.yaml:19"),
		},
		{
			"bad-remove-value",
			"petstore/oas3.yaml",
			errors.New("'remove' field within Overlay action is not boolean at testdata/oas-overlay/bad-remove-value/overlay.yaml:20"),
		},
		{
			"bad-action-op",
			"petstore/oas3.yaml",
			errors.New("action does not contain neither 'remove' nor 'update' field at testdata/oas-overlay/bad-action-op/overlay.yaml:19"),
		},
		{
			"bad-actions-value",
			"petstore/oas3.yaml",
			errors.New("'actions' field must be an array at testdata/oas-overlay/bad-actions-value/overlay.yaml:18"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			testDir := path.Join("testdata", "oas-overlay", tt.dir)
			specFile := path.Join("testdata", "specs", "oas3", tt.specFile)
			overlayFile := path.Join(testDir, "overlay.yaml")
			outputFile := path.Join(testDir, "out-oas3.yaml")
			expectedFile := path.Join(testDir, "exp-oas3.yaml")

			var err error
			err = os.RemoveAll(outputFile)
			require.NoError(t, err)

			err = OASOverlay(overlayFile, specFile, outputFile)

			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
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
