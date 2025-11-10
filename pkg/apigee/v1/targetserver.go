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

import "fmt"

type Server struct {
	Name       string `xml:"name,attr"`
	IsFallback bool   `xml:"IsFallback,omitempty"`
	Weight     int    `xml:"Weight,omitempty"`

	UnknownNode AnyList `xml:",any"`
}

func ValidateServer(v *Server, path string) []error {
	if v == nil {
		return nil
	}

	subPath := fmt.Sprintf("%s.Server", path)
	if len(v.UnknownNode) > 0 {
		return []error{NewUnknownNodeError(subPath, v.UnknownNode[0])}
	}

	return nil
}
