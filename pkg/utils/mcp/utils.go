//  Copyright 2025 Google LLC
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package mcp

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	"regexp"
	"strings"
)

// createMapEntry appends a key and value node pair to the Content slice of a yaml.MappingNode.
// It returns the created value node.
func createMapEntry(parent *yaml.Node, key string, value *yaml.Node) *yaml.Node {
	parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, value)
	return value
}

// createSchemaEntry creates a new YAML map node structure for a JSON Schema.
// It initializes the node with "type" and an empty "properties" map.
func createSchemaEntry(schemaType string) *yaml.Node {
	schemaNode := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{}}
	createMapEntry(schemaNode, "type", &yaml.Node{Kind: yaml.ScalarNode, Value: schemaType})
	schemaProperties := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{}}
	createMapEntry(schemaNode, "properties", schemaProperties)

	return schemaNode
}

// addPropertyToSchema adds a new property (key and schema node) to the "properties" field
// within the parent schema node.
func addPropertyToSchema(parent *yaml.Node, key string, schema *yaml.Node) *yaml.Node {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		propName := parent.Content[i].Value
		propNode := parent.Content[i+1]
		if propName != "properties" {
			continue
		}

		createMapEntry(propNode, key, schema)
	}

	return nil
}

// GetChildNodeByJSONPath finds a single YAML node within the root node using a JSONPath expression.
// If no node is found, it returns (nil, nil). If more than one node is found, it returns an error.
func GetChildNodeByJSONPath(root *yaml.Node, jsonPath string) (*yaml.Node, error) {
	return GetChildNodeByJSONPathOrDefault(root, jsonPath, nil)
}

// GetChildNodeByJSONPathOrDefault finds a single YAML node within the root node using a JSONPath expression.
// If no node is found, it returns the provided default node. If more than one node is found, it returns an error.
func GetChildNodeByJSONPathOrDefault(root *yaml.Node, jsonPath string, def *yaml.Node) (*yaml.Node, error) {

	var nodes, err = GetChildNodesByJSONPath(root, jsonPath)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return def, nil
	}

	if len(nodes) > 1 {
		return nil, errors.Errorf("more than one node found at JSON path '%s'", jsonPath)
	}

	return nodes[0], nil
}

// GetChildNodesByJSONPath finds all YAML nodes within the root node that match the given JSONPath expression.
func GetChildNodesByJSONPath(root *yaml.Node, jsonPath string) ([]*yaml.Node, error) {
	var err error
	var yamlPath *yamlpath.Path

	yamlPath, err = yamlpath.NewPath(jsonPath)
	if err != nil {
		return nil, errors.New(err)
	}

	var yamlNodes []*yaml.Node
	yamlNodes, err = yamlPath.Find(root)
	if err != nil {
		return nil, errors.New(err)
	}

	return yamlNodes, nil

}

// traverseAndDeleteXML recursively traverses a yaml.Node tree and removes any map key named "xml"
// unless its immediate parent is the value associated with a key named "properties".
//
// The 'parentKey' argument carries the key of the map that contains the current 'node'.
func traverseAndDeleteXML(node *yaml.Node, parentKey string) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		// Post-order traversal: First, clean up children recursively.
		// node.Content is [key1, value1, key2, value2, ...]
		for i := 1; i < len(node.Content); i += 2 {
			keyNode := node.Content[i-1]
			valueNode := node.Content[i]
			// The new parentKey for the recursive call is the current key's value (e.g., "properties")
			traverseAndDeleteXML(valueNode, keyNode.Value)
		}

		// Second, filter the current map's content.
		var newContent []*yaml.Node
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			isXMLKey := keyNode.Kind == yaml.ScalarNode && keyNode.Value == "xml"

			if isXMLKey {
				// Exclusion rule check: Do NOT delete if the immediate parent map's key was "properties".
				if parentKey == "properties" {
					// Keep the pair
					newContent = append(newContent, keyNode, valueNode)
				} else {
					// Delete: Skip this key/value pair
					continue
				}
			} else {
				// Keep all other pairs
				newContent = append(newContent, keyNode, valueNode)
			}
		}
		node.Content = newContent

	case yaml.SequenceNode:
		// Traverse sequence children (elements of an array). The parentKey remains the same.
		for _, child := range node.Content {
			traverseAndDeleteXML(child, parentKey)
		}

	case yaml.ScalarNode:
		fallthrough
	default:
		// Base case: Do nothing to scalar nodes.
		return
	}
}

