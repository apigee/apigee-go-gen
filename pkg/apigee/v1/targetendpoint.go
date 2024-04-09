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
	"fmt"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"path/filepath"
)

type TargetEndpoint struct {
	Name                  string                 `xml:"name,attr"`
	Description           string                 `xml:"Description,omitempty"`
	FaultRules            *FaultRules            `xml:"FaultRules,omitempty"`
	DefaultFaultRule      *DefaultFaultRule      `xml:"DefaultFaultRule,omitempty"`
	PreFlow               PreFlow                `xml:"PreFlow"`
	Flows                 Flows                  `xml:"Flows"`
	PostFlow              PostFlow               `xml:"PostFlow"`
	HTTPTargetConnection  *HTTPTargetConnection  `xml:"HTTPTargetConnection,omitempty"`
	LocalTargetConnection *LocalTargetConnection `xml:"LocalTargetConnection,omitempty"`

	UnknownNode AnyList `xml:",any"`
}

func (p *TargetEndpoint) XML() ([]byte, error) {
	return utils.Struct2XMLDocText(p)
}

func (p *TargetEndpoint) FileContents() ([]byte, error) {
	return p.XML()
}

func (p *TargetEndpoint) FileName() string {
	return fmt.Sprintf("%s.xml", p.Name)
}

func (p *TargetEndpoint) FilePath() string {
	return filepath.Join("apiproxy", "targets", p.FileName())
}

func ValidateTargetEndpoint(v *TargetEndpoint, path string) []error {
	if v == nil {
		return nil
	}

	subPath := fmt.Sprintf("%s.TargetEndpoint(name: %s)", path, v.Name)
	if len(v.UnknownNode) > 0 {
		return []error{&UnknownNodeError{subPath, v.UnknownNode[0]}}
	}

	var subErrors []error
	subErrors = append(subErrors, ValidateHTTPTargetConnection(v.HTTPTargetConnection, subPath)...)
	subErrors = append(subErrors, ValidatePreFlow(&v.PreFlow, subPath)...)
	subErrors = append(subErrors, ValidateFlows(&v.Flows, subPath)...)
	subErrors = append(subErrors, ValidatePostFlow(&v.PostFlow, subPath)...)
	subErrors = append(subErrors, ValidateFaultRules(v.FaultRules, subPath)...)
	subErrors = append(subErrors, ValidateDefaultFaultRule(v.DefaultFaultRule, subPath)...)
	subErrors = append(subErrors, ValidateLocalTargetConnection(v.LocalTargetConnection, subPath)...)

	return subErrors
}
