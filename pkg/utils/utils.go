// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func MustReadFileBytes(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return data
}

func MustRemoveAll(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		panic(err)
	}
}

func AddEntryToOASYAML(oas *yaml.Node, key string, value any, defaultVal *yaml.Node) (*yaml.Node, error) {
	oas.Content = append(oas.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key})

	//the kin-openapi library only uses JSON tags, so we must marshall to JSON first
	jsonText, err := json.Marshal(value)
	if err != nil {
		return nil, errors.New(err)
	}

	yamlText, err := JSONText2YAMLText(bytes.NewReader(jsonText))
	if err != nil {
		return nil, err
	}

	yamlNode := yaml.Node{}
	err = yaml.Unmarshal(yamlText, &yamlNode)
	if err != nil {
		return nil, errors.New(err)
	}

	content := yamlNode.Content[0]
	if content.Kind == yaml.ScalarNode && content.Value == "null" && defaultVal != nil {
		content = defaultVal
	}
	oas.Content = append(oas.Content, content)
	return &yamlNode, nil
}

type MultiError struct {
	Errors []error
}

func (e MultiError) Error() string {
	return errors.Join(e.Errors...).Error()
}

func PushDir(dir string) func() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	popDir := func() {
		Must(os.Chdir(wd))
	}

	return popDir
}

func RemoveYAMLComments(data []byte) []byte {
	regex := regexp.MustCompile(`(?ms)^\s*#[^\n\r]*$[\r\n]*`)
	replaced := regex.ReplaceAll(data, []byte{})
	return replaced
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

	require.Equal(t, expectedFileNames, actualFileNames, "API proxy structures do not match")
	for index, expectedFile := range expectedFiles {
		actualFile := actualFiles[index]

		expectedFileReader, err := expectedFile.Open()
		require.NoError(t, err)

		actualFileReader, err := actualFile.Open()
		require.NoError(t, err)

		extension := filepath.Ext(actualFile.Name)
		if extension == ".xml" {
			expected, err := XMLText2YAMLText(expectedFileReader)
			require.NoError(t, err)

			expected = RemoveYAMLComments(expected)
			actual, err := XMLText2YAMLText(actualFileReader)
			require.NoError(t, err)

			require.YAMLEq(t, string(expected), string(actual), fmt.Sprintf("%s XML contents do not match", expectedFile.Name))
		} else {
			expectedContents, err := io.ReadAll(expectedFileReader)
			require.NoError(t, err)
			expectedContents = RemoveYAMLComments(expectedContents)

			actualContents, err := io.ReadAll(actualFileReader)
			require.NoError(t, err)
			actualContents = RemoveYAMLComments(expectedContents)

			require.Equal(t, string(expectedContents), string(actualContents), fmt.Sprintf("%s contents do not match", expectedFile.Name))
		}
	}
}
