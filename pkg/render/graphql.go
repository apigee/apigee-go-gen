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
	parser2 "github.com/micovery/apigee-yaml-toolkit/pkg/parser"
	"github.com/vektah/gqlparser/v2/ast"
)

func RenderGraphQL(specFile string, flags *Flags) error {

	parserResult, schemaStr, err := parser2.ParseGraphQLSchema(specFile)
	if err != nil {
		return err
	}

	type RenderContext struct {
		Schema    ast.Schema
		SchemaStr string

		Values map[string]any
	}

	context := &RenderContext{
		Schema:    *parserResult,
		SchemaStr: string(schemaStr),
		Values:    *flags.Values,
	}

	return RenderGeneric(context, flags)
}
