# Mock OpenAPI Spec
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

You can use the [mock oas](./commands/mock-oas.md) command to create a mock API proxy from your OpenAPI 3 specification, allowing you to simulate API behavior without a real backend. 

## Examples

Below are a couple example of how to use the [mock oas](./commands/mock-oas.md) command

#### Create bundle zip

```shell
apigee-go-gen mock oas \
    --input ./examples/specs/oas3/petstore.yaml \
    --output ./out/mock-apiproxies/petstore.zip
```

#### Create bundle dir
```shell
apigee-go-gen mock oas \
    --input ./examples/specs/oas3/petstore.yaml \
    --output ./out/mock-apiproxies/petstore
```



## Mock API Proxy Features

The generated mock API proxy supports the following features.

### :white_check_mark: Base Path from Spec

The `Base Path` for the mock API proxy is derived from the first element of the [servers](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.1.md#oas-servers) array in your OpenAPI spec.

For example, if your server array looks like this:

```yaml
servers:
  - url: https://petstore.swagger.io/v3
  - url: https://petstore.swagger.io/staging/v3
```

The mock API proxy `Base Path` will be `/v3`

### :white_check_mark: CORS Support

The generated mock API proxy includes the Apigee CORS policy, making it easy to test your API from various browser-based clients.

Here's how it works:

* **Automatic CORS Headers:** The proxy automatically adds the necessary CORS headers (like `Access-Control-Allow-Origin`, `Access-Control-Allow-Methods`, etc.) to all responses.

* **Preflight Requests:** The proxy correctly handles preflight `OPTIONS` requests, responding with the appropriate CORS headers to indicate allowed origins, methods, and headers.

* **Permissive Configuration:** By default, the CORS policy is configured to be as permissive as possible, allowing requests from any origin with any HTTP method and headers. This maximizes flexibility for your testing.

This built-in CORS support ensures that your mock API behaves like a real API in a browser environment, simplifying your development and testing workflow.

### :white_check_mark: Request Validation


By default, the mock API proxy validates the incoming requests against your specification. 
This ensures that the HTTP headers, query parameters, and request body all conform to the defined rules.

This helps you catch errors in your client code early on.

You can disable request validation by passing the header:

```
Mock-Validate-Request: false
```


### :white_check_mark: Dynamic Response Status Code

The mock API proxy automatically generates different status codes for your mock API responses. Here's how it works:

* **Prioritizes success:** If the [operation](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.1.md#operation-object) allows `HTTP 200` status code, the proxy will use it.
* **Random selection:** If `HTTP 200` is not allowed for a particular operation, the proxy will pick a random status code from those allowed.

**Want more control?** You can use headers to select response the status code:

* **Specific status code:** Use the `Mock-Status` header in your request and set it to the desired code (e.g., `Mock-Status: 404`).
* **Random status code:** Use the `Mock-Fuzz: true` header to get a random status code from your spec.

If you use both `Mock-Status` and `Mock-Fuzz`, `Mock-Status` takes precedence.

### :white_check_mark: Dynamic Response Content-Type

The mock API proxy automatically selects the `Content-Type` for responses:

* **JSON preferred:** If the operation allows `application/json`, the proxy will default to using it.
* **Random selection:**  If `application/json` is not available, the proxy will randomly choose from the media types available for that operation.

**Want more control?** You can use headers to select the response Content-Type:

* **Standard `Accept` header:** You can use the standard `Accept` header in your request to request a specific media type (e.g., `Accept: application/xml`).
* **Random media type:**  Alternatively, use the `Mock-Fuzz: true` header to have the proxy select a random media type the available ones.

If you use both `Accept` and `Mock-Fuzz`, the `Accept` header will take precedence.


### :white_check_mark: Dynamic Response Body

The mock API proxy can generate realistic response bodies based on your OpenAPI spec.

Here's how it determines what to send back for any particular operation's response (in order):

1. **Prioritizes `example` field:** If the response includes an `example` field, the proxy will use that example.

2. **Handles multiple `examples`:** If the response has an `examples` field with multiple examples, the proxy will randomly select one.  You can use the `Mock-Example` header to specify which example you want (e.g., `Mock-Example: my-example`).

3. **Uses schema examples:** If no response examples are provided, but the schema for the response has an `example`, the proxy will use that.

4. **Generates from schema:**  As a last resort, the proxy will generate a random example based on the response schema. This works for JSON, YAML, and XML.

You can use the `Mock-Fuzz: true` header to force the proxy to always generate a random example from the schema, even if other static examples are available.


### :white_check_mark: Repeatable API Responses

The mock API proxy uses a special technique to make its responses seem random, while still allowing you to get the same response again if needed. Here's how it works:

* **Pseudo-random numbers:** The "random" choices the proxy makes (like status codes and content) are actually generated using a pseudo-random number generator (PRNG). This means the responses look random, but are determined by a starting value called a "seed."

* **Unique seeds:**  Each request uses a different seed, so responses vary. However, the seed is provided in a special response header called `Mock-Seed`.

* **Getting the same response:** To get an identical response, simply include the `Mock-Seed` header in a new request, using the value from a previous response. This forces the proxy to use the same seed and generate the same "random" choices, resulting in an identical response.

This feature is super helpful for:

* **Testing:**  Ensuring your tests always get the same response.
* **Debugging:** Easily recreating specific scenarios to pinpoint issues in application code.

Essentially, by using the `Mock-Seed` header, you can control the randomness of the mock API responses, making them repeatable for testing and debugging.

### :white_check_mark: Example Generation from JSON Schemas

The following fields are supported when generating examples from a JSON schema:

* `$ref` - local references are followed
* `$oneOf` - chooses a random schema
* `$anyOf` - chooses a random schema
* `$allOf` - combines all schemas
* `object` type
    * `required` field - all required properties are chosen
    * `properties` field - a random set of properties is chosen
    * `additionalProperties` field - only used when there are no `properties` defined
* `array` type
    * `minItems`, `maxItems` fields - array length chosen randomly between these values
    * `items` field  - determines the type of array elements
    * `prefixItems` (not supported yet)
* `null` type
* `const` type
* `boolean` type - true or false randomly chosen
* `string` type
    * `enum` field - a random value is chosen from the list
    * `pattern` field (not supported yet)
    * `format` field
        * `date-time` format
        * `date` format
        * `time` format
        * `email` format
        * `uuid` format
        * `uri` format
        * `hostname` format
        * `ipv4` format
        * `ipv6` format
        * `duration` format
    * `minLength`, `maxLength` fields - string length chosen randomly between these values
* `integer` type
    * `minimum`, `maximum` fields - a random integer value chosen randomly between these values
    * `exclusiveMinimuim` field (boolean, JSON-Schema 4)
    * `exclusiveMaximum` field  (boolean, JSON-Schema 4)
    * `multipleOf` field
* `number` type
    * `minimum`, `maximum` fields - a random float value chosen randomly between these values
    * `exclusiveMinimuim` field (boolean, JSON-Schema 4)
    * `exclusiveMaximum` field  (boolean, JSON-Schema 4)
    * `multipleOf` field

Markdown
## Enriching your OpenAPI Spec with Examples

Sometimes, your OpenAPI specification might be missing response examples or schemas. In other cases, the examples might be very large and difficult to include directly in the spec.  Overlays provide a solution for these situations.

**What is an overlay?**

An overlay is a separate file that allows you to add or modify information in your existing OpenAPI spec. This is useful for adding examples, schemas, or any other data that you want to keep separate from your main specification file. To learn more about how overlays work, you can refer to the [overlay specification](https://www.google.com/url?sa=E&source=gmail&q=link-to-overlay-spec).

**How to use an overlay**

Here's how you can use an overlay to add a static example to an API operation:

1.  **Create an overlay file:** This file defines the changes you want to make to your OpenAPI spec. Here's an example that adds a sample response for the `/pet/findByStatus` operation:

    ```yaml
    overlay: 1.0.0
    info:
      title: Add example response JSON for GET /get/findByStatus
      version: 1.0.0
    actions:
      - target: $.paths./pet/findByStatus.get.responses.200
        update:
          content:
            'application/json':
              example:
                {
                  "id": 1,
                  "photoUrls": [],
                  "name": "Rin Tin Tin",
                  "category": {
                    "id": 1,
                    "name": "Dog"
                  }
                }
    ```
    

2.  **Apply the overlay to your OpenAPI spec:** Use the `apigee-go-gen` tool to combine your overlay file with your OpenAPI spec:

    ```bash
    apigee-go-gen transform oas-overlay \
      --spec ./examples/specs/oas3/petstore.yaml \
      --overlay ./examples/overlays/petstore-dog-example.yaml \
      --output ./out/specs/petstore-overlaid.yaml
    ```

3.  **Generate a mock API proxy:** You can now use the updated OpenAPI spec to generate a mock API proxy in Apigee:

    ```bash
    apigee-go-gen mock oas \
        --input ./out/specs/petstore-overlaid.yaml \
        --output ./out/mock-apiproxies/petstore.zip
    ```

This process allows you to manage your OpenAPI spec more effectively by keeping your examples and other supplementary data in separate files.