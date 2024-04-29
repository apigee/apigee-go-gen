# Resolve Refs
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

This command takes a JSON or YAML document and resolves and replaces all external [JSONRefs](http://jsonref.org/).

## Usage

The `resolve-refs` command takes two parameters `-input` and `-output`

* `--input` is the document to transform (either as JSON or YAML)

* `--output` is the document to be created (either as JSON or YAML)

* `--output` full path is created if it does not exist (like `mkdir -p`)

* `--allow-cycles` external cyclic JSONRefs are replaced with empty placeholders `{}`

> You may omit the `--input` or `--output` flags to read or write from stdin or stdout

!!! Note
    Only JSONRefs pointing to external documents are replaced. If a JSONRef points back within the same document, it is left unchanged.

