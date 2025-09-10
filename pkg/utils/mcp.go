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

package utils

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	"regexp"
	"slices"
	"strings"
)

type MCPTool struct {
	Name         string     `yaml:"name"`
	Title        string     `yaml:"title"`
	Description  string     `yaml:"description"`
	InputSchema  *yaml.Node `yaml:"inputSchema,omitempty"`
	OutputSchema *yaml.Node `yaml:"outputSchema,omitempty"`
}

type MCPToolTarget struct {
	Verb        string `yaml:"verb"`
	PathSuffix  string `yaml:"pathSuffix"`
	ContentType string `yaml:"contentType"`
}

type MCPValuesFile struct {
	ToolsList    []*MCPTool                `yaml:"tools_list"`
	ToolsTargets map[string]*MCPToolTarget `yaml:"tools_targets"`
}

func OAS3ToMCPValues(file string) (mcpValuesMap map[string]any, err error) {
	var input []byte
	if input, err = ReadInputText(file); err != nil {
		return nil, err
	}

	var node *yaml.Node
	node = &yaml.Node{}
	if err = yaml.Unmarshal(input, node); err != nil {
		return nil, errors.New(err)
	}

	//resolve references
	if node, err = YAMLResolveAllRefs(node, file, false); err != nil {
		return nil, err
	}

	//verify we are actually working with OAS3
	if slices.IndexFunc(node.Content[0].Content, func(n *yaml.Node) bool {
		return n.Value == "openapi"
	}) < 0 {
		return nil, errors.Errorf("input file '%s' is not an OpenAPI 3.X Description", file)
	}

	var paths *yaml.Node

	//get the paths
	if paths, err = GetChildNodeByJSONPath(node, "$.paths"); err != nil {
		return nil, err
	}

	if paths == nil {
		return nil, errors.Errorf("OpenAPI description does not contain any paths")
	}

	var mcpToolsList []*MCPTool
	mcpToolsTargets := make(map[string]*MCPToolTarget)

	for i := 0; i+1 < len(paths.Content); i += 2 {
		path := paths.Content[i].Value
		pathNode := paths.Content[i+1]
		for j := 0; j+1 < len(pathNode.Content); j += 2 {
			verb := strings.ToUpper(pathNode.Content[j].Value)
			verbNode := pathNode.Content[j+1]
			if !(verb == "POST" || verb == "PUT" || verb == "GET" || verb == "DELETE") {
				continue
			}

			var paramsNode *yaml.Node
			var requestBodyNode *yaml.Node
			var responsesNode *yaml.Node
			var operationIdNode *yaml.Node
			var descriptionNode *yaml.Node
			var summaryNode *yaml.Node

			if operationIdNode, err = GetChildNodeByJSONPath(verbNode, "$.operationId"); err != nil {
				return nil, err
			}

			if paramsNode, err = GetChildNodeByJSONPath(verbNode, "$.parameters"); err != nil {
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

			var operationJSONPath = fmt.Sprintf("$.paths.%s.%s", path, verb)
			if operationIdNode == nil {
				return nil, errors.Errorf("Operation at %s is missing 'operationId' property", operationJSONPath)
			}

			if verb == "POST" || verb == "PUT" {
				if requestBodyNode == nil {
					return nil, errors.Errorf("Operation at %s is missing 'requestBody' property", operationJSONPath)
				}

				if responsesNode == nil {
					return nil, errors.Errorf("Operation at %s is missing 'responses' property", operationJSONPath)
				}
			}

			operationId := operationIdNode.Value
			summary := summaryNode.Value
			description := descriptionNode.Value

			//check if path has parameters
			var pathParams []string
			var paramsRegex = regexp.MustCompile(`(?m)\{([^}]+)}`)
			paramsMatches := paramsRegex.FindAllStringSubmatch(path, -1)
			for _, match := range paramsMatches {
				pathParams = append(pathParams, match[1])
			}

			//there has to be a parameters property
			if len(pathParams) > 0 && paramsNode == nil {
				return nil, errors.Errorf("Operation at %s uses path parameters, but has no 'parameters' property", operationJSONPath)
			}

			//each parameter must be defined
			for _, pathParam := range pathParams {
				var pathParamNode *yaml.Node
				if pathParamNode, err = GetChildNodeByJSONPath(paramsNode, fmt.Sprintf("$[?(@.name == '%s')]", pathParam)); err != nil {
					return nil, err
				}
				if pathParamNode == nil {
					return nil, errors.Errorf("Operation at %s is missing the '%s' path parameter definition", operationJSONPath, pathParam)
				}
			}

			inputSchema := createSchemaEntry("object")

			//process parameters
			pathParamsSchema := createSchemaEntry("object")
			queryParamsSchema := createSchemaEntry("object")
			headerParamsSchema := createSchemaEntry("object")

			hasPathParams := false
			hasQueryParams := false
			hasHeaderParams := false
			if paramsNode != nil {
				for k, paramNode := range paramsNode.Content {
					var paramNameNode *yaml.Node
					var paramInNode *yaml.Node
					var paramSchemaNode *yaml.Node

					if paramNameNode, err = GetChildNodeByJSONPath(paramNode, "$.name"); err != nil {
						return nil, err
					}

					if paramInNode, err = GetChildNodeByJSONPath(paramNode, "$.in"); err != nil {
						return nil, err
					}

					if paramSchemaNode, err = GetChildNodeByJSONPath(paramNode, "$.schema"); err != nil {
						return nil, err
					}

					if paramNameNode == nil {
						return nil, errors.Errorf("Parameter #%d within the '%s' operation is missing the 'in' property", k, operationId)
					}

					if paramInNode == nil {
						return nil, errors.Errorf("Parameter '%s' within the '%s' operation is missing the 'in' property", paramNameNode.Value, operationJSONPath)
					}

					if paramSchemaNode == nil {
						return nil, errors.Errorf("Parameter '%s' within the '%s' operation is missing the 'schema' property", paramNameNode.Value, operationJSONPath)
					}

					switch paramInNode.Value {
					case "header":
						hasHeaderParams = true
						addPropertyToSchema(headerParamsSchema, paramNameNode.Value, paramSchemaNode)
						break
					case "path":
						hasPathParams = true
						addPropertyToSchema(pathParamsSchema, paramNameNode.Value, paramSchemaNode)
						break
					case "query":
						hasQueryParams = true
						addPropertyToSchema(queryParamsSchema, paramNameNode.Value, paramSchemaNode)
						break
					}
				}

				if hasHeaderParams {
					addPropertyToSchema(inputSchema, "header_params", headerParamsSchema)
				}

				if hasPathParams {
					addPropertyToSchema(inputSchema, "path_params", pathParamsSchema)
				}

				if hasQueryParams {
					addPropertyToSchema(inputSchema, "query_params", queryParamsSchema)
				}

			}

			//process request body
			var requestContentType string
			if requestBodyNode != nil {
				var requestBodyContent *yaml.Node
				if requestBodyContent, err = GetChildNodeByJSONPath(requestBodyNode, "$.content"); err != nil {
					return nil, err
				}

				if requestBodyContent == nil || len(requestBodyContent.Content) < 2 {
					return nil, errors.Errorf("The 'requestBody' property witin the '%s' operation must have at least one content element", operationId)
				}

				var contentSchemaNode *yaml.Node
				if len(requestBodyContent.Content) >= 2 {
					//find the JSON content element, or get the first one
					for c := 0; c+1 < len(requestBodyContent.Content); c += 1 {
						var curSchemaNode *yaml.Node

						curContentType := requestBodyContent.Content[c].Value
						if curSchemaNode, err = GetChildNodeByJSONPath(requestBodyContent.Content[c+1], "$.schema"); err != nil {
							return nil, errors.Errorf("The 'requestBody.%s.schema' is missing witin the '%s' operation", curContentType, operationId)
						}

						if c == 0 {
							requestContentType = curContentType
							contentSchemaNode = curSchemaNode
						}

						if strings.Contains(curContentType, "json") {
							requestContentType = curContentType
							contentSchemaNode = curSchemaNode
							break
						}
					}

					addPropertyToSchema(inputSchema, "request_body", contentSchemaNode)
				}

			}

			//fmt.Printf("path = %s, verb = %s, operationId = %s, pathParams = %v\n", path, verb, operationId, pathParams)
			mcpToolsList = append(mcpToolsList, &MCPTool{
				Name:         operationId,
				Title:        summary,
				Description:  description,
				InputSchema:  inputSchema,
				OutputSchema: nil,
			})

			mcpToolsTargets[operationId] = &MCPToolTarget{
				Verb:        verb,
				PathSuffix:  path,
				ContentType: requestContentType,
			}

		}
	}

	valuesFile := &MCPValuesFile{
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

func createMapEntry(parent *yaml.Node, key string, value *yaml.Node) *yaml.Node {
	parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, value)
	return value
}

func createSchemaEntry(schemaType string) *yaml.Node {
	schemaNode := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{}}
	createMapEntry(schemaNode, "type", &yaml.Node{Kind: yaml.ScalarNode, Value: schemaType})
	schemaProperties := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{}}
	createMapEntry(schemaNode, "properties", schemaProperties)

	return schemaNode
}

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

func GetChildNodeByJSONPath(root *yaml.Node, jsonPath string) (*yaml.Node, error) {
	return GetChildNodeByJSONPathOrDefault(root, jsonPath, nil)
}
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
