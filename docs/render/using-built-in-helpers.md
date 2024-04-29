# Using Built-in Helpers
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

The template rendering commands include a set of built-in helper functions to assist you with the rendering process.

## Functions

### **include**
```go
 func include(template string, data any) string
```

This function allows you to invoke your own [custom helper functions](./using-custom-helpers.md)

e.g.

```gotemplate
{{ include "sayHello" $data }}
```


### **os_writefile**
```go
func os_writefile(dest string, content string) string
```

Write a file to the output directory

The destination path is relative to the output directory.
( ".." or absolute paths are not allowed)

This function outputs the destination file path.

e.g.
```gotemplate
{{ os_writefile "./dst/filename.txt" "contents" }}
```

### **os_copyfile**
```go
func os_copyfile(dest string, src string) string
```

Copies files to the output directory.
This function outputs the destination file path as a string.

The destination path is relative to the output directory
( ".." or absolute paths are not allowed)

The source path is relative to the main template file directory
( ".." or absolute paths are not allowed)

e.g.
```gotemplate
{{ os_copyfile "./dest/lib.jar" "./src/lib.jar" }}
```

### **os_getenvs**
```go 
func os_getenvs() map[string]string
```

Gets all environment variables as a dictionary

e.g.
```gotemplate
{{ $envs := os_getenvs }}
```

### **os_getenv**
```go
func os_getenv**(env string) string
```

Gets the value of the specified env var

e.g.
```gotemplate
{{ os_getenv "USER" }}
```

### **slug_make**
```go
func slug_make(in string) string
```

Converts string to a slug

e.g.
```gotemplate
{{ slug_maek "My API proxy" }}
```
The example above outputs "my-api-proxy"

### **url_parse**
```go
func url_parse(url string) net.URL
```

Parse a URL into its parts.

This function outputs a [net.URL](https://pkg.go.dev/net/url#URL) struct.

e.g.
```gotemplate
{{ $url := url_parse "https://example.com/foo/bar" }}
```

### **blank**
```go
func blank() string
```

Outputs empty string.
This is useful to consume the output of another function.

e.g.
```gotemplate
{{ os_writefile "./dest/file" "foo" | blank }}
```

### **deref**
```go
func deref(*any) any
```

Dereferences the input pointer.

### **fmt_printf**
```go
func fmt_printf**(pattern string, args ... string)
```

Write to stdout during the rendering process.
This function is useful for so called "printf" debugging.

For example, you can use it to trace the template rendering as it runs.
Or, can also use it to dump values to stdout in order to see the contents.

e.g.
```gotemplate
{{ fmt_printf "Hello World\n" }}
```

```gotemplate
{{ fmt_printf "url: %%v\n" $url }}
```

## Libraries
### **Sprig**

This library contains a lot of useful functions for string manipulation, accessing maps, lists, encoding, and more.

Functions from [Sprig](https://masterminds.github.io/sprig/) library are available during rendering.

e.g.
 ```gotemplate
 {{ "Hello World" | upper }}
 ``` 

```gotemplate
{{ list "hello" "world" | join "_" }}
```

      