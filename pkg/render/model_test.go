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
	"github.com/apigee/apigee-go-gen/pkg/globals"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"path/filepath"
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

func TestGenerateBundleRejectsSymlinkEscape(t *testing.T) {
	base := t.TempDir()
	outside := filepath.Join(base, "outside")
	require.NoError(t, os.MkdirAll(outside, os.ModePerm))

	templateDir := filepath.Join(base, "template")
	require.NoError(t, os.MkdirAll(templateDir, os.ModePerm))
	templateFile := filepath.Join(templateDir, "apiproxy.yaml")
	require.NoError(t, os.WriteFile(templateFile, []byte(`#{{ os_writefile "linkdir/pwned" "pwned" }}`), os.ModePerm))
	require.NoError(t, os.Symlink(outside, filepath.Join(templateDir, "linkdir")))

	cFlags := NewCommonFlags()
	cFlags.TemplateFile = flags.String(templateFile)
	cFlags.OutputFile = flags.String(filepath.Join(base, "out", "apiproxy.zip"))

	createModelFunc := func(input string) (v1.Model, error) {
		return v1.NewAPIProxyModel(input)
	}

	err := GenerateBundle(createModelFunc, cFlags, false, "", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "points outside of")
	require.NoFileExists(t, filepath.Join(outside, "pwned"))
}

func TestFileHelpersRejectSymlinkEscape(t *testing.T) {
	//helper panics are reported through recoverPanic, which reads this CLI flag
	if globals.ShowStack == nil {
		showStack := flags.Bool(false)
		globals.ShowStack = &showStack
	}

	tests := []struct {
		name     string
		template string
	}{
		{"os_writefile", `{{ os_writefile "linkdir/pwned" "pwned" }}`},
		{"os_copyfile", `{{ os_copyfile "linkdir/pwned" "data.txt" }}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := t.TempDir()
			outside := filepath.Join(base, "outside")
			require.NoError(t, os.MkdirAll(outside, os.ModePerm))

			templateFile := filepath.Join(base, "template", "template.txt")
			require.NoError(t, os.MkdirAll(filepath.Dir(templateFile), os.ModePerm))
			require.NoError(t, os.WriteFile(templateFile, []byte(tt.template), os.ModePerm))
			require.NoError(t, os.WriteFile(filepath.Join(base, "template", "data.txt"), []byte("pwned"), os.ModePerm))

			//simulate a symlink already present in the output directory
			outputFile := filepath.Join(base, "out", "out.txt")
			require.NoError(t, os.MkdirAll(filepath.Dir(outputFile), os.ModePerm))
			require.NoError(t, os.Symlink(outside, filepath.Join(base, "out", "linkdir")))

			tmpl, err := CreateTemplate(templateFile, "", nil, outputFile, false)
			require.NoError(t, err)

			_, err = RenderTemplate(tmpl, map[string]any{})
			require.Error(t, err)
			require.Contains(t, err.Error(), "path escapes from parent")
			require.NoFileExists(t, filepath.Join(outside, "pwned"))
		})
	}
}