// traverseAndRewriteExample recursively traverses a yaml.Node tree and rewrites the
// non-standard 'example: <value>' field into the standard 'examples: [<value>]' array format.
// This is done through post-order traversal to ensure child nodes are processed first.
func traverseAndRewriteExample(node *yaml.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		// Post-order traversal: 1. Clean up children first
		// node.Content is [key1, value1, key2, value2, ...]
		for i := 1; i < len(node.Content); i += 2 {
			traverseAndRewriteExample(node.Content[i])
		}

		// 2. Filter and rewrite the current map's content
		var newContent []*yaml.Node
		var exampleValue *yaml.Node

		// First pass: Find "example" and collect all other fields
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Kind == yaml.ScalarNode && keyNode.Value == "example" {
				// Found "example". Store its value and skip adding this pair to newContent.
				exampleValue = valueNode
				continue
			}

			// Keep all other pairs
			newContent = append(newContent, keyNode, valueNode)
		}

		// 3. Rewrite if "example" was found
		if exampleValue != nil {
			// Create the new "examples" key node
			examplesKey := &yaml.Node{Kind: yaml.ScalarNode, Value: "examples"}

			// Create the new value: a sequence node containing the single example value
			examplesValue := &yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{exampleValue}}

			// Append the new "examples: [value]" pair to the content
			newContent = append(newContent, examplesKey, examplesValue)
		}

		node.Content = newContent

	case yaml.SequenceNode:
		// Traverse sequence children
		for _, child := range node.Content {
			traverseAndRewriteExample(child)
		}

	case yaml.ScalarNode:
		fallthrough
	default:
		// Base case: Do nothing
		return
	}
}

// generateAlternateOperationID creates a predictable and unique operationId
// from the API Path and HTTP Verb for an operation missing an explicit ID.
//
// The format is: <http_method>-<sanitized_api_path>
//
// Example Input: apiPath="/v1/users/{id}", httpVerb="PATCH"
// Example Output: "patch-v1_users_id"
func generateAlternateOperationID(apiPath string, httpVerb string) string {

	// 1. Convert HTTP Verb to lowercase (e.g., "PATCH" -> "patch")
	httpMethod := strings.ToLower(httpVerb)

	// 2. Sanitize the API Path (e.g., "/v1/users/{id}" -> "v1_users_id")

	// Remove leading/trailing slashes (e.g., "/v1/users" -> "v1/users")
	sanitizedPath := strings.Trim(apiPath, "/")

	// Regex to replace path parameters {param} with just the parameter name 'param'.
	// This ensures the parameter name is retained but the brackets are removed.
	reParam := regexp.MustCompile(`{[a-zA-Z0-9_-]+}`)
	sanitizedPath = reParam.ReplaceAllStringFunc(sanitizedPath, func(match string) string {
		// Remove the surrounding braces: "{param}" -> "param"
		return strings.Trim(match, "{}")
	})

	// **New step to ensure the path segment is entirely lowercase**
	sanitizedPath = strings.ToLower(sanitizedPath)

	// Replace all remaining non-alphanumeric characters (including '/') with underscores.
	// This converts path separators and hyphens to underscores.
	reNonAlphanum := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sanitizedPath = reNonAlphanum.ReplaceAllString(sanitizedPath, "_")

	// Remove any redundant leading/trailing underscores created during sanitization
	sanitizedPath = strings.Trim(sanitizedPath, "_")

	// 3. Combine the method and the sanitized path

	// If the path was simply "/" (which results in an empty sanitizedPath),
	// return only the method name.
	if sanitizedPath == "" {
		return httpMethod
	}

	return fmt.Sprintf("%s-%s", httpMethod, sanitizedPath)
}

// extractParamsFromPath extracts parameter names (the part inside curly braces) from an API path string.
// Example: "/v1/users/{id}/details/{field}" returns ["id", "field"].
func extractParamsFromPath(path string) []string {
	var pathParams []string
	var paramsRegex = regexp.MustCompile(`(?m)\{([^}]+)}`)
	paramsMatches := paramsRegex.FindAllStringSubmatch(path, -1)
	for _, match := range paramsMatches {
		pathParams = append(pathParams, match[1])
	}
	return pathParams
}

