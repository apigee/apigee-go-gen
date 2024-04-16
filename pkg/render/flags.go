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
	"flag"
	"github.com/micovery/apigee-yaml-toolkit/pkg/flags"
	"github.com/micovery/apigee-yaml-toolkit/pkg/values"
)

func NewCommonFlags() *CommonFlags {
	return &CommonFlags{
		Values: &values.Map{},
	}
}

type CommonFlags struct {
	Version      bool
	TemplateFile string
	OutputFile   string
	Help         bool
	IncludeList  flags.IncludeList
	Values       *values.Map
}

func SetupCommonFlags(cFlags *CommonFlags) {
	setValue := flags.NewSetAny(cFlags.Values)
	setValueStr := flags.NewSetString(cFlags.Values)
	setValueFile := flags.NewValues(cFlags.Values)
	setFile := flags.NewSetFile(cFlags.Values)
	setOAS := flags.NewSetOAS(cFlags.Values)
	setGraphQL := flags.NewSetGraphQL(cFlags.Values)
	setGRPC := flags.NewSetGRPC(cFlags.Values)
	setJSON := flags.NewSetJSON(cFlags.Values)

	flag.BoolVar(&cFlags.Version, "version", false, "(optional) prints version text")
	flag.BoolVar(&cFlags.Help, "help", false, `(optional) prints additional help text`)
	flag.StringVar(&cFlags.TemplateFile, "template", "", `(required) path to main template e.g. "./input.tpl"`)
	flag.Var(&cFlags.IncludeList, "include", `(optional,0..*) path/glob for templates to include e.g. "./helpers/*.tpl"`)
	flag.StringVar(&cFlags.OutputFile, "output", "", "(required) output directory or file")
	flag.Var(&setValue, "set", `(optional,0..*) sets a context key/value  .e.g "boolVal=true" or "floatVal=1.37"`)
	flag.Var(&setValueStr, "set-string", `(optional,0..*) sets a context key/value .e.g key1="value1"`)
	flag.Var(&setValueFile, "values", "(optional,0..1) sets context keys/values from YAML file")
	flag.Var(&setFile, "set-file", "(optional,0..*) sets context key/value where value is the content of a file e.g. \"my_script=foo.sh\"")
	flag.Var(&setOAS, "set-oas", "(optional,0..*) sets context key/value where value is an OpenAPI spec of a file e.g. \"my_oas=petstore.yaml\"")
	flag.Var(&setGRPC, "set-grpc", "(optional,0..*) sets context key/value where value is a gRPC proto of a file e.g. \"my_proto=greeter.proto\"")
	flag.Var(&setGraphQL, "set-graphql", "(optional,0..*) sets context key/value where value is a GraphQL schema of a file e.g. \"my_schema=resorts.graphql\"")
	flag.Var(&setJSON, "set-json", "(optional,0..*) sets context key/value where value is JSON  e.g. 'servers=[\"server1\",\"server2\"]'")
}
