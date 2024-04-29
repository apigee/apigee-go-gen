# YAML to JSON
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

This command takes a YAML document and converts it into JSON.

## Usage

The `yaml-to-json` command takes two parameters `-input` and `-output`

* `--input` is the YAML document to transform

* `--output` is the JSON document to be created

* `--output` full path is created if it does not exist (like `mkdir -p`)

> You may omit the `--input` or `--output` flags to read or write from stdin or stdout

### Examples

* Reading and writing to files explicitly
```shell
apigee-go-gen transform yaml-to-json \
  --input ./examples/snippets/ducks.yaml \
  --output ./out/snippets/ducks.json 
```

* Reading from stdin (from a file) and writing to stdout
```shell
apigee-go-gen transform yaml-to-json < ./examples/snippets/ducks.yaml
```

* Reading from stdin (piped from another process) and writing to stdout
```shell
cat ./examples/snippets/ducks.yaml | apigee-go-gen transform yaml-to-json
```