// validatePathParams checks that every parameter defined in the path string (e.g., {id})
// has a corresponding parameter definition node either at the operation level or the path level.
func validatePathParams(operationJSONPath string, pathParams []string, inlinedOperationParamsNode *yaml.Node, inlinedPathParamsNode *yaml.Node) error {
	var err error
	for _, pathParam := range pathParams {
		var operationParamNode *yaml.Node
		if inlinedOperationParamsNode != nil {
			if operationParamNode, err = GetChildNodeByJSONPath(inlinedOperationParamsNode, fmt.Sprintf("$[?(@.name == '%s')]", pathParam)); err != nil {
				return err
			}
		}

		var pathParamNode *yaml.Node
		if inlinedPathParamsNode != nil {
			if pathParamNode, err = GetChildNodeByJSONPath(inlinedPathParamsNode, fmt.Sprintf("$[?(@.name == '%s')]", pathParam)); err != nil {
				return err
			}
		}

		if pathParamNode == nil && operationParamNode == nil {
			return errors.Errorf("Operation at %s is missing the '%s' path parameter definition", operationJSONPath, pathParam)
		}
	}
	return nil
}

// addMissingTypeFieldToOutputSchema ensures the outputSchema has a 'type: object' field.
// If the schema is missing 'type', it is added. If the schema has a type other than 'object',
// it is wrapped in a new schema of 'type: object' under a 'result' property.
func addMissingTypeFieldToOutputSchema(outputSchema *yaml.Node) (*yaml.Node, error) {
	var err error
	if outputSchema == nil {
		return nil, nil
	}

	var outputSchemaType *yaml.Node
	if outputSchemaType, err = GetChildNodeByJSONPath(outputSchema, "$.type"); err != nil {
		return nil, err
	}

	//INFO: MCP version 2025-06-18 only supports outputSchema of "type": "object"
	if outputSchemaType != nil && outputSchemaType.Value != "object" {
		wrappedSchema := createSchemaEntry("object")
		addPropertyToSchema(wrappedSchema, "result", outputSchema)
		outputSchema = wrappedSchema
	} else if outputSchemaType == nil {
		createMapEntry(outputSchema, "type", &yaml.Node{Kind: yaml.ScalarNode, Value: "object"})
	}

	return outputSchema, nil
}

// addDollarDefsToSchema finds all referenced schemas within the provided schema node
// and moves them into a top-level '$defs' map within the schema, as required by
// newer JSON Schema drafts. It also rewrites the original $ref pointers.
func addDollarDefsToSchema(schema *yaml.Node, oas3Node *yaml.Node) error {
	var err error
	if schema == nil {
		return nil
	}

	var references *yaml.Node
	if references, err = FindYAMLReferences(schema, oas3Node); err != nil {
		return err
	}

	schema = DeepCloneYAML(schema)
	rewriteRefs(schema)

	if references != nil && len(references.Content) > 0 {
		createMapEntry(schema, "$defs", references)
	}

	return nil
}

// processHeaderAndQueryParams extracts header, path, and query parameters from combined parameter nodes,
// adds their schemas as properties to the inputSchema, and returns lists of the parameter names.
func processHeaderAndQueryParams(operationId string, operationJSONPath string, operationParamsNode *yaml.Node, pathParamsNode *yaml.Node, inlinedOperationParamsNode *yaml.Node, inlinedPathParamsNode *yaml.Node, oas3Node *yaml.Node, inputSchema *yaml.Node) (headerParamsList []string, pathParamsList []string, queryParamsList []string, err error) {
	if operationParamsNode != nil || pathParamsNode != nil {
		var combinedParamNodes []*yaml.Node
		if combinedParamNodes, err = combineParams(operationParamsNode, pathParamsNode, inlinedOperationParamsNode, inlinedPathParamsNode, oas3Node); err != nil {
			return nil, nil, nil, err
		}

		for k, paramNode := range combinedParamNodes {
			var paramNameNode *yaml.Node
			var paramInNode *yaml.Node
			var paramSchemaNode *yaml.Node

			if isRefValueMap(paramNode) {
				if paramNode, err = InlineYAMLReferences(paramNode, oas3Node); err != nil {
					return nil, nil, nil, err
				}
			}

			if paramNameNode, err = GetChildNodeByJSONPath(paramNode, "$.name"); err != nil {
				return nil, nil, nil, err
			}

			if paramInNode, err = GetChildNodeByJSONPath(paramNode, "$.in"); err != nil {
				return nil, nil, nil, err
			}

			if paramSchemaNode, err = GetChildNodeByJSONPath(paramNode, "$.schema"); err != nil {
				return nil, nil, nil, err
			}

			if paramNameNode == nil {
				return nil, nil, nil, errors.Errorf("Parameter #%d within the '%s' operation is missing the 'in' property", k, operationId)
			}

			if paramInNode == nil {
				return nil, nil, nil, errors.Errorf("Parameter '%s' within the '%s' operation is missing the 'in' property", paramNameNode.Value, operationJSONPath)
			}

			if paramSchemaNode == nil {
				return nil, nil, nil, errors.Errorf("Parameter '%s' within the '%s' operation is missing the 'schema' property", paramNameNode.Value, operationJSONPath)
			}

			switch paramInNode.Value {
			case "header":
				addPropertyToSchema(inputSchema, paramNameNode.Value, paramSchemaNode)
				headerParamsList = append(headerParamsList, paramNameNode.Value)
				break
			case "path":
				addPropertyToSchema(inputSchema, paramNameNode.Value, paramSchemaNode)
				pathParamsList = append(pathParamsList, paramNameNode.Value)
				break
			case "query":
				addPropertyToSchema(inputSchema, paramNameNode.Value, paramSchemaNode)
				queryParamsList = append(queryParamsList, paramNameNode.Value)
				break
			}
		}
	}
	return headerParamsList, pathParamsList, queryParamsList, nil
}

