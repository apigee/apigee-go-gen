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
	"fmt"
	"github.com/micovery/apigee-yaml-toolkit/cmd/yaml2xml/resources"
	v1 "github.com/micovery/apigee-yaml-toolkit/pkg/apigee/v1"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"os"
)

func main() {

	var help bool
	var version bool
	var input string
	var output string
	var validate utils.FlagBool = true
	var dryRun utils.DryRunFlag = ""

	var err error

	flag.BoolVar(&version, "version", false, "(optional) prints version text")
	flag.BoolVar(&help, "help", false, `(optional) prints additional help text`)
	flag.StringVar(&input, "input", "", "(required) path to API Proxy YAML. e.g. \"./hello-world-v1.yaml\"")
	flag.StringVar(&output, "output", "", "(required) output zip file or directory. e.g \"./hello-world-v1\" or .\"./hello-world-v1.zip\"")
	flag.Var(&dryRun, "dry-run", "(optional) prints the full bundle text and exits, valid values are \"xml\" or \"yaml\"")
	flag.Var(&validate, "validate", "(optional) exits early when it finds unknown elements")

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

	if output == "" && dryRun.IsUnset() {
		utils.RequireParamAndExit("output")
		return
	}

	err, validationErrs := v1.APIProxyModelYAML2Bundle(input, output, bool(validate), dryRun)
	if err != nil {
		utils.PrintErrorWithStackAndExit(err)
		return
	}

	if len(validationErrs) > 0 {
		for i := 0; i < len(validationErrs) && i < 10; i++ {
			fmt.Printf("error: %s\n", validationErrs[i].Error())
		}
		os.Exit(1)
	}

}
