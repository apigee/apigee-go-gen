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
# Apigee Go Gen

[![Go Report Card](https://goreportcard.com/badge/github.com/micovery/apigee-yaml-toolkit)](https://goreportcard.com/report/github.com/micovery/apigee-yaml-toolkit)
[![GitHub release](https://img.shields.io/github/v/release/micovery/apigee-yaml-toolkit)](https://github.com/micovery/apigee-yaml-toolkit/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The `apigee-go-gen` tool streamlines your Apigee development experience using [Go style](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) templates with a [YAML-First](#why-use-yaml-first) approach.

**Here's what you'll find:**

* **Transformation commands:** Easily convert between Apigee's API proxy format and YAML for better readability and management.
* **Templating commands:**  Enjoy powerful customization and dynamic configuration options, inspired by the flexibility of Helm using the Go [text/template](https://pkg.go.dev/text/template) engine.

By using this tool alongside the [Apigee CLI](https://github.com/apigee/apigeecli), you'll unlock a highly customizable workflow. This is perfect for both streamlined local development and robust CI/CD pipelines.


## Table of Contents

* [Why use YAML-First](#why-use-yaml-first)
* [Understanding API Proxy Bundles](#understanding-api-proxy-bundles)
* [Transformation Commands](#transformation-commands)
  * [xml-to-yaml](#xml-to-yaml-command)
  * [yaml-to-xml](#yaml-to-xml-command)
  * [apiproxy-to-yaml](#api-proxy-to-yaml-command)
  * [yaml-to-apiproxy](#yaml-to-api-proxy-command)
  * [sharedflow-to-yaml](#shared-flow-to-yaml-command)
  * [yaml-to-sharedflow](#yaml-to-shared-flow-command)
  * [sharedflow-to-yaml](#shared-flow-to-yaml-command)
* [Template Rendering Commands](#template-rendering-commands)
  * [Using OpenAPI Spec](#creating-api-proxy-from-openapi-spec)
  * [Using GraphQL Schema](#creating-api-proxy-from-graphql-schema)
  * [Using gRPC Proto](#creating-api-proxy-from-grpc-proto)
  * [Using Other Formats](#creating-api-proxy-from-other-formats)
* [Installation](#installation)

    
## Why use YAML-First

The API proxy bundle format has certain characteristics that can present challenges:

* **XML Format**: XML is a polarizing format. While it offers advantages like legacy tooling 
support and well-defined schema validation, its verbosity can make it less  ideal for smaller configuration files.

* **Prescriptive Structure**: The structure of API proxy bundles can feel somewhat rigid, potentially 
limiting flexibility in terms of re-use and customization. This often leads Apigee customers to develop their 
own systems to manage, adapt, and deploy these bundles across environments.

**What about tools like the Apigee CLI, and the Apigee Maven plugins?**

While tools like the Apigee CLI and Maven plugins offer valuable abstractions on top of the Apigee REST API, there's still room to expand their capabilities 
for generating proxy bundles from formats like gRPC protos, GraphQL schemas, or OpenAPI.

Additionally, streamlining customization of existing bundles is usually required. To address these areas, Apigee customers often develop their own tailored solutions like generators or custom DSLs, demonstrating a clear need for 
enhanced flexibility in API proxy bundle creation and modification.

What if there was a better way? I propose using YAML to represent API Proxies in combination with a powerful Go-style templating engine. 
This approach has the potential to address the current challenges, offering:

* **Improved Readability**: YAML's streamlined syntax enhances clarity compared to XML.


* **Enhanced Flexibility**: Templating empowers customization and reuse.


## Understanding API Proxy Bundles


In the Apigee world, API proxy bundles are like the packages that hold all the instructions your API needs to work. Here's a simple breakdown:

* **Creating APIs:** When you design an API using the Apigee X UI and download the zip file, you're actually getting an API proxy bundle.
* **Deploying APIs:** Every time you deploy an API in Apigee, you're sending a copy of that bundle to the Apigee system to tell it how to handle your API requests.

**What's inside a bundle?**

Think of an API proxy bundle like a neatly organized package. It contains the essential components for your API to function:

* **Instructions for handling requests:** This part tells Apigee how to respond to the different types of requests your API might receive.
* **Policies for security and management:** These define how to protect your API, control access, and even things like tracking usage.
* **And more...** depending on your setup!

Below is the general structure of a bundle

```

# One metadata file
./approxy/
./apiproxy/proxy-name.xml

# One or more Proxy Endpoints (each with its own base path)
./apiproxy/proxies
./apiproxy/proxies/proxy1.xml
./apiproxy/proxies/proxy2.xml
./apiproxy/proxies/etc.xml

# Optional Target Endpoints (pointed to by RouteRules)
./apiproxy/targets/
./apiproxy/targets/target1.xml
./apiproxy/targets/target2.xml
./apiproxy/targets/etc.xml

# Optional policy files (for use in Flows Steps)
./apiproxy/policies
./apiproxy/policies/policy1.xml
./apiproxy/policies/policy2.xml
./apiproxy/policies/etc.xml

# Optional JavaScript files (for use with the JS Policy)
./apiproxy/resources/
./apiproxy/resources/jsc/
./apiproxy/resources/jsc/script.js

# Optional Java jar files (for use with the Java callout policy)
./apiproxy/resources/java/
./apiproxy/resources/java/library.jar

# Optional OpenAPI spec files (for use with the OAS policy)
./apiproxy/resources/oas
./apiproxy/resources/oas/openapi.yaml

# Optional GraphQL schema files (for use with GraphQL policy)
./apiproxy/resources/graphql/
./apiproxy/resources/graphql/schema.graphql

# Optional properties files
./apiproxy/resources/properties/
./apiproxy/resources/properties/values.properties
```

There are more resource types such as xsds, wsdls, and others. (see [docs](https://cloud.google.com/apigee/docs/api-platform/develop/resource-files))


## Transformation Commands

The `apigee-go-gen` offers a powerful set of `transform` commands for your Apigee development workflow.

**XML ↔️ YAML Conversion:** Easily switch between XML and YAML formats

  * `yaml-to-xml` - Transforms a snippet of XML to YAML snippet
  * `xml-to-yaml` - Transforms a snippet of XML to YAML snippet


**API Proxy Bundles ↔️ YAML-First API Proxies:**  Convert between traditional API proxy bundles and a more user-friendly YAML representation.

* `apiproxy-to-yaml` - Transforms an API proxy bundle to a YAML doc
* `yaml-to-apiproxy` - Transforms a YAML doc to an API proxy bundle

### XML to YAML Command

**Purpose:** This command takes XML snippets and effortlessly converts them into YAML.

**Why does this matter?** Let's say you have an Apigee policy written in XML format. Instead of manually retyping the whole thing into YAML, you can simply use this tool for instant conversion. This is especially handy when you're working with examples from the Apigee documentation – just copy, paste, convert!

The command follows a reliable set of rules to transform your XML into clean YAML (see  [docs/rules.md](docs/rules.md))


**Usage**

The command reads XML form `stdin`, and writes YAML to `stdout`. Below are a couple of examples.

  * Reading input redirected from a file
    ```shell
    apigee-go-gen transform xml-to-yaml < ./examples/snippets/check-quota.xml
    ```
  * Reading input directly from stdin
    ```shell
    apigee-go-gen transform xml-to-yaml << EOF 
    <Parent>
      <Child>Fizz</Child>
      <Child>Buzz</Child>
    </Parent>
    EOF
    ```
  * Reading input piped from another process
    ```shell
    echo '
    <Parent>
      <Child>Fizz</Child>
      <Child>Buzz</Child>
    </Parent>' | apigee-go-gen transform xml-to-yaml
    ```
    > **NOTE:** Using echo as shown above will not work properly if your input already contains single quotes


### YAML to XML Command

**Purpose:** This command converts YAML snippets to XML for quick validation and troubleshooting.

Here's how it helps:

* **Verify:** Ensure your YAML translates correctly to XML
* **Debug:** Compare the XML output to catch errors in your YAML.

**Usage**

The command reads YAML form `stdin`, and writes XML to `stdout`. Below are a few examples.

  * Reading input redirected from a file
    ```shell
    apigee-go-gen transform yaml-to-xml < ./examples/snippets/ducks.yaml
    ```
  * Reading input from `stdin` directly
    ```shell
    apigee-go-gen transform yaml-to-xml << EOF
    Parent:
      - Child: Fizz
      - Child: Buzz
    EOF
    ```
  * Reading input piped from another process
    ```shell
    echo '
    Parent:
      - Child: Fizz
      - Child: Buzz' | apigee-go-gen transform yaml-to-xml
    ```
    > **NOTE:** Using echo as shown above will not work properly if your input already contains single quotes

### API Proxy to YAML Command

**Purpose:** This command takes existing API proxy bundles and transforms them into editable YAML documents. This offers several advantages:

* **Enhanced Customization:** Tweak your API proxy configurations with the readability of YAML.
* **Workflow Integration:** YAML's compatibility opens up possibilities for version control and automation.
* **YAML-First Transition:** Start using a YAML-First approach with your existing API Proxies.


**YAML Document Structure**

The YAML document created by `apiproxy-to-yaml` contains all the elements from the bundle in a single file.

The structure looks like this

```yaml
# From ./apiproxy/proxy-name.xml
APIProxy:
  .name: hello-world
  .revision: 1
  #...

# From ./apiproxy/policies/*
Policies: 
  - AssignMessage: 
      .name: AM-SetTarget
      #...
  - RaiseFault:
      .name: RF-Set500
      #...

# From ./apiproxy/proxies/*
ProxyEndpoints: 
  - ProxyEndpoint: 
      .name: proxy1
      #...
  - ProxyEndpoint: 
      .name: proxy2
      #...

# From ./apiproxy/targets/*
TargetEndpoints: 
  - TargetEndpoint: 
      .name: target1
      #...
  - TargetEndpoint: 
      .name: target2
      #...

# From ./apiproxy/resources/*/* 
Resources: 
  - Resource:
      Type: "properties"
      Path: "./path/to/resource.properties"
  - Resource: 
      Type: "jsc"
      Path: "./path/to/script.js"
```

The command creates resource files alongside the YAML doc.

**Usage**

The `apiproxy-to-yaml` command takes two parameters `-input` and `-output`.

* `--input` is either from a bundle zip file or an existing bundle directory. 

* `--output` is the path for the YAML document to create

* `--output` full path is created if it does not exist (like `mkdir -p`)

Bundle resources are created in the same location as the `--output`

Below are a couple of examples


  * Reading bundle from a zip file
    ```shell
    apigee-go-gen transform apiproxy-to-yaml \
        --input ./examples/apiproxies/helloworld/helloworld.zip \
        --output ./out/yaml-first/helloworld1/apiproxy.yaml
    ```
  * Reading bundle from a directory
    ```shell
    apigee-go-gen transform apiproxy-to-yaml \
        --input ./examples/apiproxies/helloworld/ \
        --output ./out/yaml-first/helloworld2/apiproxy.yaml
    ```


**YAML Documents With JSONRefs**

The `apiproxy-to-yaml` command offers a starting point by converting your API proxy bundle into a single YAML document. 

To further enhance organization and reusability, you can use [JSONRef](http://jsonref.org/)s. This allows you to:

* **Rework Structure:** Split the YAML output into smaller, more manageable files.
* **Create Custom Layouts:** Arrange components (like policies) in a way that optimizes your workflow.
* **Increase Reusability:** Potentially reuse isolated elements in other API Proxies.



Below is an example for moving policies to a separate file

e.g.

```yaml
#...

Policies: 
  $ref: ./policies.yaml#/
#...
```

The new `policies.yaml` file would look like this:

```yaml
- AssignMessage:
    .name: AM-SetTarget
    #...
- RaiseFault:
    .name: RF-Set500
    #...
```

### YAML to API Proxy Command

**Purpose:** This command takes an existing YAML document and converts it into a ready-to-use API proxy bundle.

This command plays a crucial part in streamlining your Apigee development process. Here's how it works:

1. **Design:** Craft your API proxy configuration using the more readable and manageable YAML format. 
2. **Convert:** Feed your YAML document into the command to get a fully compliant API proxy bundle.
3. **Deploy:** Use the Apigee CLI to deploy the bundle

**Usage**

The `yaml-to-apiproxy` command takes two parameters `-input` and `-output`

* `--input` is the YAML document that contains the API proxy definitions

* `--output` is either a bundle zip or a bundle directory to be created

* `--output` full path is created if it does not exist (like `mkdir -p`)

Bundle resources are read relative to the location of the `--input` 

Below are a couple of examples

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


### Shared Flow to YAML Command

**Purpose:** This command takes existing shared flow bundles and transforms them into editable YAML documents. 


**YAML Document Structure**

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

The command creates resource files alongside the YAML doc.

**Usage**

The `sharedflow-to-yaml` command takes two parameters `-input` and `-output`.

* `--input` is either from a bundle zip file or an existing bundle directory.

* `--output` is the path for the YAML document to create

* `--output` full path is created if it does not exist (like `mkdir -p`)

Bundle resources are created in the same location as the `--output`

Below are a couple of examples


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



### YAML to Shared Flow Command

**Purpose:** This command takes an existing YAML document and converts it into a ready-to-use shared flow bundle.

**Usage**

The `yaml-to-sharedflow` command takes two parameters `-input` and `-output`

* `--input` is the YAML document that contains the shared flow definitions

* `--output` is either a bundle zip or a bundle directory to be created

* `--output` full path is created if it does not exist (like `mkdir -p`)

Bundle resources are read relative to the location of the `--input`

Below are a couple of examples

* Creating a bundle zip
  ```shell
  apigee-go-gen transform yaml-to-sharedflow \
      --input ./examples/yaml-first/owasp/sharedflow.yaml \
      --output ./out/sharedflows/owasp.zip 
  ```
* Creating a bundle directory
  ```shell
  apigee-go-gen transform yaml-to-sharedflow \
      --input ./examples/yaml-first/owasp/sharedflow.yaml \
      --output ./out/sharedflows/owasp
  ```


## Template Rendering Commands

The `apigee-go-gen` includes a set of `render` commands that let you create API proxy bundles based on popular formats like
OpenAPI, GraphQL, gRPC, and more using templates. Think of these templates as blueprints for your API Proxies.

* `render template` - Renders a [Go-style](https://pkg.go.dev/text/template) template
* `render apiproxy` - Combines `render template` and `yaml-to-apiproxy` into one
* `render sharedflow` - Combines `render template` and `yaml-to-sharedflow` into one


**Why use templates?**

* **Tailored Control:** Craft your API Proxies exactly how you need them, from security policies to dynamic 
behavior based on environment variables.

* **Unlock Your Specs:** Don't just describe your API – use the information in your OpenAPI specs (and others) to
automatically build out parts of your API proxy configuration.

Imagine easily adding security rules, setting target URLs based on your setup, or even having your API proxy structure adjust to match your API specifications. These tools make that possible!

### The Template Language

The `render` commands are powered by the Go [text/template](https://pkg.go.dev/text/template) engine.

Some popular tool that also use this same engine are [Helm](https://helm.sh/), and [Hugo](https://gohugo.io/).

Below are a few examples of how to use the `render apiproxy` command for generating API Proxies from OAS, GraphQL, and GRPC.

### Creating API Proxy from OpenAPI Spec

You can use the `render apiproxy` command to create a bundle using a template and an OpenAPI spec as input.

**How it Works:**

* **Start with Your Template:** This is your baseline. Include any standard policies or settings you want in your final proxy.
* **Customize the Output:** Your template uses special placeholders that the tool will replace with details from your OpenAPI spec.
* **Control the Output:** Use control logic in your template to adjust your proxy configuration based on your OpenAPI spec.
* **Access the Spec:** Use `--set-oas` to access the OpenAPI spec text and [map](https://go.dev/blog/maps) during template rendering.


**See an Example:** Check out the included OAS3 template at [examples/template/oas3](examples/templates/oas3/apiproxy.yaml). 

It demonstrates the basics of how the command creates API proxy bundles from your OpenAPI 3 spec.

Here is how you would call it

```shell
apigee-go-gen render apiproxy \
    --template ./examples/templates/oas3/apiproxy.yaml \
    --set-oas spec=./examples/specs/petstore.yaml \
    --include ./examples/templates/oas3/*.tmpl \
    --output ./out/apiproxies/petstore.zip
```

> [!NOTE]
> You may pass the `-include` flag multiple time to load template helpers from multiple sources.



**Quick Template Previews with Dry Run**

For streamlined development, you can view the rendered template output directly in your terminal. This avoids writing to disk during your iterative process. Add the `--dry-run xml` or `--dry-run yaml` flag:

```shell
apigee-go-gen render apiproxy \
    --template ./examples/templates/oas3/apiproxy.yaml \
    --set-oas spec=./examples/specs/petstore.yaml \
    --include ./examples/templates/oas3/*.tmpl \
    --dry-run xml
```


### Creating API Proxy from GraphQL Schema

You can use the `render apiproxy` command to create a bundle using a template and a GraphQL schema as input.

While GraphQL schemas might not contain all the necessary details for a complete API proxy bundle, this command offers flexibility through the `--set` and `--set-string` parameters. 
This works similar to how values are set in Helm charts.

**How it Works:**

* **Start with Your Template:** The template, your inputs, and the schema guide the generation of an intermediate YAML configuration.
* **Access the Schema:**  Use `--set-graphql` to access a GraphQL schema text and [AST](https://pkg.go.dev/github.com/vektah/gqlparser/v2/ast#Schema) during template rendering.
* **Inject Your Values:** Use `--set` and `--set-string` to provide missing values (like target URLs) for your template.


**See an Example:** Check out the [examples/templates/graphql](examples/templates/graphql) directory for an example of building the intermediate YAML for a GraphQL API proxy.

Here is how you would call it

```shell
apigee-go-gen render apiproxy \
     --template ./examples/templates/graphql/apiproxy.yaml \
     --set-graphql schema=./examples/graphql/resorts.graphql \
     --set-string "api_name=resorts-api" \
     --set-string "base_path=/graphql" \
     --set-string "target_url=https://example.com/graphql" \
     --include ./examples/templates/graphql/*.tmpl \
     --output ./out/apiproxies/resorts.zip
``` 


### Creating API Proxy from gRPC Proto

You can use the `render apiproxy` command to create bundle using a template and a gRPC proto file as input.

When working with gRPC in Apigee, it's crucial to ensure your API proxy's base path and conditional flows are configured 
correctly to handle gRPC traffic. This command simplifies the process by letting you build a template that understands 
these gRPC-specific requirements. 


**How it Works:**

* **Start with Your Template:** Input your gRPC proto file, and the template generates the intermediate YAML configuration.
* **Automate the Details:** The template handles the intricacies of gRPC integration within your API proxy.
* **Access the Proto:** Use `--set-grpc` to access the gRPC proto text and [descriptor](https://pkg.go.dev/google.golang.org/protobuf/types/descriptorpb#FileDescriptorProto) during template rendering.
* **Inject Your Values:** Use `--set` and `--set-string` to provide missing values (like target server) for your template.


**See an Example:** Check out the [examples/templates/grpc](examples/templates/grpc) directory for an example of building the intermediate YAML for a gRPC API proxy.

Here is how you would use the command with this example:

```shell
apigee-go-gen render apiproxy \
    --template ./examples/templates/grpc/apiproxy.yaml \
    --set-grpc proto=./examples/protos/greeter.proto \
    --set-string "target_server=example-target-server" \
    --include ./examples/templates/grpc/*.tmpl \
    --output ./out/apiproxies/greeter.zip
```

### Creating API Proxy from other formats

The `render apiproxy` command can be used to create a bundle from any template.
It's not necessary to start from an OpenAPI spec, GraphQL schema, or gRPC proto. 

You can use flags such as `--set` and `--set-string` to dynamically provide values for the template.
These values are available during the rendering process using `{{ $.Values.key }}`.

This allows you to create templates, and dynamically generate API proxy bundles based
on your specific requirements.


## Installation

If you already have [Go](https://go.dev/doc/install) installed in your machine, run the following command:

```shell
go install github.com/micovery/apigee-yaml-toolkit/cmd/...@latest
```

This will download, build and install the `apigee-go-gen` into your `$GOPATH/bin` directory

You can change the `@latest` tag for any other version that has been tagged. (e.g. `@v0.1.8`)

> [!NOTE]
> The Go tool (and compiler) is only necessary to build the tools in this repo. 
> Once built, you can copy the tool binaries and use them in any other
> machine of the same architecture and operating system (without needing Go).

## Support
This is not an officially supported Google product
