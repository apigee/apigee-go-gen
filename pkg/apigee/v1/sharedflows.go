//  Copyright 2024 Google LLC
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

type SharedFlowList []*SharedFlow
type SharedFlows struct {
	List SharedFlowList `xml:"SharedFlow,omitempty"`

	UnknownNode AnyList `xml:",any"`
}

func ValidateSharedFlows(v *SharedFlows, path string) []error {
	if v == nil {
		return nil
	}

	subPath := fmt.Sprintf("%s.SharedFlows", path)
	if len(v.UnknownNode) > 0 {
		return []error{&UnknownNodeError{subPath, v.UnknownNode[0]}}
	}

	for index, vv := range v.List {
		errs := ValidateSharedFlow(vv, fmt.Sprintf("%s.%v", subPath, index))
		if len(errs) > 0 {
			return errs
		}
	}

	return nil
}
