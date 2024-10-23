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

package oas_overlay

import (
	"github.com/apigee/apigee-go-gen/pkg/flags"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/spf13/cobra"
)

var spec flags.String
var overlay flags.String
var output flags.String

var Cmd = &cobra.Command{
	Use:   "oas-overlay",
	Short: "Transforms Open API spec by applying overlay file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return utils.OASOverlay(string(overlay), string(spec), string(output))
	},
}

func init() {

	Cmd.Flags().SortFlags = false
	Cmd.Flags().VarP(&spec, "spec", "s", "path to OpenAPI spec file (optional)")
	Cmd.Flags().VarP(&overlay, "overlay", "", "path to overlay file")
	Cmd.Flags().VarP(&output, "output", "o", "path to output file, or omit to use stdout")

	_ = Cmd.MarkFlagRequired("overlay")

}