// processResponseBody parses the OpenAPI responses node to determine the content type and schema
// of the response body. It prioritizes the first 2XX JSON response. If multiple 2XX JSON responses
// exist, it wraps their schemas in a 'oneOf' structure.
func processResponseBody(responsesNode *yaml.Node, operationId string) (responseContentType string, responseContentSchemaNode *yaml.Node, outputSchema *yaml.Node, err error) {
	if responsesNode != nil {
		outputSchemasOneOf := &yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{}}

		var firstResponseMimeType string
		for r := 0; r+1 < len(responsesNode.Content); r += 2 {
			var responseContentNode *yaml.Node
			responseStatusCode := responsesNode.Content[r].Value
			if responseContentNode, err = GetChildNodeByJSONPath(responsesNode.Content[r+1], "$.content"); err != nil {
				return "", nil, nil, errors.Errorf("Could not get 'responses.%s.content' element from the '%s' operation. Error: %s", responseStatusCode, operationId, err.Error())
			}

			if responseContentNode == nil {
				continue
			}

			for m := 0; m+1 < len(responseContentNode.Content); m += 2 {
				responseMimeType := responseContentNode.Content[m].Value

				if firstResponseMimeType == "" && strings.Index(responseStatusCode, "2") == 0 {
					firstResponseMimeType = responseMimeType
				}

				if strings.Contains(responseMimeType, "json") && strings.Index(responseStatusCode, "2") == 0 {
					responseContentType = responseMimeType
				}

				if !strings.Contains(responseMimeType, "json") {
					continue
				}

				var responseSchema *yaml.Node
				if responseSchema, err = GetChildNodeByJSONPath(responseContentNode.Content[m+1], "$.schema"); err != nil {
					return "", nil, nil, err
				}
				outputSchemasOneOf.Content = append(outputSchemasOneOf.Content, responseSchema)
			}
		}

		//if we could not find a 2XX, JSON response type, then use the first 2XX response type
		// (this is needed so that we give preference to JSON responses)
		if responseContentType == "" {
			responseContentType = firstResponseMimeType
		}

		if len(outputSchemasOneOf.Content) == 1 {
			outputSchema = outputSchemasOneOf.Content[0]
		} else if len(outputSchemasOneOf.Content) > 1 {

			outputSchema = &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{}}
			createMapEntry(outputSchema, "oneOf", outputSchemasOneOf)
			//INFO: MCP version 2025-06-18 only supports outputSchema of "type": "object"
			createMapEntry(outputSchema, "type", &yaml.Node{Kind: yaml.ScalarNode, Value: "object"})
		}
	}

	responseContentSchemaNode = outputSchema
	return responseContentType, outputSchema, responseContentSchemaNode, nil
}

