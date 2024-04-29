# YAML to XML
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

This command takes a YAML snippet and converts it into XML.

This is useful for validating and troubleshooting your YAML code.

**Usage**

The `yaml-to-xml` command takes two parameters `-input` and `-output`

* `--input` is the YAML document to transform

* `--output` is the XML document to be created

* `--output` full path is created if it does not exist (like `mkdir -p`)

> You may omit the `--input` or `--output` flags to read or write from stdin or stdout

### Examples
Below are a few examples for using the `yaml-to-xml` command.

#### From a file
Reading input redirected from a file
```shell
apigee-go-gen transform yaml-to-xml < ./examples/snippets/ducks.yaml
```

#### From stdin
Reading input from `stdin` directly
```shell
apigee-go-gen transform yaml-to-xml << EOF
Parent:
  - Child: Fizz
  - Child: Buzz
EOF
```

#### From a process
Reading input piped from another process
```shell
echo '
Parent:
  - Child: Fizz
  - Child: Buzz' | apigee-go-gen transform yaml-to-xml
```
