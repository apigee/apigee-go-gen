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

type ProxyEndpoint struct {
	Name                string               `xml:"name,attr"`
	Description         string               `xml:"Description,omitempty"`
	FaultRules          *FaultRules          `xml:"FaultRules"`
	DefaultFaultRule    *DefaultFaultRule    `xml:"DefaultFaultRule,omitempty"`
	PreFlow             *PreFlow             `xml:"PreFlow,omitempty"`
	Flows               *Flows               `xml:"Flows,omitempty"`
	PostFlow            *PostFlow            `xml:"PostFlow,omitempty"`
	HTTPProxyConnection *HTTPProxyConnection `xml:"HTTPProxyConnection,omitempty"`
	RouteRules          *RouteRuleList       `xml:"RouteRule,omitempty"`

	UnknownNode AnyList `xml:",any"`
}

func (p *ProxyEndpoint) XML() ([]byte, error) {
	return utils.Struct2XMLDocText(p)
}

func (p *ProxyEndpoint) FileContents() ([]byte, error) {
	return p.XML()
}

func (p *ProxyEndpoint) FileName() string {
	return fmt.Sprintf("%s.xml", p.Name)
}

func (p *ProxyEndpoint) FilePath() string {
	return filepath.Join("apiproxy", "proxies", p.FileName())
}

func ValidateProxyEndpoint(v *ProxyEndpoint, path string) []error {
	if v == nil {
		return nil
	}

	subPath := fmt.Sprintf("%s.ProxyEndpoint(name: %s)", path, v.Name)
	if len(v.UnknownNode) > 0 {
		return []error{&UnknownNodeError{subPath, v.UnknownNode[0]}}
	}

	var subErrors []error
	subErrors = append(subErrors, ValidateHTTPProxyConnection(v.HTTPProxyConnection, subPath)...)
	subErrors = append(subErrors, ValidatePreFlow(v.PreFlow, subPath)...)
	subErrors = append(subErrors, ValidateFlows(v.Flows, subPath)...)
	subErrors = append(subErrors, ValidatePostFlow(v.PostFlow, subPath)...)
	subErrors = append(subErrors, ValidateRouteRules(v.RouteRules, subPath)...)
	subErrors = append(subErrors, ValidateFaultRules(v.FaultRules, subPath)...)
	subErrors = append(subErrors, ValidateDefaultFaultRule(v.DefaultFaultRule, subPath)...)

	return subErrors
}
