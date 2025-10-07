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
	"fmt"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/gosimple/slug"
)

type APIProxy struct {
	Revision       int    `xml:"revision,attr"`
	Name           string `xml:"name,attr"`
	DisplayName    string `xml:"DisplayName,omitempty"`
	Description    string `xml:"Description,omitempty"`
	CreatedAt      int64  `xml:"CreatedAt,omitempty"`
	LastModifiedAt int64  `xml:"LastModifiedAt,omitempty"`

	//-- deprecated fields
	Spec                 *Deprecated `xml:"Spec"`
	ConfigurationVersion *Deprecated `xml:"ConfigurationVersion"`
	BasePaths            *Deprecated `xml:"BasePaths"`
	Basepaths            *Deprecated `xml:"Basepaths"`
	Policies             *Deprecated `xml:"Policies"`
	Resources            *Deprecated `xml:"Resources"`
	ProxyEndpoints       *Deprecated `xml:"ProxyEndpoints"`
	TargetEndpoints      *Deprecated `xml:"TargetEndpoints"`

	UnknownNode AnyList `xml:",any"`
}

func ValidateAPIProxy(v *APIProxy, path string) []error {
	if v == nil {
		return nil
	}

	subPath := fmt.Sprintf("%s.APIProxy", path)
	if len(v.UnknownNode) > 0 {
		return []error{NewUnknownNodeError(subPath, v.UnknownNode[0])}
	}

	return nil
}

func (a *APIProxy) FileContents() ([]byte, error) {
	return a.XML()
}

func (a *APIProxy) FileName() string {
	return fmt.Sprintf("%s.xml", slug.Make(a.Name))
}

func (a *APIProxy) FilePath() string {
	return a.FileName()
}

func (a *APIProxy) XML() ([]byte, error) {
	return utils.Struct2XMLDocText(a)
}
