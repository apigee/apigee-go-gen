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

package xml_to_yaml

import (
	"bufio"
	"fmt"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"github.com/spf13/cobra"
	"os"
)

var Cmd = &cobra.Command{
	Use:   "xml-to-yaml",
	Short: "Transform XML snippet to YAML",
	Long:  `This command reads XML (form stdin) and outputs YAML (to stdout)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		yamlText, err := utils.XMLText2YAMLText(bufio.NewReader(os.Stdin))
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", string(yamlText))
		return nil
	},
}
