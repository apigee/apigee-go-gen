# JSON to YAML
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

This command takes a JSON document and converts it into YAML.

## Usage

The `json-to-yaml` command takes two parameters `-input` and `-output`

* `--input` is the JSON document to transform

* `--output` is the YAML document to be created

* `--output` full path is created if it does not exist (like `mkdir -p`)

> You may omit the `--input` or `--output` flags to read or write from stdin or stdout

### Examples

Below are a few examples for using the `json-to-yaml` command.

#### From files
Reading and writing to files explicitly
```shell
apigee-go-gen transform json-to-yaml \
  --input ./examples/snippets/ducks.json \
  --output ./out/snippets/ducks.yaml 
```

#### From stdin / stdout
Reading from stdin (from a file) and writing to stdout
```shell
apigee-go-gen transform json-to-yaml < ./examples/snippets/ducks.json
```

#### From a process
Reading from stdin (piped from another process) and writing to stdout
```shell
cat ./examples/snippets/ducks.json | apigee-go-gen transform json-to-yaml
```

