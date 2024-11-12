# OAS Overlay
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

This command applies an [OpenAPI Overlay 1.0](https://github.com/OAI/Overlay-Specification/blob/main/versions/1.0.0.md) to an OpenAPI Description.

## Usage

The `oas-overlay` command takes the following parameters:

* `--overlay` is the path to the OpenAPI Overlay (either as JSON or YAML)

* `--spec` (*optional*)  is the path to the OpenAPI Description to transform (either as JSON or YAML)
  
* `--output` is the document to be created (either as JSON or YAML)


> The `--spec` parameter is optional. If omitted, the OAS path is read from the `extends` property of the Overlay.
> In this case, the path is relative to the location of the Overlay file itself.



> You may omit the `--output` flags to write to stdout  
>  Full path is created if the `--output` does not exist (like `mkdir -p`)



### Examples

Below are a few examples for using the `oas-overlay` command.

#### Write to file
Writing the output to a new file
```shell
apigee-go-gen transform oas-overlay \
  --spec ./examples/specs/oas3/petstore.yaml \
  --overlay ./examples/overlays/petstore.yaml \
  --output ./out/specs/oas3/petstore-overlaid.yaml 
```

#### Write To stdout
Writing the output to stdout
```shell
apigee-go-gen transform oas-overlay \
  --spec ./examples/specs/oas3/petstore.yaml \
  --overlay ./examples/overlays/petstore.yaml
```

#### Using the `extends` property 
Instead of passing `--spec`, use the value of the `extends` property in the overlay
```shell
apigee-go-gen transform oas-overlay \
  --overlay ./examples/overlays/petstore.yaml
```
