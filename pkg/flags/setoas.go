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
	"github.com/micovery/apigee-yaml-toolkit/pkg/parser"
	"github.com/micovery/apigee-yaml-toolkit/pkg/values"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type SetOAS struct {
	Data *values.Map
}

func NewSetOAS(data *values.Map) SetOAS {
	return SetOAS{Data: data}
}

func (v *SetOAS) Type() string {
	return "string"
}

func (v *SetOAS) String() string {
	return ""
}

func (v *SetOAS) Set(entry string) error {
	key, filePath, found := strings.Cut(entry, "=")
	if !found {
		return errors.Errorf("missing file path in set-oas for key=%s", key)
	}

	specFileMap := make(map[string]any)
	specFileText, err := os.ReadFile(filePath)
	if err != nil {
		return errors.New(err)
	}

	err = yaml.Unmarshal(specFileText, specFileMap)
	if err != nil {
		return errors.New(err)
	}

	oas, err := parser.ParseOAS(filePath)
	if err != nil {
		return errors.New(err)
	}
	specVersion := oas.GetSpecInfo().VersionNumeric

	if !(specVersion == 2.0 || specVersion >= 3.0) {
		return errors.Errorf(`OAS version "%v" is not supported`, specVersion)
	}

	v.Data.Set(key, specFileMap)
	v.Data.Set(fmt.Sprintf("%s_string", key), string(specFileText))

	return nil
}
