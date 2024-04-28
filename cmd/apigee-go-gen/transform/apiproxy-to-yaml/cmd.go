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

package apiproxy_to_yaml

import (
	"github.com/go-errors/errors"
	"github.com/micovery/apigee-go-gen/pkg/apiproxy"
	"github.com/micovery/apigee-go-gen/pkg/flags"
	"github.com/spf13/cobra"
	"strings"
)

var input flags.String
var output flags.String
var dryRun = flags.NewBool(false)

var Cmd = &cobra.Command{
	Use:   "apiproxy-to-yaml",
	Short: "Transforms an apiproxy into a YAML file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(string(output)) == "" && dryRun == false {
			return errors.New("required flag(s) \"output\" not set")
		}

		return apiproxy.Bundle2YAMLFile(string(input), string(output), bool(dryRun))
	},
}

func init() {

	Cmd.Flags().SortFlags = false
	Cmd.Flags().VarP(&input, "input", "i", "path to bundle zip or dir")
	Cmd.Flags().VarP(&output, "output", "o", "path to output YAML file or dir")
	Cmd.Flags().VarP(&dryRun, "dry-run", "d", "prints YAML document to stdout")

	_ = Cmd.MarkFlagRequired("input")

}
