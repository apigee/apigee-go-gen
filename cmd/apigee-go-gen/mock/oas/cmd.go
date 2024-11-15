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

package oas

import (
	"fmt"
	"github.com/apigee/apigee-go-gen/pkg/common/resources"
	"github.com/apigee/apigee-go-gen/pkg/flags"
	"github.com/apigee/apigee-go-gen/pkg/mock"
	"github.com/spf13/cobra"
)

var input = flags.NewString("")
var output = flags.NewString("")
var debug = flags.NewBool(false)

var Cmd = &cobra.Command{
	Use:   "oas",
	Short: "Generate a mock API proxy from an OpenAPI 3.X Description",
	Long:  Usage(),
	RunE: func(cmd *cobra.Command, args []string) error {
		return mock.GenerateMockProxyBundle(string(input), string(output), bool(debug))
	},
}

func init() {
	Cmd.Flags().SortFlags = false
	Cmd.Flags().VarP(&input, "input", "i", `path to OpenAPI Description (e.g. "./path/to/openapi.yaml")`)
	Cmd.Flags().VarP(&output, "output", "o", `output directory or zip file (e.g. "./path/to/apiproxy.zip")`)
	Cmd.Flags().VarP(&debug, "debug", "", `prints rendered template before creating API proxy bundle"`)

	_ = Cmd.MarkFlagRequired("input")
	_ = Cmd.MarkFlagRequired("output")
	_ = Cmd.Flags().MarkHidden("debug")

}

func Usage() string {
	usageText := `
This command generates a mock API proxy bundle from an OpenAPI 3.X Description.

The mock API proxy includes the following features:

%[1]s

`

	mockFeatures, err := resources.FS.ReadFile("mock_features.txt")
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(usageText, mockFeatures)
}
