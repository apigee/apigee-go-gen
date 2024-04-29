# Using Custom Helpers
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

Sometimes it's useful to create your own custom helper functions that you can use during the
template rendering process.

Use the `--include` flag to specify one or more files containing helper functions.

!!!Note
    You can pass the `--include` flag multiple times, or even use [glob patterns](https://pkg.go.dev/github.com/bmatcuk/doublestar#readme-patterns) to include multiple files.


## Example
Below is a sample helper function. Let's see how to use it.

```gotemplate
{{- define "say_hello" -}}
  Hello {{ $. }} !
{{- end -}}
```

1. First, place this block inside a helper file (e.g. `helper.tmpl`)

2. Then, pass  `--include ./helper.tmpl` to the `render`) command.

Finally, in order to invoke the custom helper function, use the  `{{ include ... }}` built-in function from your main template.

e.g.

```yaml
Message: {{ include "say_hello" "World" }}
```

The engine will render the `say_hello` block separately, and include the output text wherever you put `{{ include ... }}` function.

For this example, the final rendered text will look like this
```yaml
Message: Hello World !
```
