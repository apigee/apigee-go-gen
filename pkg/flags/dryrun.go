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
	"strings"
)

type DryRun string

func (f *DryRun) String() string {
	return fmt.Sprintf("%s", string(*f))
}

func (f *DryRun) Set(input string) error {
	input = strings.TrimSpace(input)
	if input != "xml" && input != "yaml" {
		return errors.Errorf("flag only allows \"xml\" or \"yaml\"")
	}

	*f = DryRun(input)
	return nil
}

func (f *DryRun) IsXML() bool {
	return f != nil && *f == "xml"
}

func (f *DryRun) IsYAML() bool {
	return f != nil && *f == "yaml"
}

func (f *DryRun) IsUnset() bool {
	return f == nil || strings.TrimSpace(string(*f)) == ""
}
