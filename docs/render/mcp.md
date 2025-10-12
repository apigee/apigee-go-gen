# MCP

!!! Note

    This MCP functionality in `apigee-go-gen` is currently in **beta** and under active development. To use this feature, you must install the latest beta build of the tool.

    The MCP specification itself is evolving. This implementation strives to align with the latest version of the [MCP spec](https://modelcontextprotocol.io/specification/latest).

    If you have feedback or encounter bugs, please open an issue on GitHub [apigee-go-gen/issues](https://github.com/apigee/apigee-go-gen/issues).

You can use the `render apiproxy` command to convert an [OpenAPI 3.x](https://openapis.org) description into an API proxy that functions as an [MCP](https://modelcontextprotocol.io) server. 

This process is useful for exposing operations from an existing REST API as MCP tools. 

This is sometimes referred to as "MCPfying" a REST API.

---

## How It Works

- [x] **Provide an OpenAPI Description**
> Start with an existing OpenAPI 3.x description of your REST API.

- [x] **Get the MCP Template**
> Use baseline MCP template is provided in the [`examples/templates/mcp/`](https://github.com/apigee/apigee-go-gen/blob/main/examples/templates/mcp/apiproxy.yaml) directory. You can use it as-is or customize it to add specific policies or settings to your final API proxy.

- [x] **Generate the MCP API proxy**
> Use the `render apiproxy` command to generate a deployable MCP API proxy bundle from your OpenAPI description and the template.

- [x] **Deploy the API proxy**
> Use the [apigeecli](https://github.com/apigee/apigeecli) tool to deploy the generated API proxy bundle to your Apigee runtime environment.

- [x] **Use the MCP Server**
> Once deployed, the API proxy acts as an MCP server. The operations from your original OpenAPI description are now exposed as MCP tools.

---

## Example

This example demonstrates how to convert the `weather.yaml` OpenAPI description into an Apigee API proxy that serves MCP tools.

### 1. Create the MCP API proxy

Generate the API proxy bundle using the `render apiproxy` command. This command uses the MCP template and the sample `weather.yaml` OpenAPI spec.

```shell
apigee-go-gen render apiproxy \
    --template ./examples/templates/mcp/apiproxy.yaml \
    --include ./examples/templates/mcp/*.tmpl \
    --set-oas spec=./examples/specs/oas3/weather.yaml \
    --set base_path=/mcp/weather \
    --output ./out/apiproxies/weather.zip
```
Note that the `--set base_path` flag overrides the base path defined in the OpenAPI spec, setting it to `/mcp/weather` for this proxy.

### 2. Deploy the MCP API proxy

Use `apigeecli` to deploy the generated API proxy bundle to your Apigee organization and environment.

```shell
# Set your environment variables
export PROJECT_ID="your-gcp-project-id"
export APIGEE_ORG="${PROJECT_ID}"
export APIGEE_ENV="eval"

# Configure gcloud
gcloud config set project "${PROJECT_ID}"
gcloud auth login

# Deploy the bundle
apigeecli apis create bundle  \
   --proxy-zip ./out/apiproxies/weather.zip \
   --name  mcp-weather \
   --org "${APIGEE_ORG}" \
   --env "${APIGEE_ENV}" \
   --ovr \
   --default-token
```
Once deployed, the URL for your new MCP server will be `https://${APIGEE_HOSTNAME}/mcp/weather`.

### 3. Test the MCP Server

You can test your new MCP server in several ways.

#### Test with `curl`
A quick way to verify functionality is to make a `tools/call` request using `curl`.

```bash
# Set your Apigee hostname
export APIGEE_HOSTNAME="your-apigee-hostname"

curl "https://${APIGEE_HOSTNAME}/mcp/weather" \
  -H "accept: application/json" \
  -H "content-type: application/json" \
  -d '{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "alerts_query"
  },
  "id": 1
}'
```

#### Test with the MCP Inspector
The [MCP Inspector](https://github.com/modelcontextprotocol/inspector) is a web-based tool for interacting with MCP servers.

1.  Start the inspector using `npx`:
    ```bash
    npx @modelcontextprotocol/inspector
    ```
2.  Open the provided URL in your browser.
3.  Enter your MCP server URL (`https://${APIGEE_HOSTNAME}/mcp/weather`), connect to it, and explore the available tools.

#### Test with an AI Assistant
To interact with your new MCP server using natural language, you can configure an AI assistant to use it as a tool source. Assistants like the [Gemini CLI](https://github.com/google-gemini/gemini-cli) and the [Claude AI Desktop App](https://claude.ai/download) can connect to remote MCP servers.

* **Claude:** [Connecting to a remote MCP server](https://modelcontextprotocol.io/docs/develop/connect-remote-servers#connecting-to-a-remote-mcp-server)
* **Gemini CLI:** [How to set up your MCP server](https://google-gemini.github.io/gemini-cli/docs/tools/mcp-server.html#how-to-set-up-your-mcp-server)


## Template Features

The baseline [MCP template](https'://github.com/apigee/apigee-go-gen/blob/main/examples/templates/mcp/apiproxy.yaml) automatically
generates an Apigee API proxy that acts as a bridge, allowing Large Language Models (LLMs) to securely interact with
existing REST APIs. The generated proxy handles the translation between the MCP format and standard REST API conventions.

---

### Automated Tool Mapping

The template parses a source **OpenAPI Description** to expose API endpoints as **MCP tools**. Every defined operation
(e.g., `GET /users`, `POST /products`) is automatically made available for an LLM to discover and call. 🗣️

The generated `tools/list` response provides a comprehensive definition for each tool, which includes:

* **`inputSchema`**: Defines the required inputs for the tool, mapped from the API's request parameters.
* **`outputSchema`**: Defines the expected output structure, mapped directly from the response schema in the **OpenAPI Description**.

Additionally, the proxy supports emitting **`structuredContent`** in the `tools/call` response. When the backend REST API
returns a JSON payload (`application/json`), the proxy automatically includes it as structured data, allowing the LLM to
parse the output directly without needing to interpret raw text.

---

### Parameter Mapping

API request parameters are automatically mapped from the MCP tool's input schema to the backend REST request.
The proxy ensures data from the LLM is correctly placed in the corresponding location for the target API. This includes:

* **Query Parameters**
* **Header Values**
* **Path Variables**
* **Request Body Content**

This seamless mapping enables the LLM to provide data for the API without needing to conform to the underlying REST structure.

---

### Transcoding

A core function of the proxy is **transcoding** requests. All MCP tool calls arrive in a standardized **JSON-RPC** format.
The API proxy automatically unwraps this payload and transforms it into a conventional REST API request that the backend
service can understand. 🤖➡️🌐

This process includes:

* **Unwrapping the Payload**: Extracting the target operation, parameters, and body from the incoming JSON-RPC request.
* **Setting HTTP Headers**: Automatically setting necessary headers, such as `Content-Type` and `Accept`, based on the **OpenAPI Description**.
* **Constructing the HTTP Request**: Assembling the final `GET`, `POST`, `PUT`, etc., request with the correct URL, headers, and body.

---

### Request Body Formats

The template provides out-of-the-box support for backend APIs that consume either JSON or XML, handling the necessary
transformations automatically.

* **`application/json`**: For JSON-based backends, the request body is simply unwrapped from the MCP tool input and passed through.
* **`application/xml`**: For XML-based backends, the proxy performs a two-step process:
    1.  It unwraps the JSON data from the MCP request.
    2.  It transforms that JSON data into the correct XML format, using the schema defined in your **OpenAPI Description** to ensure validity.

---

### Security

If the **OpenAPI Description** defines a security requirement of type `oath2` or `openIdConnect`, the generated API proxy includes a discovery endpoint to support the OAuth flow. 🔐

This endpoint serves the **Protected Resource Metadata** as required by the MCP specification.

The metadata endpoint is exposed at the following path:

`/.well-known/oauth-protected-resource{mcp_server_basepath}`

!!! Note

    It's crucial to understand that **the generated proxy does not act as an authorization server**.
    Instead, this metadata simply informs the MCP client where to find the actual OAuth authorization server.

    Additionally, whether dynamic client registration is supported is determined by the capabilities of the authorization
    server itself, not the Apigee MCP API proxy.