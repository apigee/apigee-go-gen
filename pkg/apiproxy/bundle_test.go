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

package apiproxy

import (
	v1 "github.com/apigee/apigee-go-gen/pkg/apigee/v1"
	"github.com/apigee/apigee-go-gen/pkg/render"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestProxyBundle2YAMLFile(t *testing.T) {
	tests := []struct {
		dir string
	}{
		{
			"helloworld",
		},
		{
			"oauth-validate-key-secret",
		},
		{
			"integration-target",
		},
	}

	bundlesDir := filepath.Join("testdata", "bundles")
	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			stdout, err := utils.NewStdoutCapture()
			assert.NoError(t, err)

			defer stdout.Restore()

			inputAPIProxyBundle := filepath.Join(bundlesDir, tt.dir, "apiproxy.zip")
			expectedYAMLFile := filepath.Join(bundlesDir, tt.dir, "apiproxy.yaml")
			outputYAMLFile := filepath.Join(bundlesDir, tt.dir, "out-apiproxy.yaml")

			err = os.RemoveAll(outputYAMLFile)
			require.NoError(t, err)

			err = Bundle2YAMLFile(inputAPIProxyBundle, "", true)
			require.NoError(t, err)

			data, err := stdout.Read()
			require.NoError(t, err)

			err = utils.WriteOutputText(outputYAMLFile, data)
			assert.NoError(t, err)

			expectedYAMLBytes := utils.MustReadFileBytes(expectedYAMLFile)
			outputYAMLBytes := utils.MustReadFileBytes(outputYAMLFile)

			assert.YAMLEq(t, string(expectedYAMLBytes), string(outputYAMLBytes))
		})
	}
}

func TestAPIProxyModel2BundleZip(t *testing.T) {
	tests := []struct {
		dir string
	}{
		{
			"helloworld",
		},
		{
			"oauth-validate-key-secret",
		},
		{
			"integration-target",
		},
	}

	bundlesDir := filepath.Join("testdata", "bundles")

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			var err error
			inputAPIProxyYAMLFile := filepath.Join(bundlesDir, tt.dir, "apiproxy.yaml")
			expectedAPIProxyBundleFile := filepath.Join(bundlesDir, tt.dir, "apiproxy.zip")
			outputAPIProxyBundleFile := filepath.Join(bundlesDir, tt.dir, "out-apiproxy.zip")

			err = os.RemoveAll(outputAPIProxyBundleFile)
			require.NoError(t, err)

			model, err := v1.NewAPIProxyModel(inputAPIProxyYAMLFile)
			require.NoError(t, err)

			err = render.CreateBundle(model, outputAPIProxyBundleFile, false, "")
			require.NoError(t, err)

			utils.RequireBundleZipEquals(t, expectedAPIProxyBundleFile, outputAPIProxyBundleFile)
		})
	}
}
