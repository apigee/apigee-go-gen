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

package render

import (
	"github.com/go-errors/errors"
	"github.com/micovery/apigee-yaml-toolkit/pkg/parser"
	"gopkg.in/yaml.v3"
	"os"
)

func RenderOAS(specFile string, flags *Flags) error {

	type RenderContext struct {
		OAS2    map[string]any
		OAS2Str string

		OAS3    map[string]any
		OAS3Str string

		Values map[string]any
	}

	OAS := make(map[string]any)
	specFileText, err := os.ReadFile(specFile)
	if err != nil {
		return errors.New(err)
	}

	err = yaml.Unmarshal(specFileText, OAS)
	if err != nil {
		return errors.New(err)
	}

	oas, err := parser.ParseOAS(specFile)
	if err != nil {
		return errors.New(err)
	}

	context := &RenderContext{
		Values: *flags.Values,
	}

	specVersion := oas.GetSpecInfo().VersionNumeric

	if specVersion == 2.0 {
		context.OAS2 = OAS
		context.OAS2Str = string(specFileText)

	} else if specVersion >= 3.0 {
		context.OAS3 = OAS
		context.OAS3Str = string(specFileText)
	} else {
		return errors.Errorf(`OAS version "%v" is not supported`, specVersion)
	}

	return RenderGeneric(context, flags)
}
