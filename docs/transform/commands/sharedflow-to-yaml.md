# Shared Flow to YAML
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

This command takes an Apigee shared flow bundle and converts it into an editable YAML document.

## Usage

The `sharedflow-to-yaml` command takes two parameters `-input` and `-output`.

* `--input` is either from a bundle zip file or an existing bundle directory.

* `--output` is the path for the YAML document to create

* `--output` full path is created if it does not exist (like `mkdir -p`)

Bundle resources are created in the same location as the `--output`

### Examples

* Reading bundle from a zip file
```shell
apigee-go-gen transform sharedflow-to-yaml \
  --input ./examples/sharedflows/owasp/owasp.zip \
  --output ./out/yaml-first/owasp/sharedflow.yaml
```
* Reading bundle from a directory
```shell
apigee-go-gen transform sharedflow-to-yaml \
  --input ./examples/sharedflows/owasp/ \
  --output ./out/yaml-first/owasp2/sharedflow.yaml
```


## YAML Document Structure

The YAML document created by `sharedflow-to-yaml` contains all the elements from the bundle in a single file.

The structure looks like this

```yaml
# From ./sharedflowbundle/flow-name.xml
SharedFlowBundle:
  .name: hello-world
  .revision: 1
  #...

# From ./sharedflowbundle/policies/*
Policies: 
  - AssignMessage: 
      .name: AM-SetTarget
      #...
  - RaiseFault:
      .name: RF-Set500
      #...

# From ./sharedflowbundle/sharedflows/*
SharedFlows: 
  - SharedFlow: 
      .name: proxy1
      #...

# From ./sharedflowbundle/resources/*/* 
Resources: 
  - Resource:
      Type: "properties"
      Path: "./path/to/resource.properties"
  - Resource: 
      Type: "jsc"
      Path: "./path/to/script.js"
```
