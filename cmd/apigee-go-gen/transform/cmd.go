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

//goland:noinspection GoSnakeCaseUsage
import (
	apiproxy_to_yaml "github.com/micovery/apigee-go-gen/cmd/apigee-go-gen/transform/apiproxy-to-yaml"
	json_to_yaml "github.com/micovery/apigee-go-gen/cmd/apigee-go-gen/transform/json-to-yaml"
	oas2_to_oas3 "github.com/micovery/apigee-go-gen/cmd/apigee-go-gen/transform/oas2-to-oas3"
	oas3_to_oas2 "github.com/micovery/apigee-go-gen/cmd/apigee-go-gen/transform/oas3-to-oas2"
	resolve_refs "github.com/micovery/apigee-go-gen/cmd/apigee-go-gen/transform/resolve-refs"
	sharedflow_to_yaml "github.com/micovery/apigee-go-gen/cmd/apigee-go-gen/transform/sharedflow-to-yaml"
	xml_to_yaml "github.com/micovery/apigee-go-gen/cmd/apigee-go-gen/transform/xml-to-yaml"
	yaml_to_apiproxy "github.com/micovery/apigee-go-gen/cmd/apigee-go-gen/transform/yaml-to-apiproxy"
	yaml_to_json "github.com/micovery/apigee-go-gen/cmd/apigee-go-gen/transform/yaml-to-json"
	yaml_to_sharedflow "github.com/micovery/apigee-go-gen/cmd/apigee-go-gen/transform/yaml-to-sharedflow"
	yaml_to_xml "github.com/micovery/apigee-go-gen/cmd/apigee-go-gen/transform/yaml-to-xml"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "transform",
	Short: "Transform between shared flow, API proxy, xml, yaml, and others",
}

func init() {
	Cmd.AddCommand(xml_to_yaml.Cmd)
	Cmd.AddCommand(yaml_to_xml.Cmd)
	Cmd.AddCommand(apiproxy_to_yaml.Cmd)
	Cmd.AddCommand(yaml_to_apiproxy.Cmd)
	Cmd.AddCommand(sharedflow_to_yaml.Cmd)
	Cmd.AddCommand(yaml_to_sharedflow.Cmd)
	Cmd.AddCommand(oas2_to_oas3.Cmd)
	Cmd.AddCommand(oas3_to_oas2.Cmd)
	Cmd.AddCommand(resolve_refs.Cmd)
	Cmd.AddCommand(json_to_yaml.Cmd)
	Cmd.AddCommand(yaml_to_json.Cmd)
}
