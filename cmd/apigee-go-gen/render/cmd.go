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
	"github.com/micovery/apigee-yaml-toolkit/cmd/apigee-go-gen/render/bundle"
	sharedflow "github.com/micovery/apigee-yaml-toolkit/cmd/apigee-go-gen/render/shared-flow"
	"github.com/micovery/apigee-yaml-toolkit/cmd/apigee-go-gen/render/template"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "render",
	Short: "Render a bundle shared-flow from a template",
}

func init() {
	Cmd.AddCommand(bundle.Cmd)
	Cmd.AddCommand(sharedflow.Cmd)
	Cmd.AddCommand(template.Cmd)

}
