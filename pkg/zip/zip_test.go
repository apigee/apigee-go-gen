//  Copyright 2026 Google LLC
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

package zip

import (
	"archive/zip"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func writeTestZip(t *testing.T, zipPath string, entries map[string]string) {
	t.Helper()
	f, err := os.Create(zipPath)
	require.NoError(t, err)
	w := zip.NewWriter(f)
	for name, content := range entries {
		entry, err := w.Create(name)
		require.NoError(t, err)
		_, err = entry.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())
}

func TestUnzip(t *testing.T) {
	baseDir := t.TempDir()
	srcZip := filepath.Join(baseDir, "bundle.zip")
	writeTestZip(t, srcZip, map[string]string{
		"apiproxy/":                     "",
		"apiproxy/proxy.xml":            "<APIProxy/>",
		"apiproxy/policies/AM-Test.xml": "<AssignMessage/>",
	})

	destDir := filepath.Join(baseDir, "out")
	require.NoError(t, Unzip(destDir, srcZip))

	require.FileExists(t, filepath.Join(destDir, "apiproxy", "proxy.xml"))
	content, err := os.ReadFile(filepath.Join(destDir, "apiproxy", "policies", "AM-Test.xml"))
	require.NoError(t, err)
	require.Equal(t, "<AssignMessage/>", string(content))
}

func TestUnzipRejectsNonLocalEntries(t *testing.T) {
	tests := []struct {
		name  string
		entry string
	}{
		{"parent-traversal", "../../escaped.txt"},
		{"nested-traversal", "apiproxy/../../escaped.txt"},
		{"absolute", "/tmp/escaped.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := t.TempDir()
			srcZip := filepath.Join(baseDir, "evil.zip")
			writeTestZip(t, srcZip, map[string]string{tt.entry: "pwned"})

			destDir := filepath.Join(baseDir, "a", "b", "extract")
			require.NoError(t, os.MkdirAll(destDir, os.ModePerm))

			err := Unzip(destDir, srcZip)
			require.Error(t, err)
			require.Contains(t, err.Error(), "is not a local path")

			require.NoFileExists(t, filepath.Join(baseDir, "escaped.txt"))
			require.NoFileExists(t, filepath.Join(baseDir, "a", "escaped.txt"))
		})
	}
}
