/*
 *  Copyright 2025 Google LLC
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

var isApigee = (typeof context !== "undefined");
var log = isApigee?print:console.log;

function isString(obj) {
  return (Object.prototype.toString.call(obj) === '[object String]');
}

function getPrettyJSON(value) {
  return JSON.stringify(value, null, 2);
}


function _get(obj, keyString, defaultValue) {
  if (typeof obj !== 'object' || obj === null) {
    return defaultValue;
  }

  var keys = keyString.split('.');

  var current = obj;
  for (var i = 0; i < keys.length; i++) {
    var key = keys[i];
    if (typeof current !== 'object' || current === null || typeof current[key] === 'undefined') {
      return defaultValue;
    }
    current = current[key];
  }

  return current;
}

function setErrorResponse(ctx, status, error) {
  var mcpId = ctx.getVariable("mcp.id");

  var responseBody = {
    jsonrpc: "2.0",
    error: {
      code: 500,
      message: "Internal Server Error"
    }
  }

  if (mcpId) {
    responseBody.id = mcpId;
  }


  if (isString(error)) {
    responseBody.error.message = error
  }

  if (error.status) {
    responseBody.error.code = status;
    status = error.status
  }

  if (error.message) {
    responseBody.error.message = error.message;
  }

  // if (error.stack) {
  //   responseBody.stack = error.stack;
  // }

  var headers = [];
  if (error.headers) {
    headers = headers.concat(error.headers);
  }

  headers.push(['Content-Type', 'application/json']);
  ctx.setVariable("error_body", getPrettyJSON(responseBody))
  ctx.setVariable("error_status", status)
  ctx.setVariable("error_headers", getPrettyJSON(headers))

  throw new Error(responseBody.error.message)

  //setResponse(ctx, status, [], getPrettyJSON(responseBody));
}

function setResponse(ctx, status, headers, content) {
  ctx.setVariable("response.status.code", status.toString());

  if (Array.isArray(headers)) {
    //group headers by name (for multi-value headers)
    var headerMap = {}
    for (var i = 0; i < headers.length; i++) {
      var hName = headers[i][0];
      var hValue = headers[i][1];
      if (!headerMap[hName]) {
        headerMap[hName] = [];
      }
      headerMap[hName].push(hValue);
    }

    for (var header in headerMap) {
      var headerValues = headerMap[header];
      if (headerValues.length === 1) {
        ctx.setVariable("response.header." + header.toLowerCase(), headerMap[header][0]);
        continue;
      }

      ctx.setVariable("response.header." + header.toLowerCase() + "-Count", headerMap[header].length);
      for (var j = 0; j < headerValues.length; j++) {
        ctx.setVariable("response.header." + header.toLowerCase() + "-" + j, headerMap[header][j]);
      }
    }
  }
  ctx.setVariable("response.content", content)
}


function flattenAndSetFlowVariables(ctx, prefix, obj, path) {
  for (var key in obj) {
    if (Object.prototype.hasOwnProperty.call(obj, key)) {
      var newPath = path ? path + '.' + key : key;
      var value = obj[key];

      if (typeof value === 'object' && value !== null) {
        flattenAndSetFlowVariables(ctx, prefix, value, newPath);
      } else {
        if (context && typeof ctx.setVariable === 'function') {
          ctx.setVariable(prefix + newPath, value);
        }
      }
    }
  }
}

function combinePaths(path1, path2) {
  path1 = path1.trim();
  path2 = path2.trim();

  if (path1.charAt(path1.length - 1) === '/') {
    return path1.slice(0, -1) + path2;
  } else {
    return path1 + path2;
  }
}

function parseJsonRpc(ctx, jsonString, createFlowVars) {
  var rpc;

  try {
    rpc = JSON.parse(jsonString);
  } catch (e) {
    throw new Error("Error parsing JSON: " + e.message);
  }

  if (typeof rpc !== 'object' || rpc === null) {
    throw new Error("Parsed object is not a valid object.");
  }

  if (rpc.jsonrpc !== "2.0") {
    throw new Error("Invalid JSON-RPC version. Expected '2.0', but got: " + rpc.jsonrpc);
  }

  if (!(typeof rpc.method === 'string' || typeof rpc.error === 'object' || typeof rpc.result !== 'undefined')) {
    throw new Error("Parsed object does not conform to JSON-RPC 2.0 request or response structure.");
  }

  if (!createFlowVars) {
    return rpc;
  }

  flattenAndSetFlowVariables(ctx,"mcp.", rpc, '');

  return rpc
}

function modifyRequestPath(ctx) {

  var messagePath = ctx.getVariable("message.path");
  var mcpMethod = ctx.getVariable("mcp.method");
  var mcpToolName = ctx.getVariable("mcp.params.name");

  if (mcpMethod === "tools/call" && mcpToolName) {
    ctx.setVariable("message.path", combinePaths(messagePath, "/tools/" + mcpToolName))
  }
}



function replacePathParams(requestPath, argumentsObj, pathParamNames) {
  var hasPlaceholders = /\{([a-zA-Z0-9_]+)\}/.test(requestPath);

  if (!hasPlaceholders) {
    return requestPath;
  }

  // Ensure arguments object has the expected structure
  if (!argumentsObj || !argumentsObj.arguments) {
    throw new Error("Invalid arguments structure. 'arguments' is required when path contains placeholders.");
  }

  // Ensure pathParamNames is a valid array
  if (!Array.isArray(pathParamNames)) {
    throw new Error("Invalid pathParamNames. It must be an array of strings.");
  }

  var replacedPath = requestPath.replace(/\{([a-zA-Z0-9_]+)\}/g, function(match, paramName) {
    if (pathParamNames.indexOf(paramName) === -1) {
      throw new Error("Path parameter '" + paramName + "' is not a recognized parameter. Please check the provided list of pathParamNames.");
    }

    // Retrieve the value directly from the arguments object
    var paramValue = argumentsObj.arguments[paramName];

    if (typeof paramValue === 'undefined' || paramValue === null) {
      throw new Error("Missing required path parameter: '" + paramName + "'");
    }

    return paramValue;
  });

  return replacedPath;
}

function createQueryParams(argumentsObj, queryParamNames) {
  // Ensure arguments object has the expected structure
  if (!argumentsObj || !argumentsObj.arguments) {
    return "";
  }

  // Ensure queryParamNames is a valid array
  if (!Array.isArray(queryParamNames)) {
    console.error("Invalid queryParamNames. It must be an array of strings.");
    return "";
  }

  var params = [];

  // Iterate over the provided list of valid query parameter names
  for (var i = 0; i < queryParamNames.length; i++) {
    var key = queryParamNames[i];
    var value = argumentsObj.arguments[key];

    // Only include the key-value pair if the value is defined and not null
    if (typeof value !== 'undefined' && value !== null) {
      var encodedKey = encodeURIComponent(key);
      var encodedValue = encodeURIComponent(value);
      params.push(encodedKey + "=" + encodedValue);
    }
  }

  return params.length > 0 ? "?" + params.join("&") : "";
}


function createFullUrl(baseURL, requestPath, argumentsObj, pathParams, queryParams) {
  var fullPath = replacePathParams(requestPath, argumentsObj, pathParams);
  var queryString = createQueryParams(argumentsObj, queryParams);
  return baseURL + fullPath + queryString;
}


function parseStringToArray(str, defaultValue) {
  if (!str || typeof str !== 'string') {
    return defaultValue;
  }

  try {
    var parsedArray = JSON.parse(str);

    if (Array.isArray(parsedArray)) {
      return parsedArray;
    } else {
      return defaultValue;
    }
  } catch (error) {
    return defaultValue;
  }
}

function setToolCallTarget(ctx) {
  var rpc = parseJsonRpc(ctx, ctx.getVariable("request.content"), false)

  if (rpc.method !== "tools/call") {
    throw new Error("Cannot set target on non MCP tools/call method.")
  }

  var targetUrl = ctx.getVariable("propertyset.mcp-tools." + rpc["params"]["name"] + ".target_url");
  var targetPathSuffix = ctx.getVariable("propertyset.mcp-tools." + rpc["params"]["name"] + ".target_path_suffix");
  var targetVerb = ctx.getVariable("propertyset.mcp-tools." + rpc["params"]["name"] + ".target_verb");
  var targetContentType = ctx.getVariable("propertyset.mcp-tools." + rpc["params"]["name"] + ".target_content_type");
  var payloadParam = ctx.getVariable("propertyset.mcp-tools." + rpc["params"]["name"] + ".payload_param");
  var headerParams = parseStringToArray(ctx.getVariable("propertyset.mcp-tools." + rpc["params"]["name"] + ".header_params"), []);
  var queryParams = parseStringToArray(ctx.getVariable("propertyset.mcp-tools." + rpc["params"]["name"] + ".query_params"), []);
  var pathParams = parseStringToArray(ctx.getVariable("propertyset.mcp-tools." + rpc["params"]["name"] + ".path_params"), []);


  //Build the Request Object
  //Set Verb
  ctx.setVariable("message.verb", targetVerb);

  //Set URL
  ctx.setVariable("target.url", createFullUrl(targetUrl, targetPathSuffix, rpc["params"], pathParams, queryParams));

  //Set the Body
  if (targetVerb === "GET") {
    ctx.setVariable("message.content", "");
  } else {
    //post, put, delete, options
    if (targetContentType) {
      ctx.setVariable("request.header.Content-Type", targetContentType);
    }

    var requestBody = _get(rpc, "params.arguments." + payloadParam, null);
    if (requestBody) {
      if (isString(requestBody)) {
        ctx.setVariable("message.content", requestBody)
      } else {
        ctx.setVariable("message.content", getPrettyJSON(requestBody))
      }
    }
  }

  //Set Headers
  for (var headerName in headerParams) {
    var headerValue = _get(rpc, "params.arguments." + headerName, null);
    if (headerValue) {
      ctx.setVariable("request.header." + headerName, headerValue)
    }
  }
  
}

function processToolRes(ctx) {
  var statusCode = parseInt(ctx.getVariable("response.status.code"));
  var content = ctx.getVariable("response.content");

  var statusCodePrefix = parseInt(statusCode/100);

  var isError = false;
  if (statusCodePrefix === 4 || statusCodePrefix === 5) {
    isError = true;
  }

  var headers = [["Content-Type", "application/json"]];
  var mcpId = ctx.getVariable("mcp.id");

  var rpcResponse = {
    jsonrpc: "2.0",
    id: mcpId,
    result: {
      content: [
        {
          type: "text",
          text: content
        }
      ],
      isError: isError
    }
  }
  setResponse(ctx, 200, headers,  getPrettyJSON(rpcResponse));
}


if (!isApigee) {
  module.exports = {
    "flattenAndSetFlowVariables": flattenAndSetFlowVariables,
    "parseJsonRpc": parseJsonRpc,
    "setResponse": setResponse,
    "setErrorResponse": setErrorResponse,
    "createFullUrl": createFullUrl,
    "createQueryParams": createQueryParams,
    "replacePathParams": replacePathParams
  };
}