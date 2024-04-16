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

package bundle

import (
	"archive/zip"
	"fmt"
	v1 "github.com/micovery/apigee-yaml-toolkit/pkg/apigee/v1"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

			proxyBundle := filepath.Join(bundlesDir, tt.dir, "bundle.zip")
			apiProxyYAMLFile := filepath.Join(bundlesDir, tt.dir, "apiproxy.yaml")
			apiProxyYAML := utils.MustReadFileBytes(apiProxyYAMLFile)

			err = ProxyBundle2YAMLFile(proxyBundle, "", true)
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
			expectedBundleZipPath := filepath.Join(bundlesDir, tt.dir, "bundle.zip")
			outputBundleZipPath := filepath.Join(tmpDir, "bundle.zip")

			err = v1.APIProxyModelYAML2Bundle(apiProxyModelYAMLPath, outputBundleZipPath, false, "")
			require.NoError(t, err)
			RequireBundleZipEquals(t, expectedBundleZipPath, outputBundleZipPath)
		})
	}
}

func RequireBundleZipEquals(t *testing.T, expectedBundleZip string, actualBundleZip string) {
	expectedReader, err := zip.OpenReader(expectedBundleZip)
	require.NoError(t, err)
	defer MustClose(expectedReader)

	actualReader, err := zip.OpenReader(actualBundleZip)
	require.NoError(t, err)
	defer MustClose(actualReader)

	getFilesSorted := func(reader *zip.ReadCloser) []*zip.File {
		zipFiles := []*zip.File{}
		for _, f := range reader.File {
			if f.FileInfo().IsDir() {
				continue
			}
			zipFiles = append(zipFiles, f)
		}

		slices.SortFunc(zipFiles, func(a, b *zip.File) int {
			return strings.Compare(a.Name, b.Name)
		})

		return zipFiles
	}

	expectedFiles := getFilesSorted(expectedReader)
	actualFiles := getFilesSorted(actualReader)

	getFileNames := func(files []*zip.File) []string {
		result := []string{}
		for _, file := range files {
			result = append(result, file.Name)
		}

		return result
	}

	expectedFileNames := getFileNames(expectedFiles)
	actualFileNames := getFileNames(actualFiles)

	require.Equal(t, expectedFileNames, actualFileNames, "bundle structures do not match")
	for index, expectedFile := range expectedFiles {
		actualFile := actualFiles[index]

		expectedFileReader, err := expectedFile.Open()
		require.NoError(t, err)

		actualFileReader, err := actualFile.Open()
		require.NoError(t, err)

		extension := filepath.Ext(actualFile.Name)
		if extension == ".xml" {
			expected, err := utils.XMLText2YAMLText(expectedFileReader)
			require.NoError(t, err)

			actual, err := utils.XMLText2YAMLText(actualFileReader)
			require.NoError(t, err)

			require.YAMLEq(t, string(expected), string(actual), fmt.Sprintf("%s XML contents do not match", expectedFile.Name))
		} else {
			expectedContents, err := io.ReadAll(expectedFileReader)
			require.NoError(t, err)

			actualContents, err := io.ReadAll(actualFileReader)
			require.Equal(t, string(expectedContents), string(actualContents), fmt.Sprintf("%s contents do not match", expectedFile.Name))
		}
	}
}

func MustClose(reader *zip.ReadCloser) {
	err := reader.Close()
	if err != nil {
		panic(err)
	}
}
