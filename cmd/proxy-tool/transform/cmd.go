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

package transform

import (
	bundle_to_yaml "github.com/micovery/apigee-yaml-toolkit/cmd/proxy-tool/transform/bundle-to-yaml"
	sharedflow_to_yaml "github.com/micovery/apigee-yaml-toolkit/cmd/proxy-tool/transform/sharedflow-to-yaml"
	xml_to_yaml "github.com/micovery/apigee-yaml-toolkit/cmd/proxy-tool/transform/xml-to-yaml"
	yaml_to_bundle "github.com/micovery/apigee-yaml-toolkit/cmd/proxy-tool/transform/yaml-to-bundle"
	yaml_to_sharedflow "github.com/micovery/apigee-yaml-toolkit/cmd/proxy-tool/transform/yaml-to-sharedflow"
	yaml_to_xml "github.com/micovery/apigee-yaml-toolkit/cmd/proxy-tool/transform/yaml-to-xml"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "transform",
	Short: "Transform between shared-flow, bundle, xml, and yaml",
}

func init() {
	Cmd.AddCommand(xml_to_yaml.Cmd)
	Cmd.AddCommand(yaml_to_xml.Cmd)
	Cmd.AddCommand(bundle_to_yaml.Cmd)
	Cmd.AddCommand(yaml_to_bundle.Cmd)
	Cmd.AddCommand(sharedflow_to_yaml.Cmd)
	Cmd.AddCommand(yaml_to_sharedflow.Cmd)
}
