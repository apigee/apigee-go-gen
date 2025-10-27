//  Copyright 2025 Google LLC
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

package json_to_tf

import (
	"github.com/apigee/apigee-go-gen/pkg/flags"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/spf13/cobra"
)

var input flags.String
var output flags.String

var Cmd = &cobra.Command{
	Use:   "json-to-tf",
	Short: "Transforms the input JSON into Terraform HCL",
	Long: `This command reads JSON (form stdin) and outputs Terraform HCL (to stdout).
The input file must use the JSON Configuration Syntax.
See: https://developer.hashicorp.com/terraform/language/syntax/json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return utils.JSONFileToTFFile(string(input), string(output))
	},
}

func init() {
	Cmd.Flags().SortFlags = false
	Cmd.Flags().VarP(&input, "input", "i", "path to input file, or omit it to use stdin")
	Cmd.Flags().VarP(&output, "output", "o", "path to output file, or omit it to use stdout")

}