// processRequestBody parses the OpenAPI request body node to determine the content type,
// schema, and suggested parameter name for the request body. It prioritizes the JSON content type.
// It also adds the request body schema as a property to the main input schema.
func processRequestBody(operationId string, operationJSONPath string, requestBodyNode *yaml.Node, oas3Node *yaml.Node) (requestContentType string, requestBodyParam string, requestContentSchemaNode *yaml.Node, inputSchema *yaml.Node, err error) {
	inputSchema = createSchemaEntry("object")

	requestContentSchemaNode = &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{}}

	if requestBodyNode != nil {
		var requestBodyContent *yaml.Node
		if requestBodyContent, err = GetChildNodeByJSONPath(requestBodyNode, "$.content"); err != nil {
			return "", "", nil, nil, err
		}

		if requestBodyContent == nil || len(requestBodyContent.Content) < 2 {
			return "", "", nil, nil, errors.Errorf("The 'requestBody' property witin the '%s' operation must have at least one content element", operationId)
		}

		if len(requestBodyContent.Content) >= 2 {
			//find the JSON content element, or get the first one
			for c := 0; c+1 < len(requestBodyContent.Content); c += 1 {
				var curSchemaNode *yaml.Node

				curContentType := requestBodyContent.Content[c].Value
				if curSchemaNode, err = GetChildNodeByJSONPath(requestBodyContent.Content[c+1], "$.schema"); err != nil {
					return "", "", nil, nil, errors.Errorf("The 'requestBody.%s.schema' is missing within the '%s' operation", curContentType, operationId)
				}

				//check the original spec to see if it was a $ref

				if c == 0 {
					requestContentType = curContentType
					requestContentSchemaNode = curSchemaNode
				}

				if strings.Contains(curContentType, "json") {
					requestContentType = curContentType
					requestContentSchemaNode = curSchemaNode
					break
				}
			}

			requestBodyParam = fmt.Sprintf("%sBody", operationId)
			requestBodyJSONPath := fmt.Sprintf("%s.requestBody.content.%s.schema.$ref", operationJSONPath, requestContentType)
			var dollarRefNode *yaml.Node
			if dollarRefNode, err = GetChildNodeByJSONPath(oas3Node, requestBodyJSONPath); err != nil {
				return "", "", nil, nil, err
			}

			if dollarRefNode != nil {
				var re = regexp.MustCompile(`(?msi)#/components/schemas/(.+)`)
				match := re.FindStringSubmatch(dollarRefNode.Value)
				if len(match) > 1 {
					requestBodyParam = match[1]
				}

			}

			addPropertyToSchema(inputSchema, requestBodyParam, requestContentSchemaNode)
		}
	}

	return requestContentType, requestBodyParam, requestContentSchemaNode, inputSchema, nil
}

// combineParams takes parameters from within the operation level and the path level,
// and combines them into a single array containing unique parameters.
// If a parameter is defined at both levels, the operation level definition takes precedence.
func combineParams(operationParamsNode *yaml.Node, pathParamsNode *yaml.Node,
	inlinedOperationParamsNode *yaml.Node, inlinedPathParamsNode *yaml.Node, oas3Node *yaml.Node) ([]*yaml.Node, error) {
	var combinedParamNodes []*yaml.Node
	var err error

	if operationParamsNode != nil {
		combinedParamNodes = append(combinedParamNodes, operationParamsNode.Content...)
	}

	//combine operation and path level params
	if pathParamsNode != nil {
		//only append  params that are not already listed in the operation level

		for _, pathParamNode := range pathParamsNode.Content {
			var inlinedPathParamNode = pathParamNode
			if isRefValueMap(inlinedPathParamNode) {
				if inlinedPathParamNode, err = InlineYAMLReferences(pathParamNode, oas3Node); err != nil {
					return nil, err
				}
			}

			var pathParamNodeName *yaml.Node
			if pathParamNodeName, err = GetChildNodeByJSONPath(inlinedPathParamNode, "$.name"); err != nil {
				return nil, err
			}

			if inlinedOperationParamsNode == nil {
				combinedParamNodes = append(combinedParamNodes, pathParamNode)
			} else {
				var foundOperationParamNode *yaml.Node
				if foundOperationParamNode, err = GetChildNodeByJSONPath(inlinedOperationParamsNode, fmt.Sprintf("$[?(@.name == '%s')]", pathParamNodeName.Value)); err != nil {
					return nil, err
				}

				if foundOperationParamNode == nil {
					//parameter does not already exist at the operation level, add it
					combinedParamNodes = append(combinedParamNodes, pathParamNode)
				}
			}
		}

	}
	return combinedParamNodes, nil
}
