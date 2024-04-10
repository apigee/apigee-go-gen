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
	"github.com/micovery/apigee-yaml-toolkit/cmd/render-oas/resources"
	"github.com/micovery/apigee-yaml-toolkit/pkg/render"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"strings"
)

func main() {

	var specFile string

	flag.StringVar(&specFile, "spec", "", `(required) path to OAS file e.g "./oas.yaml"`)

	flags, err := render.GetRenderFlags(utils.PrintVersion, resources.PrintUsage)
	if err != nil {
		utils.PrintErrorWithStackAndExit(err)
		return
	}

	if strings.TrimSpace(specFile) == "" {
		utils.RequireParamAndExit("spec")
		return
	}

	err = render.RenderOAS(specFile, flags)
	if err != nil {
		utils.PrintErrorWithStackAndExit(err)
		return
	}
}
