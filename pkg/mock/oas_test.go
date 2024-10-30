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

package mock

import (
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestGenerateMockProxyBundle(t *testing.T) {
	tests := []struct {
		mock string
	}{
		{
			"petstore",
		},
	}

	mocksDir := filepath.Join("testdata", "mocks")
	specsDir := filepath.Join("..", "utils", "testdata", "specs", "oas3")

	for _, tt := range tests {
		t.Run(tt.mock, func(t *testing.T) {

			inputPath := filepath.Join(specsDir, tt.mock, "oas3.yaml")
			outputPath := filepath.Join(mocksDir, tt.mock, "out-apiproxy.zip")
			expectedOutputPath := filepath.Join(mocksDir, tt.mock, "exp-apiproxy.zip")

			err := GenerateMockProxyBundle(inputPath, outputPath, false)
			require.NoError(t, err)

			utils.RequireBundleZipEquals(t, outputPath, expectedOutputPath)
		})
	}
}
