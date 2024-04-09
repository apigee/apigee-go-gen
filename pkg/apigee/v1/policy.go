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
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"path/filepath"
)

type Policy AnyNode

func (p *Policy) Type() string {
	return p.XMLName.Local
}

func (p *Policy) Name() string {
	for _, attr := range p.Attrs {
		if attr.Name.Local == "name" {
			return attr.Value
		}
	}
	return ""
}

func (p *Policy) FileContents() ([]byte, error) {
	return p.XML()
}

func (p *Policy) FileName() string {
	return fmt.Sprintf("%s.xml", p.Name())
}

func (p *Policy) FilePath() string {
	return filepath.Join("apiproxy", "policies", p.FileName())
}

func (p *Policy) XML() ([]byte, error) {
	return utils.Struct2XMLDocText(p)
}
