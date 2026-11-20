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
	"fmt"
	"github.com/go-errors/errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// CopyDir copies the contents of src into dest.
//
// Symbolic links are not recreated in dest. A link that resolves to a file or
// directory inside src is replaced by a copy of what it points to. A link that
// resolves to a location outside of src is an error, so that neither the copy
// nor anything later written into dest can reach outside of it. Dangling links,
// and links that point back to one of their own parent directories, are skipped.
func CopyDir(dest string, src string) error {
	realRoot, err := filepath.EvalSymlinks(src)
	if err != nil {
		return errors.New(err)
	}
	if realRoot, err = filepath.Abs(realRoot); err != nil {
		return errors.New(err)
	}
	return copyDir(dest, src, realRoot, nil)
}

func copyDir(dest string, src string, realRoot string, parents []string) error {
	realSrc, err := filepath.EvalSymlinks(src)
	if err != nil {
		return errors.New(err)
	}
	if realSrc, err = filepath.Abs(realSrc); err != nil {
		return errors.New(err)
	}
	if slices.Contains(parents, realSrc) {
		//symlink cycle, skip it
		return nil
	}
	parents = append(parents, realSrc)

	if err = os.MkdirAll(dest, os.ModePerm); err != nil {
		return errors.New(err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return errors.New(err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.Type()&os.ModeSymlink != 0 {
			target, err := filepath.EvalSymlinks(srcPath)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					//skip dangling symlinks
					continue
				}
				return errors.New(err)
			}
			if target, err = filepath.Abs(target); err != nil {
				return errors.New(err)
			}
			if !isWithinDir(realRoot, target) {
				return errors.Errorf("symlink %q points outside of %q", srcPath, realRoot)
			}
		}

		//os.Stat follows symlinks, so info describes the link target
		info, err := os.Stat(srcPath)
		if err != nil {
			return errors.New(err)
		}

		if info.IsDir() {
			if err = copyDir(destPath, srcPath, realRoot, parents); err != nil {
				return err
			}
			continue
		}

		if !info.Mode().IsRegular() {
			//skip other unknown file types
			continue
		}

		if err = copyRegularFile(destPath, srcPath); err != nil {
			return err
		}
	}
	return nil
}

func isWithinDir(dir string, path string) bool {
	rel, err := filepath.Rel(dir, path)
	return err == nil && filepath.IsLocal(rel)
}

func copyRegularFile(dest string, src string) error {
	r, err := os.Open(src)
	if err != nil {
		return errors.New(err)
	}
	defer r.Close()

	w, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return errors.New(err)
	}

	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		return &os.PathError{Op: "Copy", Path: dest, Err: err}
	}
	if err = w.Close(); err != nil {
		return errors.New(err)
	}
	return nil
}

// WriteFileInDir writes data to name, a path relative to dir.
//
// The write fails if name, or any symbolic link along the way, resolves to a
// location outside of dir. When createDirs is true, dir and any missing parent
// directories of name are created (the latter also confined to dir).
func WriteFileInDir(dir string, name string, data []byte, createDirs bool) error {
	if createDirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return errors.New(err)
		}
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return errors.New(err)
	}
	defer root.Close()

	if createDirs {
		parent := ""
		for _, part := range strings.Split(filepath.ToSlash(filepath.Dir(filepath.Clean(name))), "/") {
			if part == "" || part == "." {
				continue
			}
			parent = filepath.Join(parent, part)
			if err = root.Mkdir(parent, os.ModePerm); err != nil && !errors.Is(err, fs.ErrExist) {
				return errors.New(err)
			}
		}
	}

	f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return errors.New(err)
	}
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		return errors.New(err)
	}
	if err = f.Close(); err != nil {
		return errors.New(err)
	}
	return nil
}

func CopyFile(dest string, src string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.New(err)
	}
	defer func() { MustClose(srcFile) }()

	destDir := filepath.Dir(dest)
	err = os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return err
	}

	destFile, err := os.Create(dest)
	if err != nil {
		return errors.New(err)
	}
	defer func() { MustClose(destFile) }()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return errors.New(err)
	}

	err = destFile.Sync()
	if err != nil {
		return errors.New(err)
	}
	return nil
}

func MustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		panic(err)
	}
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func WriteOutputText(output string, outputText []byte) error {
	var err error

	if output == "-" || len(output) == 0 {
		fmt.Printf("%s", string(outputText))
		return nil
	}

	dir := filepath.Dir(output)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return errors.New(err)
	}

	err = os.WriteFile(output, outputText, os.ModePerm)
	if err != nil {
		return errors.New(err)
	}
	return nil
}

func ReadInputText(input string) ([]byte, error) {
	var text []byte
	var err error
	if input == "-" || len(input) == 0 {
		text, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, errors.New(err)
		}
	} else {
		text, err = os.ReadFile(input)
		if err != nil {
			return nil, errors.New(err)
		}
	}
	return text, nil
}

func ReadInputTextFile(input string) ([]byte, error) {
	var text []byte
	var err error
	text, err = os.ReadFile(input)
	if err != nil {
		return nil, errors.New(err)
	}
	return text, nil
}
