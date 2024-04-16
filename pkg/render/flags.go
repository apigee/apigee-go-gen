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

package render

import (
	"github.com/micovery/apigee-yaml-toolkit/pkg/flags"
	"github.com/micovery/apigee-yaml-toolkit/pkg/values"
)

func NewCommonFlags() *CommonFlags {
	return &CommonFlags{
		Values: &values.Map{},
	}
}

type CommonFlags struct {
	TemplateFile flags.String
	OutputFile   flags.String
	IncludeList  flags.IncludeList
	Values       *values.Map
}
