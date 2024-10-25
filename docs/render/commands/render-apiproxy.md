# Render API Proxy
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

This command takes a YAML template representing an API proxy and creates a traditional (XML-based) [Apigee API proxy](https://cloud.google.com/apigee/docs/api-platform/fundamentals/understanding-apis-and-api-proxies#whatisanapiproxy) bundle.

Under the hood, this command combines the [render template](./render-template.md) and [transform yaml-to-apiproxy](../../transform/commands/yaml-to-apiproxy.md) commands into one

Using a template based workflow offers several advantages over working directly with the traditional Apigee API proxy bundle. e.g.

- [x] **Enhanced Customization**  
> Tweak your API proxy configurations with the readability of YAML.

- [x] **Seamless Spec Synchronization**  
> Template-generated API proxy bundles can be easily synced with your specs by re-generating them when changes occur.

- [x] **Streamline Your Development**  
> YAML's versatility allows for easy version control, automation, and integration into CI/CD pipelines.

## Usage

The `render apiproxy` command takes the following parameters:


```text
  -t, --template string          path to main template"
  -i, --include string           path to helper templates (globs allowed)
  -o, --output string            output directory or file
      --debug boolean            prints rendered template before transforming into API proxy"
  -d, --dry-run enum(xml|yaml)   prints rendered template after transforming into API Proxy"
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