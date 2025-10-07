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

package v1

import (
	"encoding/xml"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/go-errors/errors"
	"gopkg.in/yaml.v3"
	"path/filepath"
)

func NewAPIProxyModel(input string) (*APIProxyModel, error) {
	proxyModel := &APIProxyModel{}
	err := proxyModel.Hydrate(input)
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

func (a *APIProxyModel) Name() string {
	return a.APIProxy.Name
}

func (a *APIProxyModel) DisplayName() string {
	return a.APIProxy.DisplayName
}

func (a *APIProxyModel) Revision() int {
	return a.APIProxy.Revision
}

func (a *APIProxyModel) GetResources() *Resources {
	return &a.Resources
}

func (a *APIProxyModel) Validate() error {
	if a == nil {
		return nil
	}

	err := utils.MultiError{Errors: []error{}}
	path := "Root"
	if len(a.UnknownNode) > 0 {
		err.Errors = append(err.Errors, NewUnknownNodeError(path, a.UnknownNode[0]))
		return err
	}

	var subErrors []error
	subErrors = append(subErrors, ValidateAPIProxy(&a.APIProxy, path)...)
	subErrors = append(subErrors, ValidateProxyEndpoints(&a.ProxyEndpoints, path)...)
	subErrors = append(subErrors, ValidateTargetEndpoints(&a.TargetEndpoints, path)...)
	subErrors = append(subErrors, ValidateResources(&a.Resources, path)...)

	if len(subErrors) > 0 {
		err.Errors = append(err.Errors, subErrors...)
		return err
	}

	return nil
}

func (a *APIProxyModel) BundleRoot() string {
	return "apiproxy"
}

func (a *APIProxyModel) BundleFiles() []BundleFile {
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

func (a *APIProxyModel) Hydrate(filePath string) error {
	var err error

	a.YAMLDoc, err = utils.YAMLFile2YAML(filePath)

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

	err = HydrateResources(a, filepath.Dir(filePath))
	if err != nil {
		return errors.New(err)
	}

	return nil
}
