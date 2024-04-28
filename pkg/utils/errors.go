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

package utils

import (
	"fmt"
)

func NewCyclicJSONRefError(paths []string) error {
	return CyclicJSONRefError{
		Paths: paths,
	}
}

type CyclicJSONRefError struct {
	Paths []string
}

func (e CyclicJSONRefError) Error() string {
	return fmt.Sprintf("cyclic JSONRef at %s", e.Paths[len(e.Paths)-1])
}
