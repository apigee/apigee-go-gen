# Using GraphQL Schema
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

You can use the [render apiproxy](./commands/render-apiproxy.md) command to create an Apigee API proxy bundle using a template and a GraphQL schema as input.

While GraphQL schemas might not contain all the necessary details for a complete API proxy bundle, this command offers flexibility through the `--set` and `--set-string` parameters.
This works similar to how values are set in Helm charts.

## How it works


- [x] **Start with Your Template** 
> The template, your inputs, and the schema guide the generation of an intermediate YAML configuration.
- [x] **Access the Schema**  
> Use `--set-graphql` to access a GraphQL schema text and [AST](https://pkg.go.dev/github.com/vektah/gqlparser/v2/ast#Schema) during template rendering. 
- [x] **Inject Your Values** 
> Use `--set` and `--set-string` to provide missing values (like target URLs) for your template.

## Examples

Check out the [examples/templates/graphql](https://github.com/micovery/apigee-go-gen/blob/main/examples/templates/graphql/apiproxy.yaml) directory for an example of building the intermediate YAML for a GraphQL API proxy.

Here is how you would use the [render apiproxy](./commands/render-apiproxy.md) command with this example:

#### Create bundle zip
```shell
apigee-go-gen render apiproxy \
     --template ./examples/templates/graphql/apiproxy.yaml \
     --set-graphql schema=./examples/graphql/resorts.graphql \
     --set-string "api_name=resorts-api" \
     --set-string "base_path=/graphql" \
     --set-string "target_url=https://example.com/graphql" \
     --include ./examples/templates/graphql/*.tmpl \
     --output ./out/apiproxies/resorts.zip
``` 

#### Create bundle dir
```shell
apigee-go-gen render apiproxy \
     --template ./examples/templates/graphql/apiproxy.yaml \
     --set-graphql schema=./examples/graphql/resorts.graphql \
     --set-string "api_name=resorts-api" \
     --set-string "base_path=/graphql" \
     --set-string "target_url=https://example.com/graphql" \
     --include ./examples/templates/graphql/*.tmpl \
     --output ./out/apiproxies/resorts.zip
``` 
