// Copyright 2024 Google LLC
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
	"os"
	"path/filepath"
)

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
