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

import (
	"fmt"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/gosimple/slug"
	"path/filepath"
)

type SharedFlow struct {
	Name  string   `xml:"name,attr"`
	Steps StepList `xml:"Step"`

	UnknownNode AnyList `xml:",any"`
}

func (a *SharedFlow) FileContents() ([]byte, error) {
	return a.XML()
}

func (a *SharedFlow) FileName() string {
	return fmt.Sprintf("%s.xml", slug.Make(a.Name))
}

func (a *SharedFlow) FilePath() string {
	return filepath.Join("sharedflows", a.FileName())
}

func (a *SharedFlow) XML() ([]byte, error) {
	return utils.Struct2XMLDocText(a)
}

func ValidateSharedFlow(v *SharedFlow, path string) []error {
	if v == nil {
		return nil
	}

	subPath := fmt.Sprintf("%s.SharedFlow", path)
	if len(v.UnknownNode) > 0 {
		return []error{&UnknownNodeError{subPath, v.UnknownNode[0]}}
	}

	var subErrors []error
	subErrors = append(subErrors, ValidateSteps(&v.Steps, subPath)...)

	return nil
}
