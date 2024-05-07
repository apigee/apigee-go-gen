//  Copyright 2024 Google LLC
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
