/*
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const {
  JsonRPCError,
  _get,
  combinePaths,
  convertJsonToXml,
  createFullUrl,
  createQueryParams,
  setErrorResponse,
  flattenAndSetFlowVariables,
  getPrettyJSON,
  getToolInfo,
  isPlainObject,
  isString,
  jsonToFormURLEncoded,
  parseJsonRpc,
  parseJsonString,
  processMCPRequest,
  processRESTRes,
  replacePathParams,
  setResponse,
  validateMcpToolsInfo,
  JSON_RPC_PARSE_ERROR,
  JSON_RPC_INVALID_REQUEST,
  JSON_RPC_METHOD_NOT_FOUND,
  JSON_RPC_INVALID_PARAMS,
  JSON_RPC_INTERNAL_ERROR
} = require("../resources/jsc/mcp.cjs");

const { expect, test, describe } = require('@jest/globals');


// Mocking the Apigee context object (ctx)
const mockContext = () => {
  const variables = {};
  return {
    getVariable: jest.fn(name => variables[name]),
    setVariable: jest.fn((name, value) => {
      variables[name] = value;
    }),
    removeVariable: jest.fn(name => {
      delete variables[name];
    }),
    // Helper to check all variables set for debugging
    getVariables: () => variables
  };
};

// Mocking global variable required by getToolInfo
global.mcpToolsInfo = {
  "fetch_data": {
    "target": {
      "url": "https://api.example.com/v1",
      "pathSuffix": "/resources/{resourceId}",
      "verb": "GET",
      "headers": {
        "accept": "application/json"
      }
    },
    "schemas": {},
    "inputParams": {
      "query": ["maxItems"],
      "path": ["resourceId"]
    }
  },
  "create_user": {
    "target": {
      "url": "https://api.example.com/v1",
      "pathSuffix": "/users",
      "verb": "POST",
      "headers": {
        "content-type": "application/json"
      }
    },
    "schemas": {
      "request": { "type": "object", "properties": { "name": { "type": "string" } } }
    },
    "inputParams": {
      "body": "userPayload",
      // inputParams.headers is defined as an array of argument names to promote to headers
      "headers": ["xRequestId", "authSecret"]
    }
  },
  "xml_update": {
    "target": {
      "url": "https://xml.example.com",
      "pathSuffix": "/update",
      "verb": "PUT",
      "headers": {
        "content-type": "application/xml"
      }
    },
    "schemas": {
      "request": {
        "type": "object",
        "xml": { "name": "UserRequest", "prefix": "ns", "namespace": "http://example.com/ns" },
        "properties": {
          "id": { "type": "string", "xml": { "attribute": false } },
          "name": { "type": "string", "xml": { "name": "UserName" } },
        }
      }
    },
    "inputParams": { "body": "userPayload" }
  }
};


// --- Core Utility Function Tests ---

describe('Type Checks (isString, isPlainObject)', () => {
  // NOTE: isString and isPlainObject are pure functions and don't rely on Apigee context.
  test('isString should correctly identify strings', () => {
    expect(isString('hello')).toBe(true);
    expect(isString(new String('hello'))).toBe(true);
    expect(isString(123)).toBe(false);
    expect(isString({})).toBe(false);
  });

  test('isPlainObject should correctly identify plain objects', () => {
    expect(isPlainObject({})).toBe(true);
    expect(isPlainObject({ a: 1 })).toBe(true);
    expect(isPlainObject([])).toBe(false);
    expect(isPlainObject(null)).toBe(false);
    expect(isPlainObject(new Date())).toBe(false);
  });
});

describe('JSON Utilities (getPrettyJSON, parseJsonString)', () => {
  test('getPrettyJSON should format JSON with 2-space indentation', () => {
    const obj = { a: 1, b: { c: 2 } };
    const expected = '{\n  "a": 1,\n  "b": {\n    "c": 2\n  }\n}';
    expect(getPrettyJSON(obj)).toBe(expected);
  });

  test('parseJsonString should successfully parse valid JSON', () => {
    const jsonStr = '{"key": 123}';
    expect(parseJsonString(jsonStr, null)).toEqual({ key: 123 });
  });

  test('parseJsonString should return defaultValue on invalid JSON', () => {
    const invalidStr = '{key: 123';
    expect(parseJsonString(invalidStr, 'default')).toBe('default');
    expect(parseJsonString(null, 'default')).toBe('default');
  });
});

describe('Object Path Getter (_get)', () => {
  const data = {
    user: {
      profile: {
        name: 'John',
        age: 30
      },
      settings: null
    },
    id: 100
  };

  test('should retrieve a top-level property', () => {
    expect(_get(data, 'id', 0)).toBe(100);
  });

  test('should retrieve a deeply nested property', () => {
    expect(_get(data, 'user.profile.name', 'N/A')).toBe('John');
  });

  test('should return defaultValue for non-existent path', () => {
    expect(_get(data, 'user.profile.email', 'none')).toBe('none');
  });

  test('should handle null intermediate paths gracefully', () => {
    expect(_get(data, 'user.settings.theme', 'light')).toBe('light');
  });

  test('should handle non-object input', () => {
    expect(_get(null, 'id', 0)).toBe(0);
    expect(_get(123, 'id', 0)).toBe(0);
  });
});

// --- Apigee Flow Variable and Response Utilities ---

describe('Apigee Flow Variable Utilities', () => {
  test('combinePaths should handle paths with and without trailing slash', () => {
    expect(combinePaths("/api/", "/resource")).toBe("/api/resource");
    expect(combinePaths("/api", "/resource")).toBe("/api/resource");
    expect(combinePaths("/api/", "resource")).toBe("/api/resource");
  });

  test('setResponse should correctly set status code and content', () => {
    const ctx = mockContext();
    setResponse(ctx, 201, [], "OK");
    expect(ctx.setVariable).toHaveBeenCalledWith("response.status.code", "201");
    expect(ctx.setVariable).toHaveBeenCalledWith("response.content", "OK");
  });

  test('setResponse should handle single-value headers', () => {
    const ctx = mockContext();
    const headers = [
      ["Content-Type", "application/json"],
      ["X-Custom", "value"]
    ];
    setResponse(ctx, 200, headers, "{}");
    expect(ctx.setVariable).toHaveBeenCalledWith("response.header.content-type", "application/json");
    expect(ctx.setVariable).toHaveBeenCalledWith("response.header.x-custom", "value");
  });

  test('setResponse should handle multi-value headers correctly', () => {
    const ctx = mockContext();
    const headers = [
      ["Set-Cookie", "session=123"],
      ["Set-Cookie", "user=abc"]
    ];
    setResponse(ctx, 200, headers, "{}");
    expect(ctx.setVariable).toHaveBeenCalledWith("response.header.set-cookie-count", 2);
    expect(ctx.setVariable).toHaveBeenCalledWith("response.header.set-cookie-0", "session=123");
    expect(ctx.setVariable).toHaveBeenCalledWith("response.header.set-cookie-1", "user=abc");
  });

  test('flattenAndSetFlowVariables should flatten nested objects with prefix', () => {
    const ctx = mockContext();
    const obj = {
      user: {
        id: 1,
        details: {
          age: 30
        }
      },
      status: 'active'
    };

    flattenAndSetFlowVariables(ctx, "mcp.", obj, '');

    expect(ctx.setVariable).toHaveBeenCalledWith("mcp.user.id", 1);
    expect(ctx.setVariable).toHaveBeenCalledWith("mcp.user.details.age", 30);
    expect(ctx.setVariable).toHaveBeenCalledWith("mcp.status", 'active');
  });

  test('setErrorResponse should set error variables and throw an error', () => {
    const ctx = mockContext();
    ctx.setVariable("mcp.id", 10001); // Using a distinct high integer ID for mcp.id

    const error = new JsonRPCError("Invalid Parameter Type", JSON_RPC_INVALID_PARAMS);
    const expectedMessage = error.message;

    expect(() => {
      setErrorResponse(ctx, 400, error);
    }).toThrow(expectedMessage);

    expect(ctx.setVariable).toHaveBeenCalledWith("error_status", 400);
    const errorBody = JSON.parse(ctx.getVariable("error_body"));
    expect(errorBody.id).toBe(10001); // Check for distinct integer ID
    expect(errorBody.error.code).toBe(JSON_RPC_INVALID_PARAMS);
    expect(errorBody.error.message).toBe("Invalid Parameter Type");
  });

});

// --- JSON-RPC Core Logic Tests ---

describe('JSON-RPC Parsing and Validation (parseJsonRpc)', () => {
  const ctx = mockContext();

  test('should successfully parse a valid request and set flow variables', () => {
    const jsonString = '{"jsonrpc": "2.0", "method": "test.method", "params": {"a": 1}, "id": 10002}';
    const rpc = parseJsonRpc(ctx, jsonString, true);

    expect(rpc.method).toBe("test.method");
    expect(ctx.getVariable("mcp.method")).toBe("test.method");
    expect(ctx.getVariable("mcp.params.a")).toBe(1);
  });

  test('should throw JsonRPCError on invalid JSON', () => {
    const invalidJson = '{invalid: json}';
    expect(() => parseJsonRpc(ctx, invalidJson, false)).toThrow(
      expect.objectContaining({ code: JSON_RPC_PARSE_ERROR })
    );
  });

  test('should throw JsonRPCError on invalid version', () => {
    const badVersion = '{"jsonrpc": "2.0", "version": "1.0", "method": "test"}'; // Changed 1.0 to 2.0 to pass initial version check on the missing property check later
    expect(() => parseJsonRpc(ctx, '{"jsonrpc": "1.0", "method": "test"}', false)).toThrow(
      expect.objectContaining({ code: JSON_RPC_INVALID_REQUEST })
    );
  });

  test('should throw JsonRPCError if missing required keys (method, result, or error)', () => {
    const missingKeys = '{"jsonrpc": "2.0", "id": 1}';
    expect(() => parseJsonRpc(ctx, missingKeys, false)).toThrow(
      expect.objectContaining({ code: JSON_RPC_INVALID_REQUEST })
    );
  });
});

// --- URL Construction Utilities ---

describe('URL Construction and Parameter Replacement', () => {
  const params = {
    arguments: {
      userId: 456,
      resourceId: 'abc',
      maxItems: 50,
      format: 'json',
      ignoredParam: 'test'
    }
  };
  const argumentsObj = params;

  test('replacePathParams should replace single path parameter', () => {
    const path = "/users/{userId}";
    const pathParams = ["userId"];
    expect(replacePathParams(path, argumentsObj, pathParams)).toBe("/users/456");
  });

  test('replacePathParams should replace multiple path parameters', () => {
    const path = "/users/{userId}/resources/{resourceId}";
    const pathParams = ["userId", "resourceId"];
    expect(replacePathParams(path, argumentsObj, pathParams)).toBe("/users/456/resources/abc");
  });

  test('replacePathParams should throw JsonRPCError if path parameter is missing', () => {
    const path = "/users/{nonexistent}";
    const pathParams = ["nonexistent"];
    expect(() => replacePathParams(path, argumentsObj, pathParams)).toThrow(
      expect.objectContaining({ code: JSON_RPC_INVALID_PARAMS })
    );
  });

  test('createQueryParams should create query string with specified parameters', () => {
    const queryParams = ["maxItems", "format"];
    expect(createQueryParams(argumentsObj, queryParams)).toBe("?maxItems=50&format=json");
  });

  test('createQueryParams should ignore parameters not in the list', () => {
    const queryParams = ["maxItems"];
    expect(createQueryParams(argumentsObj, queryParams)).toBe("?maxItems=50");
  });

  test('createFullUrl should combine base URL, path, and query params', () => {
    const baseURL = "http://api.base.com";
    const requestPath = "/path/{resourceId}";
    const pathParams = ["resourceId"];
    const queryParams = ["maxItems"];

    const fullUrl = createFullUrl(baseURL, requestPath, argumentsObj, pathParams, queryParams);
    expect(fullUrl).toBe("http://api.base.com/path/abc?maxItems=50");
  });
});

// --- Data Transcoding Utilities (XML) ---

describe('Data Transcoding (convertJsonToXml)', () => {
  // --- Test Case 1: Attributes, Namespaces, Custom Element Names, and Simple Types ---
  test('should exactly match XML with attributes, namespace, and custom names', () => {
    const jsonBody = {
      id: "U101",
      name: "Alice",
      age: 30
    };

    const schema = {
      type: 'object',
      xml: {
        name: 'UserRequest',
        prefix: 'ns',
        namespace: 'http://example.com/ns'
      },
      properties: {
        id: {
          type: 'string',
          xml: {
            attribute: true,
            name: 'UserID'
          }
        },
        name: {
          type: 'string',
          xml: {
            name: 'UserName'
          }
        },
        age: {
          type: 'number'
        }
      }
    };

    const xml = convertJsonToXml(jsonBody, schema);

    const expectedXml = `<?xml version="1.0" encoding="UTF-8"?>
<ns:UserRequest xmlns:ns="http://example.com/ns" UserID="U101">
  <UserName>Alice</UserName>
  <age>30</age>
</ns:UserRequest>`;

    expect(xml).toBe(expectedXml);
  });

  // ----------------------------------------------------------------------------------

  // --- Test Case 2: Wrapped Arrays ---
  test('should exactly match XML with a wrapped array structure', () => {
    const jsonBody = {
      orderId: 999,
      items: [{
        sku: "A-123",
        quantity: 1
      }, {
        sku: "B-456",
        quantity: 5
      }]
    };

    const schema = {
      type: 'object',
      xml: {
        name: 'Order',
      },
      properties: {
        orderId: {
          type: 'number'
        },
        items: {
          type: 'array',
          xml: {
            name: 'ItemsWrapper',
            wrapped: true
          },
          items: {
            type: 'object',
            xml: {
              name: 'Item'
            },
            properties: {
              sku: {
                type: 'string'
              },
              quantity: {
                type: 'number'
              }
            }
          }
        }
      }
    };

    const xml = convertJsonToXml(jsonBody, schema);

    // Note the 4-space indentation for the inner <Item> properties
    const expectedXml = `<?xml version="1.0" encoding="UTF-8"?>
<Order>
  <orderId>999</orderId>
  <ItemsWrapper>
    <Item>
      <sku>A-123</sku>
      <quantity>1</quantity>
    </Item>
    <Item>
      <sku>B-456</sku>
      <quantity>5</quantity>
    </Item>
  </ItemsWrapper>
</Order>`;

    expect(xml).toBe(expectedXml);
  });

  // --- Test Case 3: Unwrapped Arrays (Repeating Elements) ---
  test('should exactly match XML with an unwrapped array (repeating elements)', () => {
    const jsonBody = {
      locationId: "WH-East",
      products: [{
        productId: "Gizmo",
        inStock: true
      }, {
        productId: "Widget",
        inStock: false
      }]
    };

    const schema = {
      type: 'object',
      xml: {
        name: 'Inventory',
      },
      properties: {
        locationId: {
          type: 'string'
        },
        products: {
          type: 'array',
          items: {
            type: 'object',
            xml: {
              name: 'Product'
            },
            properties: {
              productId: {
                type: 'string'
              },
              inStock: {
                type: 'boolean'
              }
            }
          }
        }
      }
    };

    const xml = convertJsonToXml(jsonBody, schema);

    // Note the repeating <Product> elements, all indented 2 spaces from the root
    const expectedXml = `<?xml version="1.0" encoding="UTF-8"?>
<Inventory>
  <locationId>WH-East</locationId>
  <Product>
    <productId>Gizmo</productId>
    <inStock>true</inStock>
  </Product>
  <Product>
    <productId>Widget</productId>
    <inStock>false</inStock>
  </Product>
</Inventory>`;

    expect(xml).toBe(expectedXml);
  });

  // --- Test Case 4: XML Escaping ---
  test('should exactly match XML with proper character escaping', () => {
    const jsonBody = {
      id: "U<101>&",
      name: "Alice & Bob > C",
      age: 42
    };

    const schema = {
      type: 'object',
      xml: {
        name: 'UserRequest',
        prefix: 'ns',
        namespace: 'http://example.com/ns'
      },
      properties: {
        id: {
          type: 'string',
          xml: {
            attribute: true,
            name: 'UserID'
          }
        },
        name: {
          type: 'string',
          xml: {
            name: 'UserName'
          }
        },
        age: {
          type: 'number'
        }
      }
    };

    const xml = convertJsonToXml(jsonBody, schema);

    // Check for `&lt;`, `&gt;`, and `&amp;` in both attribute and element content
    const expectedXml = `<?xml version="1.0" encoding="UTF-8"?>
<ns:UserRequest xmlns:ns="http://example.com/ns" UserID="U&lt;101&gt;&amp;">
  <UserName>Alice &amp; Bob &gt; C</UserName>
  <age>42</age>
</ns:UserRequest>`;

    expect(xml).toBe(expectedXml);
  });


});

describe('Data Transcoding (jsonToFormURLEncoded)', () => {

  test('jsonToFormURLEncoded should convert flat JSON', () => {
    const data = { a: 1, b: 'value' };
    expect(jsonToFormURLEncoded(data)).toBe('a=1&b=value');
  });

  test('jsonToFormURLEncoded should handle arrays', () => {
    const data = { colors: ['red', 'blue'] };
    // Arrays flatten to multiple key/value pairs
    expect(jsonToFormURLEncoded(data)).toBe('colors=red&colors=blue');
  });

  test('jsonToFormURLEncoded should flatten nested objects using dot notation', () => {
    const data = { user: { name: 'Bob', age: 40 } };
    expect(jsonToFormURLEncoded(data)).toBe('user.name=Bob&user.age=40');
  });
});

// --- MCP Request/Response Processing (Integration Style) ---

describe('MCP Request/Response Processing', () => {

  test('getToolInfo should retrieve a valid tool definition', () => {
    const ctx = mockContext();
    const info = getToolInfo(ctx, "fetch_data");
    expect(info.target.verb).toBe("GET");
  });

  test('getToolInfo should throw JsonRPCError if tool is not found', () => {
    const ctx = mockContext();
    expect(() => getToolInfo(ctx, "nonexistent_tool")).toThrow(
      expect.objectContaining({ code: JSON_RPC_METHOD_NOT_FOUND })
    );
  });

  test('processMCPRequest should handle GET requests', () => {
    const ctx = mockContext();
    const rpcRequest = {
      jsonrpc: "2.0",
      method: "tools/call",
      params: {
        name: "fetch_data",
        arguments: {
          resourceId: "R123",
          maxItems: 10
        }
      },
      id: 10003 // Distinct integer ID
    };
    ctx.setVariable("request.content", JSON.stringify(rpcRequest));

    processMCPRequest(ctx);

    expect(ctx.getVariable("message.verb")).toBe("GET");
    expect(ctx.getVariable("target.url")).toBe("https://api.example.com/v1/resources/R123?maxItems=10");
    expect(ctx.getVariable("request.header.Accept")).toBe("application/json");
    expect(ctx.removeVariable).toHaveBeenCalledWith("message.content");
    expect(ctx.getVariable("mcp_tool.target.verb")).toBe("GET");
  });

  // REVISED TEST: Checks the promotion of arguments to headers using the array logic
  test('processMCPRequest should handle POST requests promoting arguments to headers', () => {
    const ctx = mockContext();
    const rpcRequest = {
      jsonrpc: "2.0",
      method: "tools/call",
      params: {
        name: "create_user",
        arguments: {
          userPayload: { name: "Jane" },
          xRequestId: "req-12345", // Promoted
          authSecret: "secret-token", // Promoted
          // This argument is NOT defined in inputParams.headers, so it should be ignored
          ignoredArg: "ignore-me"
        }
      },
      id: 10004 // Distinct integer ID
    };
    ctx.setVariable("request.content", JSON.stringify(rpcRequest));

    processMCPRequest(ctx);

    expect(ctx.getVariable("message.verb")).toBe("POST");
    expect(ctx.getVariable("request.header.Content-Type")).toBe("application/json");

    // 1. Argument `xRequestId` is promoted to header `xRequestId`
    expect(ctx.getVariable("request.header.xRequestId")).toBe("req-12345");

    // 2. Argument `authSecret` is promoted to header `authSecret`
    expect(ctx.getVariable("request.header.authSecret")).toBe("secret-token");

    // 3. Ignored argument should NOT be set as a header
    expect(ctx.getVariable("request.header.ignoredArg")).toBeUndefined();

    // Check body content
    expect(JSON.parse(ctx.getVariable("message.content"))).toEqual({ name: "Jane" });
  });

  test('processRESTRes should wrap successful REST response (plain object) directly', () => {
    const ctx = mockContext();
    // Using integer for response.status.code
    ctx.setVariable("response.status.code", 200);
    ctx.setVariable("response.content", '{"name": "result"}');
    ctx.setVariable("mcp.id", 10005); // Distinct integer ID

    processRESTRes(ctx);

    expect(ctx.setVariable).toHaveBeenCalledWith("response.status.code", "200");
    const rpcResponse = JSON.parse(ctx.getVariable("response.content"));

    expect(rpcResponse.id).toBe(10005); // Check for distinct integer ID
    expect(rpcResponse.result.isError).toBe(false);
    // Should be the plain object directly
    expect(rpcResponse.result.structuredContent).toEqual({ name: "result" });
  });

  test('processRESTRes should wrap non-plain object responses (array/literal) in a "result" key', () => {
    const ctx = mockContext();
    const arrayContent = '[1, "two", {"key": 3}]';
    // Using integer for response.status.code
    ctx.setVariable("response.status.code", 200);
    ctx.setVariable("response.content", arrayContent);
    ctx.setVariable("mcp.id", 10006); // Distinct integer ID

    processRESTRes(ctx);

    const rpcResponse = JSON.parse(ctx.getVariable("response.content"));

    expect(rpcResponse.id).toBe(10006); // Check for distinct integer ID
    expect(rpcResponse.result.isError).toBe(false);
    // Array response should be wrapped under the "result" key
    expect(rpcResponse.result.structuredContent).toEqual({ result: [1, "two", { key: 3 }] });
    expect(rpcResponse.result.content[0].text).toBe(arrayContent);
  });

  test('processRESTRes should mark 4xx/5xx responses as isError: true', () => {
    const ctx = mockContext();
    // Using integer for response.status.code
    ctx.setVariable("response.status.code", 404);
    ctx.setVariable("response.content", '{"error": "Not Found"}');
    ctx.setVariable("mcp.id", 10007); // Distinct integer ID

    processRESTRes(ctx);

    const rpcResponse = JSON.parse(ctx.getVariable("response.content"));
    expect(rpcResponse.id).toBe(10007); // Check for distinct integer ID
    expect(rpcResponse.result.isError).toBe(true);
  });
});

describe('MCP Tool Info Validation (validateMcpToolsInfo) -  Tests', () => {

  // --- POSITIVE TEST CASES ---

  test('P-1: Should successfully validate a minimal tool structure', () => {
    const validToolList = {
      "minimal_tool": {
        "target": { "url": "http://min.com", "pathSuffix": "/", "verb": "GET" },
        "inputParams": { "body": "", "path": [], "query": [], "headers": [] }
      }
    };
    expect(() => validateMcpToolsInfo(validToolList)).not.toThrow();
  });

  test('P-2: Should successfully validate a full tool structure with all optional fields', () => {
    const validToolList = {
      "full_tool": {
        "target": {
          "url": "http://full.com",
          "pathSuffix": "/data",
          "verb": "POST",
          "headers": { "Accept": "application/json", "Api-Key": "str-123" } // Valid object/string map
        },
        "schemas": { "request": { "type": "object" } }, // Valid nested object
        "inputParams": {
          "body": "payload",
          "path": ["id", "version"],
          "query": ["limit"],
          "headers": ["auth"]
        } // Valid string arrays
      }
    };
    expect(() => validateMcpToolsInfo(validToolList)).not.toThrow();
  });

  test('P-3: Should successfully validate a tool with only target ', () => {
    const validToolList = {
      "good_tool": {
        "target": { "url": "http://a.com", "pathSuffix": "/", "verb": "GET" }
      }
    };

    expect(() => validateMcpToolsInfo(validToolList)).not.toThrow();

  });

  // --- NEGATIVE TEST CASES (Top Level/Required Fields) ---

  test('N-1: Should throw for missing required top-level key: target', () => {
    const invalidToolList = {
      "bad_tool": {
        "inputParams": { "body": "", "path": [], "query": [], "headers": [] }
      }
    };
    expect(() => validateMcpToolsInfo(invalidToolList)).toThrow(
      expect.objectContaining({
        message: expect.stringContaining("Missing required top-level key: target.")
      })
    );
  });


  // --- NEGATIVE TEST CASES (Target Sub-fields) ---

  test('N-2: Should throw for missing required string property: target.url', () => {
    const invalidToolList = {
      "bad_tool": {
        "target": { "pathSuffix": "/", "verb": "GET" }, // URL is missing
        "inputParams": { "body": "", "path": [], "query": [], "headers": [] }
      }
    };
    expect(() => validateMcpToolsInfo(invalidToolList)).toThrow(
      expect.objectContaining({
        message: expect.stringContaining("target is missing required string property: url.")
      })
    );
  });

  test('N-3: Should throw for target.headers provided but not an object', () => {
    const invalidToolList = {
      "bad_tool": {
        "target": { "url": "http://a.com", "pathSuffix": "/", "verb": "GET", "headers": "not_object" },
        "inputParams": { "body": "", "path": [], "query": [], "headers": [] }
      }
    };
    expect(() => validateMcpToolsInfo(invalidToolList)).toThrow(
      expect.objectContaining({
        message: expect.stringContaining("target.headers must be an object if provided.")
      })
    );
  });

  test('N-4: Should throw if any value in target.headers is not a string', () => {
    const invalidToolList = {
      "bad_tool": {
        "target": {
          "url": "http://a.com",
          "pathSuffix": "/",
          "verb": "POST",
          "headers": { "X-Limit": "10", "X-Timeout": 1000 } // 1000 is not a string
        },
        "inputParams": { "body": "payload", "path": [], "query": [], "headers": [] }
      }
    };
    expect(() => validateMcpToolsInfo(invalidToolList)).toThrow(
      expect.objectContaining({
        message: expect.stringContaining("All values in target.headers must be strings (Header key: X-Timeout).")
      })
    );
  });

  // --- NEGATIVE TEST CASES (Schemas) ---

  test('N-5: Should throw if schemas is provided but not an object', () => {
    const invalidToolList = {
      "bad_tool": {
        "target": { "url": "http://a.com", "pathSuffix": "/", "verb": "GET" },
        "schemas": "not_object", // Invalid
        "inputParams": { "body": "", "path": [], "query": [], "headers": [] }
      }
    };
    expect(() => validateMcpToolsInfo(invalidToolList)).toThrow(
      expect.objectContaining({
        message: expect.stringContaining("schemas must be an object if provided.")
      })
    );
  });

  test('N-6: Should throw if schemas.request is provided but not an object', () => {
    const invalidToolList = {
      "bad_tool": {
        "target": { "url": "http://a.com", "pathSuffix": "/", "verb": "GET" },
        "schemas": { "request": 123 }, // Invalid
        "inputParams": { "body": "", "path": [], "query": [], "headers": [] }
      }
    };
    expect(() => validateMcpToolsInfo(invalidToolList)).toThrow(
      expect.objectContaining({
        message: expect.stringContaining("schemas.request must be an object if provided.")
      })
    );
  });

  // --- NEGATIVE TEST CASES (InputParams Sub-fields) ---

  test('N-7: Should throw if inputParams.body is present but not a string', () => {
    const invalidToolList = {
      "bad_tool": {
        "target": { "url": "http://a.com", "pathSuffix": "/", "verb": "GET" },
        "inputParams": { "body": null, "path": [], "query": [], "headers": [] } // Invalid
      }
    };
    expect(() => validateMcpToolsInfo(invalidToolList)).toThrow(
      expect.objectContaining({
        message: expect.stringContaining("inputParams.body must be a string")
      })
    );
  });

  test('N-8: Should throw if inputParams.path is present but not an array', () => {
    const invalidToolList = {
      "bad_tool": {
        "target": { "url": "http://a.com", "pathSuffix": "/", "verb": "GET" },
        "inputParams": { "body": "", "path": "not_array", "query": [], "headers": [] } // Invalid
      }
    };
    expect(() => validateMcpToolsInfo(invalidToolList)).toThrow(
      expect.objectContaining({
        message: expect.stringContaining("nputParams.path must be an array")
      })
    );
  });

  test('N-9: Should throw if inputParams.query contains non-string elements', () => {
    const invalidToolList = {
      "bad_tool": {
        "target": { "url": "http://a.com", "pathSuffix": "/", "verb": "GET" },
        "inputParams": { "body": "", "path": [], "query": ["a", 123, "b"], "headers": [] } // 123 is not a string
      }
    };
    expect(() => validateMcpToolsInfo(invalidToolList)).toThrow(
      expect.objectContaining({
        message: expect.stringContaining("All elements in inputParams.query array must be strings.")
      })
    );
  });
});

