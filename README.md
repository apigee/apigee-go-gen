## Apigee YAML Toolkit 


**Welcome!** This repo offers a set of tools to streamline your Apigee API Proxy development experience using a YAML-First approach.

**Here's what you'll find:**

* **Transformation tools:** Easily convert between Apigee's API Proxy format and YAML for better readability and management.
* **Templating tools:**  Enjoy powerful customization and dynamic configuration options, inspired by the flexibility of Helm using the Go [text/template](https://pkg.go.dev/text/template) engine.

By using these tools alongside the Apigee CLI, you'll unlock a highly customizable YAML-First workflow. This is perfect for both streamlined local development and robust CI/CD pipelines.


## Table of Contents

* [Why use YAML-First](#why-use-yaml-first)
* [Understanding API Proxy Bundles](#understanding-api-proxy-bundles)
* [Transformation Tools](#transformation-tools)
  * [xml2yaml](#tool-xml2yaml)
  * [yaml2xml](#tool-yaml2xml)
  * [bundle2yaml](#tool-bundle2yaml)
  * [yaml2bundle](#tool-yaml2bundle)
* [Template Rendering Tools](#template-rendering-tools)
  * [render-oas](#tool-render-oas)
  * [render-graphql](#tool-render-graphql)
  * [render-grpc](#tool-render-grpc)
  * [render-template ](#tool-render-template)
* [Installation](#installation)

    
## Why use YAML-First

The API Proxy bundle format has certain characteristics that can present challenges:

* **XML Format**: XML is a polarizing format. While it offers advantages like legacy tooling 
support and well-defined schema validation, its verbosity can make it less  ideal for smaller configuration files.

* **Prescriptive Structure**: The structure of API Proxy bundles can feel somewhat rigid, potentially 
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


In the Apigee world, API Proxy bundles are like the packages that hold all the instructions your API needs to work. Here's a simple breakdown:

* **Creating APIs:** When you design an API using the Apigee X UI and download the zip file, you're actually getting an API Proxy bundle.
* **Deploying APIs:** Every time you deploy an API in Apigee, you're sending a copy of that bundle to the Apigee system to tell it how to handle your API requests.

**What's inside a bundle?**

Think of an API Proxy bundle like a neatly organized package. It contains the essential components for your API to function:

* **Instructions for handling requests:** This part tells Apigee how to respond to the different types of requests your API might receive.
* **Policies for security and management:** These define how to protect your API, control access, and even things like tracking usage.
* **And more...** depending on your setup!

Below is the general structure of a bunle

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

I've only showed some of the resources that could exist in a `bundle`. There are other a few others, that are not used as frequently.
(i.e. for xsds, wsdls and others)

## Transformation Tools

This toolkit offers a powerful set of conversion tools for your Apigee development workflow.

**XML ↔️ YAML Conversion:** Easily switch between XML and YAML formats

  * `yaml2xml` - Converts a snippet of XML to YAML snippet
  * `xml2yaml` - Converts a snippet of XML to YAML snippet


**API Proxy Bundles ↔️ YAML-First API Proxies:**  Convert between traditional API Proxy bundles and a more user-friendly YAML representation.

* `bundle2yaml` - Converts an API Proxy bundle to a YAML doc
* `yaml2bundle` - Converts a YAML doc to an API Proxy bundle

### Tool: xml2yaml

**Purpose:** This tool takes XML snippets and effortlessly converts them into YAML.

**Why does this matter?** Let's say you have an Apigee policy written in XML format. Instead of manually retyping the whole thing into YAML, you can simply use this tool for instant conversion. This is especially handy when you're working with examples from the Apigee documentation – just copy, paste, convert!

The tool follows a reliable set of rules to transform your XML into clean YAML (see  [docs/rules.md](docs/rules.md))


**Usage**

The tool reads XML form `stdin`, and writes YAML to `stdout`. Below are a couple of examples.

  * Reading input redirected from a file
    ```shell
    xml2yaml < ./examples/snippets/check-quota.xml
    ```
  * Reading input directly from stdin
    ```shell
    xml2yaml << EOF 
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
    </Parent>' | xml2yaml
    ```
    > **NOTE:** Using echo as shown above will not work properly if your input already contains single quotes


### Tool: yaml2xml

**Purpose:** This tool converts YAML snippets to XML for quick validation and troubleshooting.

Here's how it helps:

* **Verify:** Ensure your YAML translates correctly to XML
* **Debug:** Compare the XML output to catch errors in your YAML.

**Usage**

The tool reads YAML form `stdin`, and writes XML to `stdout`. Below are a few examples.

  * Reading input redirected from a file
    ```shell
    yaml2xml < ./examples/snippets/ducks.yaml
    ```
  * Reading input from `stdin` directly
    ```shell
    yaml2xml << EOF
    Parent:
      - Child: Fizz
      - Child: Buzz' | yaml2xml
    EOF
    ```
  * Reading input piped from another process
    ```shell
    echo '
    Parent:
      - Child: Fizz
      - Child: Buzz' | yaml2xml
    ```
    > **NOTE:** Using echo as shown above will not work properly if your input already contains single quotes

### Tool: bundle2yaml

**Purpose:** This tool takes existing API Proxy bundles and transforms them into editable YAML documents. This offers several advantages:

* **Enhanced Customization:** Tweak your API Proxy configurations with the readability of YAML.
* **Workflow Integration:** YAML's compatibility opens up possibilities for version control and automation.
* **YAML-First Transition:** Start using a YAML-First approach with your existing API Proxies.


**YAML Document Structure**

The YAML document created by `bundle2yaml` contains all the elements from the bundle in a single file.

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

The tool creates resource files alongside the YAML doc.

**Usage**

The `bundle2yaml` tool takes two parameters `-input` and `-output`.

* `-input` is either from a bundle zip file or an existing bundle directory. 

* `-output` is the path for the YAML document to create

* `-output` full path is created if it does not exist (like `mkdir -p`)

Bundle resources are created in the same location as the `-output`

Below are a couple of examples


  * Reading bundle from a zip file
    ```shell
    bundle2yaml -input ./examples/bundles/helloworld/helloworld.zip \
                -output ./out/yaml-first/helloworld1/apiproxy.yaml
    ```
  * Reading bundle from a directory
    ```shell
    bundle2yaml -input ./examples/bundles/helloworld/ \
                -output ./out/yaml-first/helloworld2/apiproxy.yaml
    ```


**YAML Documents With JSONRefs**

The `bundle2yaml` tool offers a starting point by converting your API Proxy bundle into a single YAML document. 

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

### Tool: yaml2bundle

**Purpose:** This tool takes an existing YAML document and converts it into a ready-to-use API Proxy bundle.

This tool plays a crucial part in streamlining your Apigee development process. Here's how it works:

1. **Design:** Craft your API Proxy configuration using the more readable and manageable YAML format. 
2. **Convert:** Feed your YAML document into the tool to get a fully compliant API Proxy bundle.
3. **Deploy:** Use the Apigee CLI to deploy the bundle

**Usage**

The `yaml2bundle` tool takes two parameters `-input` and `-output`

* `-input` is the YAML document that contains the API Proxy definitions

* `-output` is either a bundle zip or a bundle directory to be created

* `-output` full path is created if it does not exist (like `mkdir -p`)

Bundle resources are read relative to the location of the `-input` 

Below are a couple of examples

  * Creating a bundle zip
    ```shell
    yaml2bundle -input ./examples/yaml-first/petstore/apiproxy.yaml \
                -output ./out/bundles/petstore.zip 
    ```
  * Creating a bundle directory
    ```shell
    yaml2bundle -input ./examples/yaml-first/petstore/apiproxy.yaml \
                -output ./out/bundles/petstore
    ```

## Template Rendering Tools

This toolkit includes tools that let you create API Proxy bundles based on popular formats like
OpenAPI, GraphQL, and gRPC. Think of them as blueprints for your API Proxies.

* `render-template` - Renders a [Go-style](https://pkg.go.dev/text/template) template
* `render-oas` - Renders a [Go-style](https://pkg.go.dev/text/template) template using an OpenAPI spec context
* `render-graphql` - Renders a [Go-style](https://pkg.go.dev/text/template) template using a GraphQL schema context
* `render-grpc` - Renders a [Go-style](https://pkg.go.dev/text/template) template using a gRPC proto context


**Why use templates?**

* **Tailored Control:** Craft your API Proxies exactly how you need them, from security policies to dynamic 
behavior based on environment variables.

* **Unlock Your Specs:** Don't just describe your API – use the information in your OpenAPI specs (and others) to
automatically build out parts of your API Proxy configuration.

Imagine easily adding security rules, setting target URLs based on your setup, or even having your API proxy structure adjust to match your API specifications. These tools make that possible!

**The template language**

The tools here use Go [text/template](https://pkg.go.dev/text/template) engine behind the scenes to render the input template.
The Go templating engine is very powerful and gives you lots of flexibility and features like 
loop constructs, conditional elements, template blocks for re-use and much more. 



### Tool: render-oas

**Purpose:** This tool takes your OpenAPI spec and a customizable template, generating an intermediate YAML configuration for your Apigee API Proxy.

**How it Works:**

* **Start with Your Template:** This is your baseline. Include any standard policies or settings you want in your final proxy.
* **Customize the Output:** Your template uses special placeholders that the tool will replace with details from your OpenAPI spec.
* **Control the Output:** Use control logic in your template to adjust your proxy configuration based on your OpenAPI spec.
* **Access the Spec:** The OpenAPI text and [map](https://go.dev/blog/maps) are available during template rendering.

Once you render the template, you then use the `yaml2bundle` tool to transform this YAML output into a deployable API Proxy bundle.

**See an Example:** Check out the included OAS3 template at [examples/template/oas3](examples/templates/oas3/apiproxy.yaml). 

It demonstrates the basics of how the tool creates Apigee-compatible YAML from your OpenAPI 3 spec.

Here is how you would call it

```shell
render-oas -template ./examples/templates/oas3/apiproxy.yaml \
           -spec ./examples/specs/petstore.yaml \
           -include ./examples/templates/oas3/*.tmpl \
           -output ./out/yaml-first/petstore/apiproxy.yaml
```

> [!NOTE]
> You may pass the `-include` flag multiple time to load template helpers from multiple sources.



**Quick Template Previews with Dry Run**

For streamlined development, you can view the rendered template output directly in your terminal. This avoids writing to disk during your iterative process. Add the `-dry-run true` flag:

```shell
render-oas -template ./examples/templates/oas3/apiproxy.yaml \
           -spec ./examples/specs/petstore.yaml \
           -include ./examples/templates/oas3/*.tmpl \
           -dry-run true
```


### Tool: render-graphql

**Purpose:** This tool is used for rendering a template using a GraphQL schema as input.

While GraphQL schemas might not contain all the necessary details for a complete API Proxy bundle, this tool offers flexibility through the `--set` and `--set-string` parameters. This works similar to how values are set in Helm charts.

**How it Works:**

* **Start with Your Template:** The template, your inputs, and the schema guide the generation of an intermediate YAML configuration.
* **Inject Your Values:** Use `--set` and `--set-string` to provide missing values (like target URLs) for your template.
* **Access the Schema:** The GraphQL schema text and [AST](https://pkg.go.dev/github.com/vektah/gqlparser/v2/ast#Schema) are available during template rendering.


**See an Example:** Check out the [examples/templates/graphql](examples/templates/graphql) directory for an example of building the intermediate YAML for a GraphQL API Proxy.

Here is how you would call it

```shell
render-graphql -template ./examples/templates/graphql/apiproxy.yaml \
               -schema ./examples/graphql/resorts.graphql \
               -set-string "api_name=resorts-api" \
               -set-string "base_path=/graphql" \
               -set-string "target_url=https://example.com/graphql" \
               -include ./examples/templates/graphql/*.tmpl \
               -output ./out/yaml-first/resorts/apiproxy.yaml
``` 

### Tool: render-grpc

**Purpose:** This tool is used for rendering a template using a gRPC proto file as input.

When working with gRPC in Apigee, it's crucial to ensure your API Proxy's base path and conditional flows are configured 
correctly to handle gRPC traffic. This tool simplifies the process by letting you build a template that understands 
these gRPC-specific requirements. 


**How it Works:**

* **Start with Your Template:** Input your gRPC proto file, and the template generates the intermediate YAML configuration.
* **Automate the Details:** The template handles the intricacies of gRPC integration within your API Proxy.
* **Access the Proto:** The proto text and [descriptor](https://pkg.go.dev/google.golang.org/protobuf/types/descriptorpb#FileDescriptorProto) are available during template rendering.


Any additional information (such as target server name) that is not available within the proto file
can be supplied as values tom the rendering process using the `-set` and `-set-string` params.


**See an Example:** Check out the [examples/templates/grpc](examples/templates/grpc) directory for an example of building the intermediate YAML for a gRPC API Proxy.

Here is how you would use the tool with this example:

```shell
render-grpc -template ./examples/templates/grpc/apiproxy.yaml \
            -proto ./examples/protos/greeter.proto \
            -set-string "target_server=example-target-server" \
            -include ./examples/templates/grpc/*.tmpl \
            -output ./out/yaml-first/greeter/apiproxy.yaml
```

### Tool: render-template

**Purpose:** This is the generic version of the rendering tools

If you're working with a spec format beyond OpenAPI, GraphQL, or gRPC, this tool gives you the freedom to design your own templates for
generating Apigee API Proxies. Here's how it works:

* **Start with Your Template:** Use the familiar Go templating syntax, enhanced with helpful functions specifically for building API Proxies.
* **Inject Your Values:** Use `-set` and `-set-string` to provide essential details that your template will use during the rendering process.

> [!NOTE]
> All rendering tools in this toolkit use the same underlying Go templating logic (including helper functions)

For a full list of all available helper functions, see [helper_functions.txt](pkg/common/resources/helper_functions.txt)


## Installation

If you already have [Go](https://go.dev/doc/install) installed in your machine, run the following command:

```shell
go install github.com/micovery/apigee-yaml-toolkit/cmd/...@latest
```

This will download, build and install all the tools into your `$GOPATH/bin` directory

You can change the `@latest` tag for any other version that has been tagged. (e.g. `@v0.1.1`)

> [!NOTE]
> The Go tool (and compiler) is only necessary to build the tools in this repo. 
> Once built, you can copy the tool binaries and use them in any other
> machine of the same architecture and operating system (without needing Go).

## Support
This is not an officially supported Google product
