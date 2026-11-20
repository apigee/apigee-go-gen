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

package utils

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func mustWrite(t *testing.T, path string, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), os.ModePerm))
	require.NoError(t, os.WriteFile(path, []byte(content), os.ModePerm))
}

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "apiproxy.yaml"), "main")
	mustWrite(t, filepath.Join(src, "resources", "jsc", "script.js"), "js")
	require.NoError(t, os.Symlink("apiproxy.yaml", filepath.Join(src, "link-to-file.yaml")))
	require.NoError(t, os.Symlink("resources", filepath.Join(src, "link-to-dir")))
	require.NoError(t, os.Symlink("does-not-exist", filepath.Join(src, "dangling")))
	require.NoError(t, os.Symlink("..", filepath.Join(src, "resources", "cycle")))

	dest := t.TempDir()
	require.NoError(t, CopyDir(dest, src))

	readDest := func(rel string) string {
		content, err := os.ReadFile(filepath.Join(dest, rel))
		require.NoError(t, err)
		return string(content)
	}
	require.Equal(t, "main", readDest("apiproxy.yaml"))
	require.Equal(t, "js", readDest("resources/jsc/script.js"))

	//symlinks inside src are replaced by copies of their targets
	require.Equal(t, "main", readDest("link-to-file.yaml"))
	require.Equal(t, "js", readDest("link-to-dir/jsc/script.js"))
	for _, rel := range []string{"link-to-file.yaml", "link-to-dir"} {
		info, err := os.Lstat(filepath.Join(dest, rel))
		require.NoError(t, err)
		require.Zero(t, info.Mode()&os.ModeSymlink, "%s must not be a symlink", rel)
	}

	//dangling and cyclic links are skipped
	require.NoFileExists(t, filepath.Join(dest, "dangling"))
	require.NoDirExists(t, filepath.Join(dest, "resources", "cycle"))
}

func TestCopyDirRejectsSymlinksOutsideSrc(t *testing.T) {
	base := t.TempDir()
	outside := filepath.Join(base, "outside")
	mustWrite(t, filepath.Join(outside, "secret.txt"), "secret")

	tests := []struct {
		name   string
		target string
	}{
		{"absolute-dir", outside},
		{"absolute-file", filepath.Join(outside, "secret.txt")},
		{"relative-dir", "../outside"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := filepath.Join(base, "src-"+tt.name)
			mustWrite(t, filepath.Join(src, "apiproxy.yaml"), "main")
			require.NoError(t, os.Symlink(tt.target, filepath.Join(src, "escape")))

			dest := t.TempDir()
			err := CopyDir(dest, src)
			require.Error(t, err)
			require.Contains(t, err.Error(), "points outside of")

			_, err = os.Lstat(filepath.Join(dest, "escape"))
			require.True(t, os.IsNotExist(err))
		})
	}
}

func TestWriteFileInDir(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, WriteFileInDir(dir, "./a.txt", []byte("a"), false))
	require.NoError(t, WriteFileInDir(dir, "nested/deep/b.txt", []byte("b"), true))
	require.Error(t, WriteFileInDir(dir, "missing/c.txt", []byte("c"), false))

	//overwrites truncate existing content
	require.NoError(t, WriteFileInDir(dir, "a.txt", []byte("x"), false))

	content, err := os.ReadFile(filepath.Join(dir, "a.txt"))
	require.NoError(t, err)
	require.Equal(t, "x", string(content))
	content, err = os.ReadFile(filepath.Join(dir, "nested", "deep", "b.txt"))
	require.NoError(t, err)
	require.Equal(t, "b", string(content))
}

func TestWriteFileInDirRejectsEscapes(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "out")
	outside := filepath.Join(base, "outside")
	require.NoError(t, os.MkdirAll(dir, os.ModePerm))
	require.NoError(t, os.MkdirAll(outside, os.ModePerm))
	require.NoError(t, os.Symlink(outside, filepath.Join(dir, "abs-link")))
	require.NoError(t, os.Symlink("../outside", filepath.Join(dir, "rel-link")))

	for _, name := range []string{"abs-link/pwned", "rel-link/pwned", "../outside/pwned", "abs-link/sub/pwned"} {
		t.Run(name, func(t *testing.T) {
			require.Error(t, WriteFileInDir(dir, name, []byte("pwned"), true))
			require.NoFileExists(t, filepath.Join(outside, "pwned"))
			require.NoDirExists(t, filepath.Join(outside, "sub"))
		})
	}
}
