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
	"github.com/go-errors/errors"
	libopenapijson "github.com/pb33f/libopenapi/json"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"strings"
)

func RemoveExtensions(input string, output string) error {
	text, err := ReadInputText(input)
	if err != nil {
		return err
	}

	//unmarshall input as YAML
	var yamlNode *yaml.Node
	yamlNode = &yaml.Node{}
	err = yaml.Unmarshal(text, yamlNode)
	if err != nil {
		return errors.New(err)
	}

	yamlNode, err = RemoveOASExtensions(yamlNode)
	if err != nil {
		return err
	}

	//convert back to text
	ext := filepath.Ext(output)
	if ext == "" {
		ext = filepath.Ext(input)
	}

	//depending on the file extension write output as either JSON or YAML
	var outputText []byte
	if ext == ".json" {
		outputText, err = libopenapijson.YAMLNodeToJSON(yamlNode, "  ")
		if err != nil {
			return errors.New(err)
		}
	} else {
		outputText, err = YAML2Text(UnFlowYAMLNode(yamlNode), 2)
		if err != nil {
			return err
		}
	}

	return WriteOutputText(output, outputText)
}

func RemoveOASExtensions(root *yaml.Node) (*yaml.Node, error) {
	modified, err := RemoveOASExtensionsRecursive(root, "")
	if err != nil {
		return nil, err
	}
	return modified, nil
}

func RemoveOASExtensionsRecursive(node *yaml.Node, parentField string) (*yaml.Node, error) {
	var err error

	if node == nil {
		return nil, errors.Errorf("nil node detected")
	}

	var modifiedNode *yaml.Node

	if node.Kind == yaml.MappingNode && isYAMLRef(node) {
		if err != nil {
			return nil, err
		}
		return node, nil
	} else if node.Kind == yaml.MappingNode {
		var newContent []*yaml.Node
		for i := 0; i+1 < len(node.Content); i += 2 {
			fieldName := node.Content[i].Value
			if strings.Index(fieldName, "x-") == 0 &&
				!(parentField == "headers" ||
					parentField == "properties" ||
					parentField == "responses" ||
					parentField == "schemas" ||
					parentField == "paths" ||
					parentField == "variables" ||
					parentField == "securitySchemes" ||
					parentField == "examples" ||
					parentField == "links" ||
					parentField == "callbacks" ||
					parentField == "requestBodies" ||
					parentField == "mapping" ||
					parentField == "scopes" ||
					parentField == "encoding" ||
					parentField == "definitions" ||
					parentField == "securityDefinitions" ||
					parentField == "parameters") {
				continue
			}
			if modifiedNode, err = RemoveOASExtensionsRecursive(node.Content[i+1], fieldName); err != nil {
				return nil, err
			}
			newContent = append(newContent, node.Content[i], modifiedNode)
		}
		node.Content = newContent
		return node, nil
	} else if node.Kind == yaml.DocumentNode {
		if modifiedNode, err = RemoveOASExtensionsRecursive(node.Content[0], ""); err != nil {
			return nil, err
		}
		node.Content[0] = modifiedNode
		return node, nil
	} else if node.Kind == yaml.SequenceNode {
		for i := 0; i < len(node.Content); i += 1 {
			if modifiedNode, err = RemoveOASExtensionsRecursive(node.Content[i], ""); err != nil {
				return nil, err
			}
			node.Content[i] = modifiedNode
		}
		return node, nil
	}

	return node, nil
}

func RemoveSchemaExtensions(input string, output string) error {
	text, err := ReadInputText(input)
	if err != nil {
		return err
	}

	//unmarshall input as YAML
	var yamlNode *yaml.Node
	yamlNode = &yaml.Node{}
	err = yaml.Unmarshal(text, yamlNode)
	if err != nil {
		return errors.New(err)
	}

	yamlNode, err = RemoveOASSchemaExtensions(yamlNode)
	if err != nil {
		return err
	}

	//convert back to text
	ext := filepath.Ext(output)
	if ext == "" {
		ext = filepath.Ext(input)
	}

	//depending on the file extension write output as either JSON or YAML
	var outputText []byte
	if ext == ".json" {
		outputText, err = libopenapijson.YAMLNodeToJSON(yamlNode, "  ")
		if err != nil {
			return errors.New(err)
		}
	} else {
		outputText, err = YAML2Text(UnFlowYAMLNode(yamlNode), 2)
		if err != nil {
			return err
		}
	}

	return WriteOutputText(output, outputText)
}

