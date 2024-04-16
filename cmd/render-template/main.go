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
	"github.com/micovery/apigee-yaml-toolkit/cmd/render-template/resources"
	"github.com/micovery/apigee-yaml-toolkit/pkg/flags"
	"github.com/micovery/apigee-yaml-toolkit/pkg/render"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"strings"
)

func main() {
	var err error

	var dryRun = flags.NewBool(false)
	cFlags := render.NewCommonFlags()

	flag.Var(&dryRun, "dry-run", "(optional) prints the bundle contents and exits, valid values are \"true\" or \"false\"")
	render.SetupCommonFlags(cFlags)
	flag.Parse()

	if cFlags.Version {
		utils.PrintVersion()
		return
	}

	if cFlags.Help {
		resources.PrintUsage()
		return
	}

	if strings.TrimSpace(cFlags.OutputFile) == "" && dryRun == false {
		utils.RequireParamAndExit("output")
		return
	}

	err = render.RenderGenericTemplate(cFlags, bool(dryRun))
	if err != nil {
		utils.PrintErrorWithStackAndExit(err)
		return
	}
}
