# Render Shared Flow
<!--
  Copyright 2024 Google LLC

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
-->

This command takes a YAML template representing shared flow and creates a traditional (XML-based) [Apigee shared flow](https://cloud.google.com/apigee/docs/api-platform/fundamentals/shared-flows) bundle.

Under the hood, this command combines the [render template](./render-template.md) and [transform yaml-to-sharedflow](../../transform/commands/yaml-to-apiproxy.md) commands into one

Using a template workflow offers several advantages over working directly with the traditional Apigee API proxy bundle:


## Usage

The `render sharedflow` command takes the following parameters:


```text
  -t, --template string          path to main template"
  -i, --include string           path to helper templates (globs allowed)
  -o, --output string            output directory or file
      --debug boolean            prints rendered template before transforming into shared flow"
  -d, --dry-run enum(xml|yaml)   prints rendered template after transforming into shared flow"
  -v, --validate boolean         check for unknown elements
      --set string               sets a key=value (bool,float,string), e.g. "use_ssl=true"
      --set-string string        sets key=value (string), e.g. "base_path=/v1/hello" 
      --values string            sets keys/values from YAML file, e.g. "./values.yaml"
      --set-file string          sets key=value where value is the content of a file, e.g. "my_data=./from/file.txt"
      --set-oas string           sets key=value where value is an OpenAPI spec, e.g. "my_spec=./petstore.yaml"
      --set-grpc string          sets key=value where value is a gRPC proto, e.g. "my_proto=./greeter.proto"
      --set-graphql string       sets key=value where value is a GraphQL schema, e.g. "my_schema=./resorts.graphql"
      --set-json string          sets key=value where value is JSON, e.g. 'servers=["server1","server2"]'
```