func RemoveOASSchemaExtensions(root *yaml.Node) (*yaml.Node, error) {
	var err error

	// The root of the document is typically a DocumentNode holding the main MappingNode
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		rootMapping := root.Content[0]

		// 1. Attempt to find and clean OpenAPI 3.x schemas: /components/schemas
		if componentsNode := findMappingNode(rootMapping, []string{"components"}); componentsNode != nil {
			if schemasNode := findMappingNode(componentsNode, []string{"schemas"}); schemasNode != nil {
				if _, err = removeSchemaExtensionsRecursive(schemasNode, ""); err != nil {
					return nil, errors.Errorf("error cleaning components/schemas: %w", err)
				}
			}
		}

		// 2. Attempt to find and clean OpenAPI 2.0 (Swagger) definitions: /definitions
		if definitionsNode := findMappingNode(rootMapping, []string{"definitions"}); definitionsNode != nil {
			if _, err = removeSchemaExtensionsRecursive(definitionsNode, ""); err != nil {
				return nil, errors.Errorf("error cleaning definitions: %w", err)
			}
		}

		// 3. Handle potential $defs (Draft 2020-12/OAS 3.1)
		if defsNode := findMappingNode(rootMapping, []string{"$defs"}); defsNode != nil {
			if _, err = removeSchemaExtensionsRecursive(defsNode, ""); err != nil {
				return nil, errors.Errorf("error cleaning $defs: %w", err)
			}
		}
	}

	return root, nil
}

// removeExtensionsFromSchemaNode recursively traverses a schema-like node (e.g., /schemas or /definitions)
// and removes any 'x-' fields unless the immediate parent is "properties".
func removeSchemaExtensionsRecursive(node *yaml.Node, parentField string) (*yaml.Node, error) {
	if node == nil {
		return nil, nil
	}

	if node.Kind == yaml.MappingNode {
		var newContent []*yaml.Node
		for i := 0; i+1 < len(node.Content); i += 2 {
			fieldNameNode := node.Content[i]
			valueNode := node.Content[i+1]
			fieldName := fieldNameNode.Value

			// 1. Check if it's an x- extension
			if strings.HasPrefix(fieldName, "x-") {
				// CRITICAL EXCEPTION: DO NOT remove if the parent is "properties"
				if parentField != "properties" {
					// Remove the x- extension by continuing the loop without appending
					continue
				}
				// If parentField IS "properties", we fall through to process the value recursively
			}

			// 2. If the field is $ref, keep it but do not recurse into its value
			// (since the value is a scalar string reference, not a structure to traverse).
			// This is safe and prevents issues if the ref is malformed or circular.
			if fieldName == "$ref" {
				newContent = append(newContent, fieldNameNode, valueNode)
				continue
			}

			// 3. Regular field or a preserved x- field (under "properties"), recurse and append
			modifiedValueNode, err := removeSchemaExtensionsRecursive(valueNode, fieldName)
			if err != nil {
				return nil, err
			}
			newContent = append(newContent, fieldNameNode, modifiedValueNode)
		}
		node.Content = newContent
		return node, nil
	}

	// For SequenceNode (Arrays) like `required: [...]` or `allOf: [...]`
	if node.Kind == yaml.SequenceNode {
		for i := 0; i < len(node.Content); i += 1 {
			modifiedNode, err := removeSchemaExtensionsRecursive(node.Content[i], "")
			if err != nil {
				return nil, err
			}
			node.Content[i] = modifiedNode
		}
		return node, nil
	}

	// For DocumentNode, ScalarNode, and others, return as is.
	return node, nil
}

// findMappingNode navigates a YAML mapping node by path (keys) and returns the final node if found.
func findMappingNode(node *yaml.Node, path []string) *yaml.Node {
	current := node
	for _, key := range path {
		if current == nil || current.Kind != yaml.MappingNode {
			return nil
		}

		found := false
		// Mapping nodes have Content as [Key1, Value1, Key2, Value2, ...]
		for i := 0; i < len(current.Content); i += 2 {
			if current.Content[i].Value == key {
				current = current.Content[i+1]
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return current
}
