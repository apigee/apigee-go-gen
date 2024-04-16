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
	"regexp"
	"strings"
)

type IncludeList []string

func NewIncludeList() IncludeList {
	return IncludeList{}
}

func (i *IncludeList) Type() string {
	return "string"
}

func (i *IncludeList) String() string {
	return ""
}

func (i *IncludeList) Set(input string) error {
	zp := regexp.MustCompile(`(\s+|,+)`) // spaces and one comma

	split := zp.Split(input, -1)
	for _, s := range split {
		if len(strings.TrimSpace(s)) == 0 {
			continue
		}
		*i = append(*i, s)
	}

	return nil
}
