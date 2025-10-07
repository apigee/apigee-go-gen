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

type Authentication struct {
	GoogleIDToken     *GoogleIDToken     `xml:"GoogleIDToken"`
	GoogleAccessToken *GoogleAccessToken `xml:"GoogleAccessToken"`

	UnknownNode AnyList `xml:",any"`
}

func ValidateAuthentication(v *Authentication, path string) []error {
	if v == nil {
		return nil
	}

	subPath := fmt.Sprintf("%s.Authentication", path)
	if len(v.UnknownNode) > 0 {
		return []error{NewUnknownNodeError(subPath, v.UnknownNode[0])}
	}

	var subErrors []error
	subErrors = append(subErrors, ValidateGoogleIDToken(v.GoogleIDToken, subPath)...)
	subErrors = append(subErrors, ValidateGoogleAccessToken(v.GoogleAccessToken, subPath)...)

	return subErrors
}
