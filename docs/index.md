# Apigee Go Gen
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

[![Go Report Card](https://goreportcard.com/badge/github.com/micovery/apigee-go-gen)](https://goreportcard.com/report/github.com/micovery/apigee-go-gen)
[![GitHub release](https://img.shields.io/github/v/release/micovery/apigee-go-gen)](https://github.com/micovery/apigee-go-gen/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The `apigee-go-gen` CLI tool streamlines your Apigee development experience using [Go style](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) templates with a YAML centric workflow.

**Here's what you'll find:**

* **[Transformation commands](./transform/index.md)** Easily convert between Apigee's API proxy format and YAML for better readability and management.
* **[Template rendering commands](./render/index.md)**  Enjoy powerful customization and dynamic configuration options, inspired by the flexibility of Helm using the Go [text/template](https://pkg.go.dev/text/template) engine.

By using this tool alongside the [Apigee CLI](https://github.com/apigee/apigeecli), you'll unlock a highly customizable workflow. This is perfect for both streamlined local development and robust CI/CD pipelines.


## Why use YAML 

The traditional Apigee API proxy bundle format has certain characteristics that can present challenges:

* **XML Format**: XML is a polarizing format. While it offers advantages like legacy tooling
  support and well-defined schema validation, its verbosity can make it less  ideal for smaller configuration files.

* **Prescriptive Structure**: The structure of API proxy bundles can feel somewhat rigid, potentially
  limiting flexibility in terms of re-use and customization. This often leads Apigee customers to develop their
  own systems to manage, adapt, and deploy these bundles across environments.

**What if there was a better way?** ...

You can define Apigee API Proxies using YAML configuration files and customize them with a flexible templating system.

This approach has the potential to address the current challenges, offering:

- [x] **Improved readability**
> YAML's streamlined syntax enhances clarity compared to XML.

- [x] **Enhanced flexibility**
> Templating empowers customization and reuse.

- [x] **Repeatable engagements**
> Leverage consistent set of tools to address API proxy use-cases across the business

- [x] **Faster time to production** 
> Leverage Apigee community templates to save time and resources.

- [x] **Stay in sync with API specs** 
> Auto-update templated API Proxies to stay in sync with the spec, while preserving customizations.

## Support

This is not an officially supported Google product.



