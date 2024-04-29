# YAML to API Proxy
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

This command takes a YAML document and converts it into a ready-to-use Apigee API proxy bundle.

## Workflow

This command plays a crucial part in streamlining your Apigee development process.

1. **Design:** Craft your API proxy configuration using the more readable and manageable YAML format.
2. **Convert:** Feed your YAML document into the command to get a fully compliant API proxy bundle.
3. **Deploy:** Use the Apigee CLI to deploy the bundle

## Usage

The `yaml-to-apiproxy` command takes two parameters `-input` and `-output`

* `--input` is the YAML document that contains the API proxy definitions

* `--output` is either a bundle zip or a bundle directory to be created

* `--output` full path is created if it does not exist (like `mkdir -p`)

Bundle resources are read relative to the location of the `--input`

### Examples

* Creating a bundle zip
```shell
apigee-go-gen transform yaml-to-apiproxy \
  --input ./examples/yaml-first/petstore/apiproxy.yaml \
  --output ./out/apiproxies/petstore.zip 
```
* Creating a bundle directory
```shell
apigee-go-gen transform yaml-to-apiproxy \
  --input ./examples/yaml-first/petstore/apiproxy.yaml \
  --output ./out/apiproxies/petstore
```
