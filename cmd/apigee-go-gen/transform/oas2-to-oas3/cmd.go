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

package oas2_to_oas3

import (
	"github.com/apigee/apigee-go-gen/pkg/flags"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/spf13/cobra"
)

var input flags.String
var output flags.String
var allowCycles = flags.NewBool(false)

var Cmd = &cobra.Command{
	Use:   "oas2-to-oas3",
	Short: "Transforms the input OpenAPI 2 Description into an OpenAPI 3 Description",
	RunE: func(cmd *cobra.Command, args []string) error {
		return utils.OAS2FileToOAS3File(string(input), string(output), bool(allowCycles))
	},
}

func init() {

	Cmd.Flags().SortFlags = false
	Cmd.Flags().VarP(&input, "input", "i", "path to input file, or omit to use stdin")
	Cmd.Flags().VarP(&output, "output", "o", "path to output file, or omit to use stdout")
	Cmd.Flags().VarP(&allowCycles, "allow-cycles", "c", "allow cyclic JSONRefs")

}
