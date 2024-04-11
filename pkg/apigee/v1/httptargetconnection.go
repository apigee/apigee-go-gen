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

import "fmt"

type HTTPTargetConnection struct {
	URL            string          `xml:"URL,omitempty"`
	Path           string          `xml:"Path,omitempty"`
	Authentication *Authentication `xml:"Authentication,omitempty"`
	LoadBalancer   *LoadBalancer   `xml:"LoadBalancer,omitempty"`
	SSLInfo        *SSLInfo        `xml:"SSLInfo,omitempty"`
	Properties     *Properties     `xml:"Properties"`

	UnknownNode AnyList `xml:",any"`
}

func ValidateHTTPTargetConnection(v *HTTPTargetConnection, path string) []error {
	if v == nil {
		return nil
	}

	subPath := fmt.Sprintf("%s.HTTPTargetConnection", path)
	if len(v.UnknownNode) > 0 {
		return []error{&UnknownNodeError{subPath, v.UnknownNode[0]}}
	}

	var subErrors []error
	subErrors = append(subErrors, ValidateProperties(v.Properties, subPath)...)
	subErrors = append(subErrors, ValidateLoadBalancer(v.LoadBalancer, subPath)...)
	subErrors = append(subErrors, ValidateSSLInfo(v.SSLInfo, subPath)...)
	subErrors = append(subErrors, ValidateAuthentication(v.Authentication, subPath)...)

	return subErrors
}
