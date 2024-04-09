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
		PrintVersion()
		return
	}

	if help {
		PrintUsage()
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

	proxyModel, err := v1.NewAPIProxyModel(input)
	if err != nil {
		utils.PrintErrorWithStackAndExit(err)
		return
	}

	if dryRun.IsXML() {
		xml, err := proxyModel.XML()
		if err != nil {
			utils.PrintErrorWithStackAndExit(err)
			return
		}
		fmt.Println(string(xml))
	} else if dryRun.IsYAML() {
		yaml, err := proxyModel.YAML()
		if err != nil {
			utils.PrintErrorWithStackAndExit(err)
			return
		}
		fmt.Println(string(yaml))
	}

	if validate {
		errs := v1.ValidateAPIProxyModel(proxyModel)
		for i := 0; i < len(errs) && i < 10; i++ {
			fmt.Printf("error: %s\n", errs[i].Error())
		}

		if len(errs) > 0 {
			os.Exit(1)
			return
		}
	}

	if dryRun != "" {
		return
	}

	err = v1.WriteBundleToDisk(proxyModel, output)
	if err != nil {
		utils.PrintErrorWithStackAndExit(err)
		return
	}

}
