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

package render

import (
	"fmt"
	v1 "github.com/apigee/apigee-go-gen/pkg/apigee/v1"
	"github.com/apigee/apigee-go-gen/pkg/flags"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func TestGenerateBundle(t *testing.T) {

	tests := []struct {
		dir          string
		templateFile string
		specFile     string
		wantErr      error
	}{
		{
			"github-oas3",
			"https://github.com/apigee/apigee-go-gen/blob/main/examples/templates/oas3/apiproxy.yaml",
			"../utils/testdata/specs/oas3/petstore/oas3.yaml",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {

			testDir := path.Join("testdata", "render-apiproxy", tt.dir)

			expectedFile := path.Join(testDir, "exp-apiproxy.zip")
			outputFile := path.Join(testDir, "out-apiproxy.zip")

			var err error
			err = os.RemoveAll(outputFile)
			require.NoError(t, err)

			cFlags := NewCommonFlags()
			cFlags.TemplateFile = flags.String(tt.templateFile)
			cFlags.OutputFile = flags.String(outputFile)

			f := flags.NewSetOAS(cFlags.Values)
			err = f.Set(fmt.Sprintf("spec=%s", tt.specFile))
			require.NoError(t, err)

			createModelFunc := func(input string) (v1.Model, error) {
				return v1.NewAPIProxyModel(input)
			}

			err = GenerateBundle(createModelFunc, cFlags, false, "", false)
			require.NoError(t, err)

			utils.RequireBundleZipEquals(t, expectedFile, outputFile)

		})
	}
}
