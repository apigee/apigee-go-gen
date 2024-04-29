# Render Commands
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


The `apigee-go-gen` tool includes the following set of template rendering commands to help you create Apigee API proxy bundles and shared flows.

* [render template](./commands/render-template.md) - Renders a [Go-style](https://pkg.go.dev/text/template) template
* [render apiproxy](./commands/render-apiproxy.md) - Combines [render template](./commands/render-template.md) and [yaml-to-apiproxy](../transform/commands/yaml-to-apiproxy.md) into one
* [render sharedflow](./commands/render-sharedflow.md) - Combines [render template](./commands/render-template.md) and [yaml-to-sharedflow](../transform/commands/yaml-to-sharedflow.md) into one



## Why use templates

Templates act like blueprints for creating Apigee API proxies, making it easy to generate them from popular formats like OpenAPI, GraphQL, and gRPC.

- [x] **Tailored Control**  
> Craft your API Proxies exactly how you need them, from security policies to dynamic behavior based on environment variables, and more!


- [x] **Unlock Your Specs**  
> Don't just describe your API – use the information in your OpenAPI specs (and others) to automatically build out parts of
your API proxy configuration.


## Template Language

The `render commands` are powered by the Go [text/template](https://pkg.go.dev/text/template) engine.

Some popular tools that also use this same engine are [Helm](https://helm.sh/), and [Hugo](https://gohugo.io/).

## Learn the Language

Here are some resources to get you started with the Go template language:

* Official [text/template](https://pkg.go.dev/text/template) package docs
* Hashicorp's [Go Template Tutorial](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax)


