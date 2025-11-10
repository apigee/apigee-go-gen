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
	"path/filepath"
)

type IntegrationEndpoint struct {
	Name           string `xml:"name,attr"`
	AsyncExecution bool   `xml:"AsyncExecution"`

	UnknownNode AnyList `xml:",any"`
}

func (p *IntegrationEndpoint) XML() ([]byte, error) {
	return utils.Struct2XMLDocText(p)
}

func (p *IntegrationEndpoint) FileContents() ([]byte, error) {
	return p.XML()
}

func (p *IntegrationEndpoint) FileName() string {
	return fmt.Sprintf("%s.xml", p.Name)
}

func (p *IntegrationEndpoint) FilePath() string {
	return filepath.Join("integration-endpoints", p.FileName())
}

func ValidateIntegrationEndpoint(v *IntegrationEndpoint, path string) []error {
	if v == nil {
		return nil
	}

	subPath := fmt.Sprintf("%s.IntegrationEndpoint(name: %s)", path, v.Name)
	if len(v.UnknownNode) > 0 {
		return []error{NewUnknownNodeError(subPath, v.UnknownNode[0])}
	}

	var subErrors []error
	return subErrors
}
