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

package utils

import (
	"fmt"
	"github.com/go-errors/errors"
	"regexp"
	"strconv"
	"strings"
)

type DryRunFlag string

func (f *DryRunFlag) String() string {
	return fmt.Sprintf("%s", string(*f))
}

func (f *DryRunFlag) Set(input string) error {
	input = strings.TrimSpace(input)
	if input != "xml" && input != "yaml" {
		return errors.Errorf("flag only allows \"xml\" or \"yaml\"")
	}

	*f = DryRunFlag(input)
	return nil
}

func (f *DryRunFlag) IsXML() bool {
	return f != nil && *f == "xml"
}

func (f *DryRunFlag) IsYAML() bool {
	return f != nil && *f == "yaml"
}

func (f *DryRunFlag) IsUnset() bool {
	return f == nil || strings.TrimSpace(string(*f)) == ""
}

type FlagBool bool

func (f *FlagBool) String() string {
	return fmt.Sprintf("%v", bool(*f))
}

func (f *FlagBool) Set(input string) error {

	value, err := strconv.ParseBool(input)
	if err != nil {
		return errors.New(err)
	}

	*f = FlagBool(value)
	return nil
}

type IncludeList []string

func (i *IncludeList) String() string {
	return strings.Join(([]string)(*i), ",")
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
