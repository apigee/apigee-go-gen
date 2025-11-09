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
)

func CopyDir(dest string, src string) error {
	fSys := os.DirFS(src)
	return fs.WalkDir(fSys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fpath, err := filepath.Localize(path)
		if err != nil {
			return errors.New(err)
		}
		newPath := filepath.Join(dest, fpath)
		if d.IsDir() {
			return os.MkdirAll(newPath, os.ModePerm)
		}

		if !d.Type().IsRegular() {
			if d.Type()&os.ModeSymlink == os.ModeSymlink {
				srcLink := filepath.Join(src, path)
				var oldName string
				if oldName, err = os.Readlink(srcLink); err != nil {
					return errors.New(err)
				}
				if err = os.Symlink(oldName, newPath); err != nil {
					return errors.New(err)
				}
				return nil
			}
			//skip other unknown file types
			return nil
		}

		r, err := fSys.Open(path)
		if err != nil {
			return errors.New(err)
		}
		defer r.Close()
		_, err = r.Stat()
		if err != nil {
			return errors.New(err)
		}
		w, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return errors.New(err)
		}

		if _, err := io.Copy(w, r); err != nil {
			_ = w.Close()
			return &os.PathError{Op: "Copy", Path: newPath, Err: err}
		}
		if err = w.Close(); err != nil {
			return errors.New(err)
		}
		return nil
	})
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
