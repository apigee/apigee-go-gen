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
	}

	bundlesDir := filepath.Join("testdata", "bundles")
	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			stdout, err := utils.NewStdoutCapture()
			assert.NoError(t, err)

			defer stdout.Restore()

			proxyBundle := filepath.Join(bundlesDir, tt.dir, "apiproxy.zip")
			apiProxyYAMLFile := filepath.Join(bundlesDir, tt.dir, "apiproxy.yaml")
			apiProxyYAML := utils.MustReadFileBytes(apiProxyYAMLFile)

			err = Bundle2YAMLFile(proxyBundle, "", true)
			require.NoError(t, err)

			data, err := stdout.Read()
			require.NoError(t, err)

			actualYAML := string(data)
			expectedYAML := string(apiProxyYAML)
			assert.YAMLEq(t, expectedYAML, actualYAML)
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
	}

	bundlesDir := filepath.Join("testdata", "bundles")

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {

			tmpDir, err := os.MkdirTemp("", tt.dir+"-*")
			require.NoError(t, err)

			apiProxyModelYAMLPath := filepath.Join(bundlesDir, tt.dir, "apiproxy.yaml")
			expectedBundleZipPath := filepath.Join(bundlesDir, tt.dir, "apiproxy.zip")
			outputBundleZipPath := filepath.Join(tmpDir, "apiproxy.zip")

			model, err := v1.NewAPIProxyModel(apiProxyModelYAMLPath)
			require.NoError(t, err)

			err = render.CreateBundle(model, outputBundleZipPath, false, "")
			require.NoError(t, err)

			utils.RequireBundleZipEquals(t, expectedBundleZipPath, outputBundleZipPath)
		})
	}
}
