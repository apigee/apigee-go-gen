// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"fmt"
	"github.com/beevik/etree"
	"github.com/go-errors/errors"
	libopenapijson "github.com/pb33f/libopenapi/json"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func YAMLText2XMLText(reader io.Reader) ([]byte, error) {
	var err error
	var yamlNode *yaml.Node
	if yamlNode, err = Text2YAML(reader); err != nil {
		return nil, err
	}

	docNode := &yaml.Node{Kind: yaml.DocumentNode}
	docNode.Content = append(docNode.Content, yamlNode)

	var xmlText []byte
	if xmlText, err = YAML2XMLText(docNode); err != nil {
		return nil, err
	}
	return xmlText, nil
}

func YAML2XML(node *yaml.Node) (*etree.Document, error) {
	var err error
	doc := etree.NewDocument()
	if _, err = YAML2XMLRecursive(node, &doc.Element); err != nil {
		return nil, err
	}
	return doc, nil

}

func YAML2Text(node *yaml.Node, indent int) ([]byte, error) {
	var buffer bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buffer)
	yamlEncoder.SetIndent(indent)
	err := yamlEncoder.Encode(node)
	if err != nil {
		return nil, errors.New(err)
	}

	return buffer.Bytes(), nil
}

func Text2YAML(reader io.Reader) (*yaml.Node, error) {
	var err error
	decoder := yaml.NewDecoder(reader)
	yamlNode := yaml.Node{}
	if err = decoder.Decode(&yamlNode); err != nil {
		return nil, errors.New(err)
	}

	filePath := "./input.yaml"
	resultNode, err := YAMLResolveRefs(&yamlNode, filePath, false)
	if err != nil {
		return nil, err
	}

	if resultNode.Kind == yaml.DocumentNode {
		//unwrap the content from document to make things simpler/consistent
		return resultNode.Content[0], nil
	}

	return resultNode, nil

}

func YAML2XMLText(node *yaml.Node) ([]byte, error) {
	var err error
	var doc *etree.Document
	if doc, err = YAML2XML(node); err != nil {
		return nil, err
	}

	return XML2Text(doc)
}

func YAMLText2XML(reader io.Reader) (*etree.Document, error) {
	var err error
	var yamlNode *yaml.Node

	if yamlNode, err = Text2YAML(reader); err != nil {
		return nil, err
	}

	return YAML2XML(yamlNode)
}

func YAML2XMLRecursive(node *yaml.Node, parent *etree.Element) (*etree.Element, error) {
	if node == nil {
		return nil, nil
	}

	if node.Kind == yaml.DocumentNode {
		parent.CreateProcInst("xml", `version="1.0" encoding="UTF-8" standalone="yes"`)

		if len(node.Content) == 0 {
			return parent, nil
		}
		return YAML2XMLRecursive(node.Content[0], parent)
	} else if node.Kind == yaml.ScalarNode {
		if parent != nil {
			parent.CreateText(node.Value)
		}

		return nil, nil
	} else if node.Kind == yaml.SequenceNode {
		for i := 0; i < len(node.Content); i += 1 {
			_, _ = YAML2XMLRecursive(node.Content[i], parent)
		}
		return nil, nil
	} else if node.Kind == yaml.MappingNode {
		if parent != nil {
			for i := 0; i+1 < len(node.Content); i += 2 {
				if len(node.Content[i].Value) > 1 && node.Content[i].Value[0] == '.' &&
					node.Content[i+1].Kind == yaml.ScalarNode {
					parent.CreateAttr(node.Content[i].Value[1:], node.Content[i+1].Value)
				}
			}
		}

		for i := 0; i+1 < len(node.Content); i += 2 {
			if len(node.Content[i].Value) > 1 && node.Content[i].Value[0] == '.' {
				continue
			} else if strings.Index(node.Content[i].Value, "-") == 0 {
				_, _ = YAML2XMLRecursive(node.Content[i+1], parent)
			} else {
				child := parent.CreateElement(node.Content[i].Value)
				_, _ = YAML2XMLRecursive(node.Content[i+1], child)
			}
		}

		return nil, nil
	}

	return nil, fmt.Errorf("unknown yaml node kind %v", node.Kind)
}

func YAMLDoc2File(docNode *yaml.Node, outputFile string) error {
	var err error
	var docBytes []byte
	if docBytes, err = YAML2Text(docNode, 2); err != nil {
		return err
	}

	//generate output directory
	outputDir := filepath.Dir(outputFile)
	if err = os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return errors.New(err)
	}

	//generate the main YAML file
	if err = os.WriteFile(outputFile, docBytes, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func YAMLFile2YAML(filePath string) (*yaml.Node, error) {
	var file *os.File
	var err error
	if file, err = os.Open(filePath); err != nil {
		return nil, errors.New(err)
	}
	defer func() { MustClose(file) }()

	//switch to directory relative to the YAML file so that JSON $refs are valid
	popd := PushDir(filepath.Dir(filePath))
	defer popd()

	dataNode, err := Text2YAML(file)
	if err != nil {
		return nil, err
	}

	return dataNode, nil
}

func UnFlowYAMLNode(node *yaml.Node) *yaml.Node {
	switch node.Kind {
	case yaml.DocumentNode:
		fallthrough
	case yaml.SequenceNode:
		node.Style = 0
		for _, v := range node.Content {
			v.Style = 0
			UnFlowYAMLNode(v)
		}
	case yaml.MappingNode:
		node.Style = 0
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i]
			val := node.Content[i+1]
			key.Style = 0
			if key.Value == "$ref" {
				continue
			}
			UnFlowYAMLNode(val)
		}
	case yaml.ScalarNode:
		node.Style = 0
	}

	return node
}

func YAMLFile2XMLFile(input string, output string) error {
	text, err := ReadInputText(input)
	if err != nil {
		return err
	}

	outputText, err := YAMLText2XMLText(bytes.NewReader(text))
	if err != nil {
		return errors.New(err)
	}

	return WriteOutputText(output, outputText)
}

func YAMLText2JSONText(reader io.Reader) ([]byte, error) {
	text, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	//unmarshall input as YAML
	var yamlNode *yaml.Node
	yamlNode = &yaml.Node{}
	err = yaml.Unmarshal(text, yamlNode)
	if err != nil {
		return nil, errors.New(err)
	}

	//convert back to JSON text
	outputText, err := libopenapijson.YAMLNodeToJSON(yamlNode, "  ")
	if err != nil {
		return nil, errors.New(err)
	}

	return outputText, nil
}

func YAMLFile2JSONFile(input string, output string) error {
	text, err := ReadInputText(input)
	if err != nil {
		return err
	}

	outputText, err := YAMLText2JSONText(bytes.NewReader(text))
	if err != nil {
		return err
	}

	return WriteOutputText(output, outputText)
}
