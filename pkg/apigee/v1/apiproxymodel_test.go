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

package v1

import (
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

// this test round-trips the input YAML through the in-memory model, and back to YAML
// if there are no errors, and the output YAML matches the input, it means that the in-memory
// model was hydrated correctly from the input YAML

func TestNewAPIProxyModel(t *testing.T) {

	tests := []struct {
		name string
	}{
		{
			"simple",
		},
		{
			"postclient",
		},
		{
			"health-monitor-http",
		},
		{
			"health-monitor-tcp",
		},
		{
			"lb-server-unhealthy",
		},
		{
			"lb-weighted",
		},
		{
			"google-id-token-auth",
		},
	}
	for _, tt := range tests {
		ttDir := filepath.Join("testdata", "yaml-first", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			inputFile := filepath.Join(ttDir, "apiproxy.yaml")
			expFile := filepath.Join(ttDir, "exp-apiproxy.yaml")

			if _, err := os.Stat(expFile); errors.Is(err, os.ErrNotExist) {
				//if there is no expected file, use the input itself as expected
				expFile = inputFile
			}

			outFile := filepath.Join(ttDir, "out-apiproxy.yaml")

			model, err := NewAPIProxyModel(inputFile)
			require.NoError(t, err)

			err = model.Validate()
			require.NoError(t, err)

			outData, err := model.YAML()
			require.NoError(t, err)

			//for convenience write the output to disk (makes it easier to diff)
			err = os.WriteFile(outFile, outData, os.ModePerm)
			require.NoError(t, err)

			expData := utils.MustReadFileBytes(expFile)

			require.YAMLEq(t, string(expData), string(outData))
		})
	}
}
