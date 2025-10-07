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
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/go-errors/errors"
	"gopkg.in/yaml.v3"
	"strings"
)

type Tool struct {
	Name         string     `yaml:"name"`
	Title        string     `yaml:"title"`
	Description  string     `yaml:"description"`
	InputSchema  *yaml.Node `yaml:"inputSchema,omitempty"`
	OutputSchema *yaml.Node `yaml:"outputSchema,omitempty"`
}

type ToolTarget struct {
	Verb           string     `yaml:"verb"`
	PathSuffix     string     `yaml:"pathSuffix"`
	ContentType    string     `yaml:"contentType"`
	Accept         string     `yaml:"accept"`
	QueryParams    []string   `yaml:"queryParams"`
	HeaderParams   []string   `yaml:"headerParams"`
	PathParams     []string   `yaml:"pathParams"`
	PayloadParam   string     `yaml:"payloadParam"`
	PayloadSchema  *yaml.Node `yaml:"payloadSchema"`
	ResponseSchema *yaml.Node `yaml:"responseSchema"`
}

type ValuesFile struct {
	ToolsList    []*Tool                `yaml:"tools_list"`
	ToolsTargets map[string]*ToolTarget `yaml:"tools_targets"`
}

func OAS3ToMCPValues(file string) (mcpValuesMap map[string]any, err error) {
	var input []byte
	if input, err = utils.ReadInputText(file); err != nil {
		return nil, err
	}

	var oas3Node *yaml.Node

	oas3Node = &yaml.Node{}
	if err = yaml.Unmarshal(input, oas3Node); err != nil {
		return nil, errors.New(err)
	}

	var versionNode *yaml.Node
	if versionNode, err = GetChildNodeByJSONPathOrDefault(oas3Node, "$.openapi", nil); err != nil {
		return nil, errors.Errorf("could not find 'openapi' field in file '%s': %s", file, err.Error())
	}

	if versionNode == nil {
		return nil, errors.Errorf("input file '%s' does not contain 'openapi' field", file)
	}

	if strings.Index(versionNode.Value, "3") != 0 {
		return nil, errors.Errorf("input file '%s' is not an OpenAPI 3.x description", file)
	}

	var paths *yaml.Node

	//get the paths
	if paths, err = GetChildNodeByJSONPath(oas3Node, "$.paths"); err != nil {
		return nil, err
	}

	if paths == nil {
		return nil, errors.Errorf("OpenAPI description does not contain any paths")
	}

	var mcpToolsList []*Tool
	mcpToolsTargets := make(map[string]*ToolTarget)

	for i := 0; i+1 < len(paths.Content); i += 2 {
		path := paths.Content[i].Value
		pathNode := paths.Content[i+1]

		var pathParamsNode *yaml.Node //parameters defined at path level
		if pathParamsNode, err = GetChildNodeByJSONPath(pathNode, "$.parameters"); err != nil {
			return nil, err
		}

		for j := 0; j+1 < len(pathNode.Content); j += 2 {
			verb := strings.ToUpper(pathNode.Content[j].Value)
			verbNode := pathNode.Content[j+1]
			if !(verb == "POST" || verb == "PUT" || verb == "GET" || verb == "DELETE") {
				continue
			}

			var operationParamsNode *yaml.Node //parameters defined at operation level
			var requestBodyNode *yaml.Node
			var responsesNode *yaml.Node
			var operationIdNode *yaml.Node
			var descriptionNode *yaml.Node
			var summaryNode *yaml.Node

			if operationIdNode, err = GetChildNodeByJSONPath(verbNode, "$.operationId"); err != nil {
				return nil, err
			}

			if operationParamsNode, err = GetChildNodeByJSONPath(verbNode, "$.parameters"); err != nil {
				return nil, err
			}

			if requestBodyNode, err = GetChildNodeByJSONPath(verbNode, "$.requestBody"); err != nil {
				return nil, err
			}

			if responsesNode, err = GetChildNodeByJSONPath(verbNode, "$.responses"); err != nil {
				return nil, err
			}

			if descriptionNode, err = GetChildNodeByJSONPathOrDefault(verbNode, "$.description", &yaml.Node{Kind: yaml.ScalarNode, Value: ""}); err != nil {
				return nil, err
			}

			if summaryNode, err = GetChildNodeByJSONPathOrDefault(verbNode, "$.summary", &yaml.Node{Kind: yaml.ScalarNode, Value: ""}); err != nil {
				return nil, err
			}

			var operationJSONPath = fmt.Sprintf("$.paths.%s.%s", path, strings.ToLower(verb))
			if operationIdNode == nil {
				operationIdNode = &yaml.Node{Kind: yaml.ScalarNode, Value: generateAlternateOperationID(path, verb)}
			}

			operationId := operationIdNode.Value
			summary := summaryNode.Value
			description := descriptionNode.Value

			//check if path has parameters
			pathParams := extractParamsFromPath(path)

			//there has to be a parameters property
			if len(pathParams) > 0 && (operationParamsNode == nil && pathParamsNode == nil) {
				return nil, errors.Errorf("Operation at %s uses path parameters, but has no 'parameters' property", operationJSONPath)
			}

			var inlinedOperationParamsNode *yaml.Node
			if inlinedOperationParamsNode, err = InlineYAMLReferences(operationParamsNode, oas3Node); err != nil {
				return nil, err
			}

			var inlinedPathParamsNode *yaml.Node
			if inlinedPathParamsNode, err = InlineYAMLReferences(pathParamsNode, oas3Node); err != nil {
				return nil, err
			}

			//cross-check to make sure all defined URL {...} parameters exist either at operation level, or path level
			if err = validatePathParams(operationJSONPath, pathParams, inlinedOperationParamsNode, inlinedPathParamsNode); err != nil {
				return nil, err
			}

			var inputSchema *yaml.Node //contains request body schema, and other stuff like headers, path params, and query params
			var requestContentType string
			var requestContentSchemaNode *yaml.Node //contains only the request body schema
			var requestBodyParam string

			if requestContentType, requestBodyParam, requestContentSchemaNode, inputSchema, err = processRequestBody(operationId, operationJSONPath, requestBodyNode, oas3Node); err != nil {
				return nil, err
			}

			//process parameters
			pathParamsList := []string{}
			queryParamsList := []string{}
			headerParamsList := []string{}

			if headerParamsList, pathParamsList, queryParamsList, err = processHeaderAndQueryParams(operationId, operationJSONPath, operationParamsNode, pathParamsNode, inlinedOperationParamsNode, inlinedPathParamsNode, oas3Node, inputSchema); err != nil {
				return nil, err
			}

			//output schema contains the response body schema
			var outputSchema *yaml.Node
			var responseContentType string
			var responseContentSchemaNode *yaml.Node

			if responseContentType, responseContentSchemaNode, outputSchema, err = processResponseBody(responsesNode, operationId); err != nil {
				return nil, err
			}

			//INFO: commenting this out for now, most MCP clients do not support $defs
			//if err = addDollarDefsToSchema(inputSchema, oas3Node); err != nil {
			//	return nil, err
			//}
			//
			//if err = addDollarDefsToSchema(outputSchema, oas3Node); err != nil {
			//	return nil, err
			//}

			//INFO: most MCP clients do not support $defs, in-line all schemas
			if inputSchema, err = InlineYAMLReferences(inputSchema, oas3Node); err != nil {
				return nil, err
			}

			if outputSchema, err = InlineYAMLReferences(outputSchema, oas3Node); err != nil {
				return nil, err
			}

			if requestContentSchemaNode, err = InlineYAMLReferences(requestContentSchemaNode, oas3Node); err != nil {
				return nil, err
			}

			if responseContentSchemaNode, err = InlineYAMLReferences(responseContentSchemaNode, oas3Node); err != nil {
				return nil, err
			}

			if outputSchema, err = addMissingTypeFieldToOutputSchema(outputSchema); err != nil {
				return nil, err
			}

			var cleanInputSchema *yaml.Node
			if inputSchema != nil {
				cleanInputSchema = DeepCloneYAML(inputSchema)
				//remove "xml" elements
				traverseAndDeleteXML(cleanInputSchema, "")
				//rewrite "example" (from old OpenAPI schema) into "examples" (from JSON-Schema)
				traverseAndRewriteExample(cleanInputSchema)
			}

			var cleanOutputSchema *yaml.Node
			if outputSchema != nil {
				cleanOutputSchema = DeepCloneYAML(outputSchema)
				//remove "xml" elements
				traverseAndDeleteXML(cleanOutputSchema, "")
				//rewrite "example" (from old OpenAPI schema) into "examples" (from JSON-Schema)
				traverseAndRewriteExample(cleanOutputSchema)
			}

			mcpToolsList = append(mcpToolsList, &Tool{
				Name:         operationId,
				Title:        summary,
				Description:  description,
				InputSchema:  cleanInputSchema,
				OutputSchema: cleanOutputSchema,
			})

			mcpToolsTargets[operationId] = &ToolTarget{
				Verb:           verb,
				PathSuffix:     path,
				ContentType:    requestContentType,
				Accept:         responseContentType,
				PathParams:     pathParamsList,
				QueryParams:    queryParamsList,
				HeaderParams:   headerParamsList,
				PayloadParam:   requestBodyParam,
				PayloadSchema:  requestContentSchemaNode,
				ResponseSchema: responseContentSchemaNode,
			}

		}
	}

	valuesFile := &ValuesFile{
		ToolsList:    mcpToolsList,
		ToolsTargets: mcpToolsTargets}

	var valuesFileContent []byte
	if valuesFileContent, err = yaml.Marshal(valuesFile); err != nil {
		return nil, errors.New(err)
	}

	valuesFileMap := make(map[string]any)

	err = yaml.Unmarshal(valuesFileContent, valuesFileMap)
	if err != nil {
		return nil, errors.New(err)
	}

	return valuesFileMap, err
}
