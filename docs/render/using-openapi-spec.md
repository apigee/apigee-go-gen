# Using OpenAPI Spec
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

You can use the [render apiproxy](./commands/render-apiproxy.md) command to create an Apigee API proxy bundle using a template and an [OpenAPI spec](https://www.openapis.org/) as input.

## How it works

- [x] **Start with Your Template**
> This is your baseline. Include any standard policies or settings you want in your final proxy.
- [x] **Customize the Output** 
> Your template uses special placeholders that are replaced with details from your OpenAPI spec.
- [x] **Control the Output** 
> Use control logic in your template to adjust your proxy configuration based on your OpenAPI spec.
- [x] **Access the Spec** 
> Use `--set-oas` to access the OpenAPI as a [map](https://go.dev/blog/maps) (and as text) during template rendering.

!!! Note
    Both OAS2 and OAS3 are supported using the `--set-oas` flag

## Examples

Check out the included OAS3 template at [examples/templates/oas3](https://github.com/apigee/apigee-go-gen/blob/main/examples/templates/oas3/apiproxy.yaml).

Here is how you would use the [render apiproxy](./commands/render-apiproxy.md) command with this example:

#### Create bundle zip

```shell
apigee-go-gen render apiproxy \
    --template ./examples/templates/oas3/apiproxy.yaml \
    --set-oas spec=./examples/specs/oas3/petstore.yaml \
    --include ./examples/templates/oas3/*.tmpl \
    --output ./out/apiproxies/petstore.zip
```

#### Create bundle dir
```shell
apigee-go-gen render apiproxy \
    --template ./examples/templates/oas3/apiproxy.yaml \
    --set-oas spec=./examples/specs/oas3/petstore.yaml \
    --include ./examples/templates/oas3/*.tmpl \
    --output ./out/apiproxies/petstore
```

## Dry run

For rapid development, you can print the rendered template directly to stdout in your terminal. 

Add the `--dry-run xml` or `--dry-run yaml` flag. e.g.

=== "XML output"
```shell
apigee-go-gen render apiproxy \
    --template ./examples/templates/oas3/apiproxy.yaml \
    --set-oas spec=./examples/specs/oas3/petstore.yaml \
    --include ./examples/templates/oas3/*.tmpl \
    --dry-run xml
```