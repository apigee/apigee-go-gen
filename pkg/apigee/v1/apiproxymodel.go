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

package v1

import (
	"encoding/xml"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"github.com/micovery/apigee-yaml-toolkit/pkg/zip"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"path/filepath"
)

func NewAPIProxyModel(input string) (*APIProxyModel, error) {
	proxyModel := &APIProxyModel{}
	err := proxyModel.HydrateModelFromYAMLDoc(input)
	if err != nil {
		return nil, err
	}
	return proxyModel, nil
}

type APIProxyModel struct {
	APIProxy        APIProxy        `xml:"APIProxy"`
	Policies        Policies        `xml:"Policies"`
	ProxyEndpoints  ProxyEndpoints  `xml:"ProxyEndpoints"`
	TargetEndpoints TargetEndpoints `xml:"TargetEndpoints"`
	Resources       Resources       `xml:"Resources"`

	UnknownNode AnyList `xml:",any"`

	YAMLDoc *yaml.Node `xml:"-"`
}

func ValidateAPIProxyModel(v *APIProxyModel) []error {
	if v == nil {
		return nil
	}

	path := "Root"
	if len(v.UnknownNode) > 0 {
		return []error{&UnknownNodeError{path, v.UnknownNode[0]}}
	}

	var subErrors []error
	subErrors = append(subErrors, ValidateAPIProxy(&v.APIProxy, path)...)
	subErrors = append(subErrors, ValidateProxyEndpoints(&v.ProxyEndpoints, path)...)
	subErrors = append(subErrors, ValidateTargetEndpoints(&v.TargetEndpoints, path)...)
	subErrors = append(subErrors, ValidateResources(&v.Resources, path)...)

	return subErrors
}

type BundleFile interface {
	FileContents() ([]byte, error)
	FileName() string
	FilePath() string
}

func (a *APIProxyModel) GetBundleFiles() []BundleFile {
	bundleFiles := []BundleFile{
		BundleFile(&a.APIProxy),
	}

	for _, item := range a.Policies.List {
		bundleFiles = append(bundleFiles, BundleFile(item))
	}

	for _, item := range a.ProxyEndpoints.List {
		bundleFiles = append(bundleFiles, BundleFile(item))
	}

	for _, item := range a.TargetEndpoints.List {
		bundleFiles = append(bundleFiles, BundleFile(item))
	}

	for _, item := range a.Resources.List {
		bundleFiles = append(bundleFiles, BundleFile(item))
	}

	return bundleFiles
}

func (a *APIProxyModel) XML() ([]byte, error) {
	return utils.Struct2XMLDocText(a)
}

func (a *APIProxyModel) YAML() ([]byte, error) {
	return utils.YAML2Text(a.YAMLDoc, 2)
}

func (a *APIProxyModel) HydrateModelFromYAMLDoc(filePath string) error {
	var err error

	a.YAMLDoc, err = ReadAPIProxyModelFromYAML(filePath)

	if err != nil {
		return err
	}

	wrapper := &yaml.Node{Kind: yaml.MappingNode}
	wrapper.Content = append(wrapper.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: "APIProxyModel"}, a.YAMLDoc)

	xmlText, err := utils.YAML2XMLText(wrapper)
	if err != nil {
		return err
	}

	if err = xml.Unmarshal(xmlText, a); err != nil {
		return errors.New(err)
	}

	err = LoadAPIProxyModelResources(a, filepath.Dir(filePath))
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func ReadAPIProxyModelFromYAML(filePath string) (*yaml.Node, error) {
	var file *os.File
	var err error
	if file, err = os.Open(filePath); err != nil {
		return nil, errors.New(err)
	}
	defer file.Close()

	//switch to directory relative to the YAML file so that JSON $refs are valid
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.New(err)
	}
	defer os.Chdir(wd)

	err = os.Chdir(filepath.Dir(filePath))
	if err != nil {
		return nil, errors.New(err)
	}

	dataNode, err := utils.Text2YAML(file)
	if err != nil {
		return nil, err
	}

	return dataNode, nil
}

func LoadAPIProxyModelResources(proxyModel *APIProxyModel, fromDir string) error {
	// switch to directory relative to the YAML file so that resource paths are valid
	wd, err := os.Getwd()
	if err != nil {
		return errors.New(err)
	}
	defer os.Chdir(wd)

	err = os.Chdir(fromDir)
	if err != nil {
		return errors.New(err)
	}

	for _, resource := range proxyModel.Resources.List {
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

func APIProxyModel2Bundle(proxyModel *APIProxyModel, output string) error {
	extension := filepath.Ext(output)
	if extension == ".zip" {
		err := APIProxyModel2BundleZip(proxyModel, output)
		if err != nil {
			return err
		}
	} else if extension != "" {
		return errors.Errorf("output extension %s is not supported", extension)
	} else {
		err := APIProxyModel2BundleDir(proxyModel, output)
		if err != nil {
			return err
		}
		//directory
	}

	return nil
}

func APIProxyModel2BundleDir(proxyModel *APIProxyModel, output string) error {

	err := os.MkdirAll(output, os.ModePerm)
	if err != nil {
		return errors.New(err)
	}

	bundleFiles := proxyModel.GetBundleFiles()
	for _, bundleFile := range bundleFiles {
		filePath := bundleFile.FilePath()
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

func APIProxyModel2BundleZip(proxyModel *APIProxyModel, outputZip string) error {
	tmpDir, err := os.MkdirTemp("", "unzipped-bundle-*")
	if err != nil {
		return errors.New(err)
	}

	err = APIProxyModel2Bundle(proxyModel, tmpDir)
	if err != nil {
		return err
	}

	err = zip.Zip(outputZip, tmpDir)
	if err != nil {
		return err
	}

	return nil
}

func APIProxyModelYAML2Bundle(input string, output string, validate bool, dryRun string) (err error, validationErrs []error) {
	proxyModel, err := NewAPIProxyModel(input)
	if err != nil {
		return err, nil
	}

	if dryRun == "xml" {
		xmlText, err := proxyModel.XML()
		if err != nil {
			return err, nil
		}
		fmt.Println(string(xmlText))
	} else if dryRun == "yaml" {
		yamlText, err := proxyModel.YAML()
		if err != nil {
			return err, nil
		}
		fmt.Println(string(yamlText))
	}

	if validate {
		validationErrs = ValidateAPIProxyModel(proxyModel)
		if len(validationErrs) > 0 {
			return nil, validationErrs
		}
	}

	if dryRun != "" {
		return nil, nil
	}

	err = APIProxyModel2Bundle(proxyModel, output)
	if err != nil {
		return err, nil
	}

	return nil, nil
}
