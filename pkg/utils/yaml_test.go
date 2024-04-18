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

package utils

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"regexp"
	"testing"
)

func TestYAMLText2XMLText(t *testing.T) {
	tests := []struct {
		dir string
	}{
		{
			"simple_nested",
		},
		{
			"scalar_without_attrs",
		},
		{
			"element_with_attr",
		},
		{
			"scalar_with_attrs",
		},
		{
			"sequence_parent_without_attrs",
		},
		{
			"sequence_parent_with_attrs",
		},
		{
			"sequence_without_parent",
		},
		{
			"sequence_without_parent_with_attrs",
		},
		{
			"unique_children_with_attrs_parent_without_attrs",
		},
		{
			"unique_children_without_attrs_parent_without_attrs",
		},
		{
			"repeated_children_without_attrs_parent_without_attrs",
		},
		{
			"complex_raise_fault_policy",
		},
		{
			"flow_callout_policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			dir := filepath.Join("testdata", "snippets", tt.dir)
			inFile := filepath.Join(dir, "data.yaml")
			inBytes := MustReadFileBytes(inFile)

			wantFile := filepath.Join(dir, "data.xml")
			wantBytes := MustReadFileBytes(wantFile)

			gotBytes, err := YAMLText2XMLText(bytes.NewReader(inBytes))
			assert.NoError(t, err)
			wantBytes = RemoveXMLComments(wantBytes)
			assert.Equal(t, string(wantBytes), string(gotBytes))
		})
	}
}

func RemoveXMLComments(data []byte) []byte {
	regex := regexp.MustCompile(`(?ms)^\s*<!--.*-->\s*[\r\n]?`)
	replaced := regex.ReplaceAll(data, []byte{})
	return replaced
}

func TestSplitJSONRef(t *testing.T) {
	type args struct {
		refStr string
	}
	tests := []struct {
		name         string
		args         args
		wantLocation string
		wantJsonPath string
		wantErr      bool
	}{
		{
			"absolute file path",
			args{
				"/hello/world.yaml",
			},
			"/hello/world.yaml",
			"$",
			false,
		},
		{
			"relative file path",
			args{
				"../relative.yaml",
			},
			"../relative.yaml",
			"$",
			false,
		},
		{
			"file with fragment",
			args{
				"hello.yaml#/foo",
			},
			"hello.yaml",
			"$.foo",
			false,
		},
		{
			"no path no fragment",
			args{
				"",
			},
			"",
			"$",
			false,
		},
		{
			"deep fragment only",
			args{
				"#/foo/bar/fizz",
			},
			"",
			"$.foo.bar.fizz",
			false,
		},
		{
			"root fragment #",
			args{
				"#",
			},
			"",
			"$",
			false,
		},
		{
			"root fragment #/",
			args{
				"#/",
			},
			"",
			"$",
			false,
		},
		{
			"file with deep pointer",
			args{
				"hello.yaml#/foo/bar/fizz",
			},
			"hello.yaml",
			"$.foo.bar.fizz",
			false,
		},
		{
			"remote JSONRef",
			args{
				"https://example.com/hello.yaml#/foo/bar/fizz",
			},
			"",
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLocation, gotJsonPath, err := SplitJSONRef(tt.args.refStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitJSONRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLocation != tt.wantLocation {
				t.Errorf("SplitJSONRef() gotLocation = %v, want %v", gotLocation, tt.wantLocation)
			}
			if gotJsonPath != tt.wantJsonPath {
				t.Errorf("SplitJSONRef() gotJsonPath = %v, want %v", gotJsonPath, tt.wantJsonPath)
			}
		})
	}
}
