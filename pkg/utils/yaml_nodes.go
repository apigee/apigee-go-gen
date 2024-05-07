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

import "gopkg.in/yaml.v3"

func NewRefNode(refJSONPath string) *yaml.Node {
	mapNode := NewMapNode()
	mapNode.Content = append(mapNode.Content, NewStringNode("$ref", yaml.SingleQuotedStyle))
	mapNode.Content = append(mapNode.Content, NewStringNode(refJSONPath, yaml.SingleQuotedStyle))
	return mapNode
}

func NewStringNode(value string, style yaml.Style) *yaml.Node {
	strNode := yaml.Node{Kind: yaml.ScalarNode, Style: style}
	strNode.SetString(value)
	return &strNode
}

func NewMapNode() *yaml.Node {
	mapNode := yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	return &mapNode
}

func GetFieldOrCreateNew(node *yaml.Node, key string, value *yaml.Node) *yaml.Node {

	for i := 0; i < len(node.Content); i += 2 {
		curKey := node.Content[i]
		curVal := node.Content[i+1]

		if curKey.Value == key {
			return curVal
		}
	}

	node.Content = append(node.Content, NewStringNode(key, 0))

	node.Content = append(node.Content, value)
	return value
}

func GetDocMapRoot(yamlNode *yaml.Node) *yaml.Node {
	if yamlNode.Kind != yaml.DocumentNode {
		return nil
	}

	if len(yamlNode.Content) == 0 {
		return nil
	}

	root := yamlNode.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil
	}

	return root
}
