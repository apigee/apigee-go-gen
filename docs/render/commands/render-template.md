# Render Template
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

This command takes input template, and renders it using context data provided through the `--set` flags.

You can use this command to render any template (input template can be any text file)

Under the hood, this command uses the Go [text/template](https://pkg.go.dev/text/template) engine to render the input template.


## Usage

The `render template` command takes the following  parameters:

```text
-t, --template string      path to main template
-i, --include string       path to helper templates (globs allowed)
-o, --output string        output directory or file
-d, --dry-run boolean      prints rendered template to stdout"
    --set string           sets a key=value (bool,float,string), e.g. "use_ssl=true"
    --set-string string    sets key=value (string), e.g. "base_path=/v1/hello"
    --values string        sets keys/values from YAML file, e.g. "./values.yaml"
    --set-file string      sets key=value where value is the content of a file, e.g. "my_data=./from/file.txt"
    --set-oas string       sets key=value where value is an OpenAPI Description, e.g. "my_spec=./petstore.yaml"
    --set-grpc string      sets key=value where value is a gRPC proto, e.g. "my_proto=./greeter.proto"
    --set-graphql string   sets key=value where value is a GraphQL schema, e.g. "my_schema=./resorts.graphql"
    --set-json string      sets key=value where value is JSON, e.g. 'servers=["server1","server2"]'
```


## Rendering Context

During the rendering process, the there is a global context variable called `$.Values`. 
This variable contains all the information passed using the `--set` flags.

You can use the `--set` flags to inject values into the rendering process, or even control
the flow of the rendering process dynamically.

