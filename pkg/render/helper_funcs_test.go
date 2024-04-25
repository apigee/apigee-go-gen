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
	"github.com/micovery/apigee-go-gen/pkg/utils"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func Test_getOSCopyFileFunc(t *testing.T) {

	tests := []struct {
		name         string
		outputFile   string
		templateFile string
		src          string
		dst          string
	}{
		{
			"deepcopy",
			"./out/out.yaml",
			"./in.yaml",
			"./data.yaml",
			"./parent/child/data.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDataDir := filepath.Join("testdata", tt.name)

			//cleanup the output dir
			outputDir := filepath.Join(testDataDir, filepath.Dir(tt.outputFile))
			err := os.RemoveAll(outputDir)
			require.NoError(t, err)

			//set the correct path for the output and the template files
			templateFile := filepath.Join(testDataDir, tt.templateFile)
			outputFile := filepath.Join(testDataDir, tt.outputFile)

			//create instance of the copy function
			copyFunc := getOSCopyFileFunc(templateFile, outputFile, false)

			//invoke the actual copy function
			dstRes := copyFunc(tt.dst, tt.src)
			require.Equal(t, tt.dst, dstRes)

			//verify the directory structure actually got created on the destination
			require.DirExists(t, filepath.Join(filepath.Dir(outputFile), filepath.Dir(tt.dst)))

			expectedFile := filepath.Join(filepath.Dir(templateFile), tt.src)
			actualFile := filepath.Join(filepath.Dir(outputFile), tt.dst)

			//verify the file contents actually match
			expectedFileData := utils.MustReadFileBytes(expectedFile)
			actualFileData := utils.MustReadFileBytes(actualFile)
			require.Equal(t, string(expectedFileData), string(actualFileData))
		})
	}
}
