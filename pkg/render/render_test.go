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

package render

import (
	"fmt"
	"github.com/apigee/apigee-go-gen/pkg/flags"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderGeneric(t *testing.T) {

	tests := []struct {
		dir          string
		templateFile string
		valuesFile   string
		includesFile string
		setFileFlag  string
		wantErr      error
	}{
		{
			"using-files",
			"input.yaml",
			"",
			"",
			"",
			nil,
		},
		{
			"using-helpers",
			"input.yaml",
			"",
			"_helpers.tmpl",
			"",
			nil,
		},
		{
			"policies",
			"apiproxy.yaml",
			"values.yaml",
			"",
			"",
			nil,
		},
		{
			"set-file",
			"input.yaml",
			"",
			"",
			"data=./data.json",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {

			testDir := path.Join("testdata", "render", tt.dir)
			templateFile := path.Join(testDir, tt.templateFile)

			outputFile := path.Join(testDir, fmt.Sprintf("out-%s", tt.templateFile))
			expectedFile := path.Join(testDir, fmt.Sprintf("exp-%s", tt.templateFile))

			var err error
			err = os.RemoveAll(outputFile)
			require.NoError(t, err)

			type ctx struct{}
			cFlags := NewCommonFlags()
			cFlags.TemplateFile = flags.String(templateFile)
			cFlags.OutputFile = flags.String(outputFile)

			if tt.valuesFile != "" {
				v := flags.NewValues(cFlags.Values)
				valuesFile := path.Join(testDir, tt.valuesFile)
				err := v.Set(valuesFile)
				require.NoError(t, err)
			}

			if tt.includesFile != "" {
				includesFile := path.Join(testDir, tt.includesFile)
				err := cFlags.IncludeList.Set(includesFile)
				require.NoError(t, err)
			}

			if tt.setFileFlag != "" {
				key, filePath, _ := strings.Cut(tt.setFileFlag, "=")
				tt.setFileFlag = fmt.Sprintf("%s=%s", key, path.Join(testDir, filePath))
				f := flags.NewSetFile(cFlags.Values)
				err := f.Set(tt.setFileFlag)
				require.NoError(t, err)
			}

			err = RenderGenericTemplate(cFlags, false)

			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}
			require.NoError(t, err)

			outputBytes := utils.MustReadFileBytes(outputFile)
			expectedBytes := utils.MustReadFileBytes(expectedFile)

			if filepath.Ext(expectedFile) == ".txt" {
				require.Equal(t, string(expectedBytes), string(outputBytes))
			} else if filepath.Ext(expectedFile) == ".json" {
				require.JSONEq(t, string(expectedBytes), string(outputBytes))
			} else if filepath.Ext(expectedFile) == ".yaml" {
				outputBytes = utils.RemoveYAMLComments(outputBytes)
				expectedBytes = utils.RemoveYAMLComments(expectedBytes)
				require.YAMLEq(t, string(expectedBytes), string(outputBytes))
			} else {
				t.Error("unknown output format in testcase")
			}

		})
	}
}
