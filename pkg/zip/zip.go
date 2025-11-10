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

package zip

import (
	"archive/zip"
	"github.com/go-errors/errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func Unzip(destDir string, srcZip string) error {
	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return errors.New(err)
	}
	defer MustClose(r)

	unzipFile := func(zipFile *zip.File) error {
		path := filepath.Join(destDir, zipFile.Name)

		if zipFile.FileInfo().IsDir() {
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return errors.New(err)
			}
			return nil
		}

		srcFile, err := zipFile.Open()
		if err != nil {
			return errors.New(err)
		}
		defer MustClose(srcFile)

		//some zip files don't have directory entries, make one
		dir := filepath.Dir(path)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return errors.New(err)
		}

		destFile, err := os.Create(path)
		if err != nil {
			return errors.New(err)
		}
		defer MustClose(destFile)

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return errors.New(err)
		}

		return nil
	}

	for _, zipFile := range r.File {
		err := unzipFile(zipFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func Zip(destZip string, srcDir string) error {
	destDir := filepath.Dir(destZip)
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return errors.New(err)
	}

	outFile, err := os.Create(destZip)
	if err != nil {
		return errors.New(err)
	}
	defer MustClose(outFile)

	writer := zip.NewWriter(outFile)
	defer MustClose(writer)

	fSys := os.DirFS(srcDir)

	err = fs.WalkDir(fSys, ".", func(name string, dir fs.DirEntry, err error) error {
		if err != nil || name == "." {
			return err
		}

		if dir.IsDir() {
			_, err = writer.Create(name + "/")
			if err != nil {
				return errors.New(err)
			}
			return nil
		}

		dstFile, err := writer.Create(name)
		if err != nil {
			return err
		}

		srcFile, err := fSys.Open(name)
		if err != nil {
			return err
		}
		defer MustClose(srcFile)

		_, err = io.Copy(dstFile, srcFile)
		return err
	})

	if err != nil {
		return errors.New(err)
	}

	err = writer.Flush()
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
