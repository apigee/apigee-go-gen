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

package yaml_to_bundle

import (
	"fmt"
	"github.com/go-errors/errors"
	v1 "github.com/micovery/apigee-yaml-toolkit/pkg/apigee/v1"
	"github.com/micovery/apigee-yaml-toolkit/pkg/flags"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var input flags.String
var output flags.String
var dryRun = flags.NewEnum([]string{"xml", "yaml"})
var validate = flags.NewBool(true)

var Cmd = &cobra.Command{
	Use:   "yaml-to-bundle",
	Short: "Transforms a YAML file into a bundle",
	RunE: func(cmd *cobra.Command, args []string) error {

		if strings.TrimSpace(string(output)) == "" && dryRun.IsUnset() {
			return errors.New("required flag(s) \"output\" not set")
		}

		err := v1.APIProxyModelYAML2Bundle(string(input), string(output), bool(validate), dryRun.Value)
		if errs, ok := err.(v1.ValidationErrors); ok {
			for i := 0; i < len(errs.Errors) && i < 10; i++ {
				_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", errs.Errors[i].Error())
			}
			os.Exit(1)
		}

		return err

	},
}

func init() {

	Cmd.Flags().SortFlags = false
	Cmd.Flags().VarP(&input, "input", "i", "path to API Proxy YAML file")
	Cmd.Flags().VarP(&output, "output", "o", "path to output zip file or dir")
	Cmd.Flags().VarP(&dryRun, "dry-run", "d", "print XML or YAML to stdout")
	Cmd.Flags().VarP(&validate, "validate", "v", "check for unknown elements")

	_ = Cmd.MarkFlagRequired("input")
}
