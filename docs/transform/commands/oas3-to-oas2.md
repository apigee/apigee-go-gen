# OpenAPI 3 to OpenAPI 2
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

This command takes an OpenAPI 3 spec and converts into an OpenAPI 2 spec (also known as Swagger).

## Usage

The `oas3-to-oas2` command takes two parameters `-input` and `-output`

* `--input` is the OpenAPI 3 document to transform (either as JSON or YAML)

* `--output` is the OpenAPI 2 document to be created (either as JSON or YAML)

* `--output` full path is created if it does not exist (like `mkdir -p`)

* `--allow-cycles` external cyclic JSONRefs are replaced with empty placeholders `{}`

> You may omit the `--input` or `--output` flags to read or write from stdin or stdout

!!! Note
    Under the hood, this command uses the [kin-openapi](https://pkg.go.dev/github.com/getkin/kin-openapi) Go library to do the conversion


### Examples

* Reading and writing to files explicitly
```shell
apigee-go-gen transform oas3-to-oas2 \
  --input ./examples/specs/oas3/npr.yaml \
  --output ./out/specs/oas2/npr.yaml 
```

* Reading from stdin (from a file) and writing to stdout
```shell
apigee-go-gen transform oas3-to-oas2 < ./examples/specs/oas3/npr.yaml
```

* Reading from stdin (piped from another process) and writing to stdout
```shell
cat ./examples/specs/oas3/npr.yaml | apigee-go-gen transform oas3-to-oas2
```