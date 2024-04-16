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

package flags

import (
	"fmt"
	"github.com/go-errors/errors"
	"slices"
	"strings"
)

type Enum struct {
	Value   string
	Allowed []string
}

func NewEnum(allowed []string) Enum {
	return Enum{"", allowed}
}

func (f *Enum) String() string {
	return fmt.Sprintf("%s", f.Value)
}

func (f *Enum) Set(input string) error {
	input = strings.TrimSpace(input)

	index := slices.Index(f.Allowed, input)
	if index < 0 {
		return errors.Errorf("flag only allows %v", f.Allowed)
	}

	f.Value = input
	return nil
}

func (f *Enum) IsUnset() bool {
	return f == nil || strings.TrimSpace(string(f.Value)) == ""
}
