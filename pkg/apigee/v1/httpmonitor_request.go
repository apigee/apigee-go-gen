//  Copyright 2025 Google LLC
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

import "fmt"

type HttpMonitorRequest struct {
	ConnectTimeoutInSec        int         `xml:"ConnectTimeoutInSec,omitempty"`
	SocketReadTimeoutInSec     int         `xml:"SocketReadTimeoutInSec,omitempty"`
	Port                       int         `xml:"Port,omitempty"`
	Verb                       string      `xml:"Verb,omitempty"`
	Path                       string      `xml:"Path,omitempty"`
	UseTargetServerSSLInfo     bool        `xml:"UseTargetServerSSLInfo,omitempty"`
	IncludeHealthCheckIdHeader bool        `xml:"IncludeHealthCheckIdHeader,omitempty"`
	Payload                    string      `xml:"Payload,omitempty"`
	Headers                    HeadersList `xml:"Header,omitempty"`
	IsSSL                      bool        `xml:"IsSSL,omitempty"`
	TrustAllSSL                bool        `xml:"TrustAllSSL,omitempty"`

	UnknownNode AnyList `xml:",any"`
}

func ValidateHttpMonitorRequest(v *HttpMonitorRequest, path string) []error {
	if v == nil {
		return nil
	}

	subPath := fmt.Sprintf("%s.Request", path)
	if len(v.UnknownNode) > 0 {
		return []error{NewUnknownNodeError(subPath, v.UnknownNode[0])}
	}

	var subErrors []error
	subErrors = append(subErrors, ValidateHeaders(&v.Headers, subPath)...)

	return subErrors
}
