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

package v1

import (
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/apigee/apigee-go-gen/pkg/zip"
	"github.com/go-errors/errors"
	"net/url"
	"os"
	"path/filepath"
)

type Model interface {
	Name() string
	DisplayName() string
	Revision() int
	Validate() error
	BundleFiles() []BundleFile
	BundleRoot() string
	Hydrate(filePath string) error
	XML() ([]byte, error)
	YAML() ([]byte, error)
	GetResources() *Resources
}

func Model2Bundle(model Model, output string) error {
	extension := filepath.Ext(output)
	if extension == ".zip" {
		err := Model2BundleZip(model, output)
		if err != nil {
			return err
		}
	} else if extension != "" {
		return errors.Errorf("output extension %s is not supported", extension)
	} else {
		err := Model2BundleDir(model, output)
		if err != nil {
			return err
		}
		//directory
	}

	return nil
}

func Model2BundleDir(model Model, output string) error {
	err := os.MkdirAll(output, os.ModePerm)
	if err != nil {
		return errors.New(err)
	}

	bundleFiles := model.BundleFiles()
	for _, bundleFile := range bundleFiles {
		filePath := filepath.Join(model.BundleRoot(), bundleFile.FilePath())
		fileDir := filepath.Dir(filePath)

		dirDiskPath := filepath.Join(output, fileDir)
		err := os.MkdirAll(dirDiskPath, os.ModePerm)
		if err != nil {
			return errors.New(err)
		}

		fileContent, err := bundleFile.FileContents()
		if err != nil {
			return err
		}

		fileDiskPath := filepath.Join(output, filePath)
		err = os.WriteFile(fileDiskPath, fileContent, os.ModePerm)
		if err != nil {
			return errors.New(err)
		}
	}

	return nil
}

func Model2BundleZip(model Model, outputZip string) error {
	tmpDir, err := os.MkdirTemp("", "unzipped-bundle-*")
	if err != nil {
		return errors.New(err)
	}

	err = Model2BundleDir(model, tmpDir)
	if err != nil {
		return err
	}

	err = zip.Zip(outputZip, tmpDir)
	if err != nil {
		return err
	}

	return nil
}

func HydrateResources(model Model, fromDir string) error {
	// switch to directory relative to the YAML file so that resource paths are valid
	wd, err := os.Getwd()
	if err != nil {
		return errors.New(err)
	}
	defer func() { utils.Must(os.Chdir(wd)) }()

	err = os.Chdir(fromDir)
	if err != nil {
		return errors.New(err)
	}

	for _, resource := range model.GetResources().List {
		parsedUrl, err := url.Parse(resource.Path)
		if err != nil {
			return errors.New(err)
		}

		content, err := os.ReadFile(parsedUrl.Path)
		if err != nil {
			return errors.New(err)
		}
		resource.Content = content
	}
	return nil
}
