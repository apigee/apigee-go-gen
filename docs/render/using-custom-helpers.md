# Using Custom Helpers
<!--
  Copyright 2025 Google LLC

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

You can define helper functions within the main template file itself, or in separate helper files.

By default, the tool will look for a file named `_helpers.tpl` in the same directory as the template, and include it.

Alternatively, you can use the `--include` flag to specify one or more files containing helper functions.


## Example
Below is a sample helper function. Let's see how to use it.

```gotemplate
{{- define "say_hello" -}}
  Hello {{ $. }} !
{{- end -}}
```

1. First, place this block inside a helper file (e.g. `my_helpers.tpl`)

2. Then, pass  `--include ./my_helpers.tpl` to the `render`) command.

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
