# API Proxy to YAML
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

This command takes an Apigee API proxy bundle and converts it into editable YAML document. 

## Usage

The `apiproxy-to-yaml` command takes two parameters `-input` and `-output`.

* `--input` is either from a bundle zip file or an existing bundle directory.

* `--output` is the path for the YAML document to create

* `--output` full path is created if it does not exist (like `mkdir -p`)

Bundle resources are created in the same location as the `--output`

### Examples
Below are a few examples for using the `apiproxy-to-yaml` command.

#### From a zip
Reading bundle from a zip file
```shell
apigee-go-gen transform apiproxy-to-yaml \
  --input ./examples/apiproxies/helloworld/helloworld.zip \
  --output ./out/yaml-first/helloworld1/apiproxy.yaml
```

#### From a dir
Reading bundle from a directory
```shell
apigee-go-gen transform apiproxy-to-yaml \
  --input ./examples/apiproxies/helloworld/ \
  --output ./out/yaml-first/helloworld2/apiproxy.yaml
```


## API Proxy Bundle Structure
In the Apigee world, API proxy bundles are like the packages that hold all the instructions your API needs to work.

- [x] **Creating APIs**
> When you design an API using the Apigee X UI and download the zip file, you're actually getting an API proxy bundle.
- [x] **Deploying APIs**
> Every time you deploy an API in Apigee, you're sending a copy of that bundle to the Apigee runtime to tell it how to handle your API requests.

**What's inside a bundle?**

Think of an API proxy bundle like a neatly organized package. It contains the essential components for your API to function:

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


## YAML Document Structure

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


## YAML Documents with JSONRef

To further enhance organization and reusability, you can use [JSONRef](http://jsonref.org/)s.

This allows you to:

- [x] **Rework Structure**   
> Split the YAML output into smaller, more manageable files.

- [x] **Create Custom Layouts**   
> Arrange components (like policies) in a way that optimizes your workflow.

- [x] **Increase Reusability**  
> Potentially reuse isolated elements in other API Proxies.



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