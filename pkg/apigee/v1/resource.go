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

package v1

import (
	"fmt"
	"path/filepath"
)

type Resource struct {
	Type    string `xml:"Type"`
	Path    string `xml:"Path"`
	Content []byte `xml:"-"`

	UnknownNode AnyList `xml:",any"`
}

func (p *Resource) FileContents() ([]byte, error) {
	return p.Content, nil
}

func (p *Resource) FileName() string {
	return filepath.Base(p.Path)
}

func (p *Resource) FilePath() string {
	return filepath.Join("resources", p.Type, p.FileName())
}

func ValidateResource(v *Resource, path string) []error {
	if v == nil {
		return nil
	}

	subPath := fmt.Sprintf("%s.Resource", path)
	if len(v.UnknownNode) > 0 {
		return []error{&UnknownNodeError{subPath, v.UnknownNode[0]}}
	}

	return nil
}
