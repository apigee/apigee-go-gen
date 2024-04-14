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

package main

import (
	"flag"
	"github.com/micovery/apigee-yaml-toolkit/cmd/bundle2yaml/resources"
	"github.com/micovery/apigee-yaml-toolkit/pkg/bundle"
	"github.com/micovery/apigee-yaml-toolkit/pkg/flags"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
)

func main() {
	var help bool
	var version bool
	var input string
	var outputFile string
	dryRun := flags.Bool(false)

	var err error

	flag.BoolVar(&version, "version", false, "(optional) prints version text")
	flag.BoolVar(&help, "help", false, `(optional) prints additional help text`)
	flag.StringVar(&input, "input", "", "(required) path to bundle zip or dir. e.g. \"./hello-world-rev1.zip\" or \"./hello-world-rev1\" ")
	flag.StringVar(&outputFile, "output", "", "(required) output YAML file. e.g \"./hello-world/apiproxy.yaml\"")
	flag.Var(&dryRun, "dry-run", "(optional) prints YAML document to stdout")

	flag.Parse()

	if version {
		utils.PrintVersion()
		return
	}

	if help {
		resources.PrintUsage()
		return
	}

	if input == "" {
		utils.RequireParamAndExit("input")
		return
	}

	if outputFile == "" {
		utils.RequireParamAndExit("output")
		return
	}

	if err = bundle.ProxyBundle2YAMLFile(input, outputFile, bool(dryRun)); err != nil {
		utils.PrintErrorWithStackAndExit(err)
		return
	}
}
