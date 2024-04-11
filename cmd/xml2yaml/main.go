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
	"bufio"
	"flag"
	"fmt"
	"github.com/micovery/apigee-yaml-toolkit/cmd/xml2yaml/resources"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"os"
)

func main() {

	var help bool
	var version bool

	flag.BoolVar(&version, "version", false, "(optional) prints version text")
	flag.BoolVar(&help, "help", false, `(optional) prints additional help text`)

	flag.Parse()

	if version {
		utils.PrintVersion()
		return
	}

	if help {
		resources.PrintUsage()
		return
	}

	var err error
	yamlText, err := utils.XMLText2YAMLText(bufio.NewReader(os.Stdin))
	if err != nil {
		utils.PrintErrorWithStackAndExit(err)
		return
	}

	fmt.Printf("%s\n", string(yamlText))
}
