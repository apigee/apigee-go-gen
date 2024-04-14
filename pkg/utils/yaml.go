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
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func YAMLText2XMLText(reader io.Reader) ([]byte, error) {
	var err error
	var yamlNode *yaml.Node
	if yamlNode, err = Text2YAML(reader); err != nil {
		PrintErrorWithStackAndExit(err)
		return nil, err
	}

	docNode := &yaml.Node{Kind: yaml.DocumentNode}
	docNode.Content = append(docNode.Content, yamlNode)

	var xmlText []byte
	if xmlText, err = YAML2XMLText(docNode); err != nil {
		PrintErrorWithStackAndExit(err)
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
	yamlEncoder.Encode(node)

	return buffer.Bytes(), nil
}

func Text2YAML(reader io.Reader) (*yaml.Node, error) {
	var err error
	decoder := yaml.NewDecoder(reader)
	yamlNode := yaml.Node{}
	if err = decoder.Decode(&yamlNode); err != nil {
		return nil, errors.New(err)
	}

	resultNode, err := ResolveYAMLRefs(&yamlNode, ".")
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

func isYAMLRef(node *yaml.Node) bool {
	if node == nil {
		return false
	}

	return node.Kind == yaml.MappingNode &&
		len(node.Content) == 2 &&
		node.Content[0].Value == "$ref"
}

func ParseYAMLFile(filePath string) (*yaml.Node, error) {
	var err error

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, errors.New(err)
	}

	rootNode, ok := ParsedYAMLFiles[absPath]
	if ok {
		return rootNode, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New(err)
	}

	decoder := yaml.NewDecoder(file)
	yamlNode := yaml.Node{}
	if err = decoder.Decode(&yamlNode); err != nil {
		return nil, errors.Errorf("could not parse %s. %s", filePath, err.Error())
	}

	resolvedNode, err := ResolveYAMLRefs(&yamlNode, filepath.Dir(absPath))
	if err != nil {
		return nil, err
	}

	ParsedYAMLFiles[filePath] = resolvedNode
	return resolvedNode, nil
}

func JSONPointer2JSONPath(jsonPointer string) (jsonPath string, err error) {

	pointer := strings.TrimSpace(jsonPointer)
	if pointer == "" ||
		pointer == "#" ||
		pointer == "#/" {
		return "$", nil
	}

	if strings.Index(pointer, "#/") != 0 {
		return "", errors.Errorf("relative JSONPointer %s is not supported", jsonPointer)
	}

	pointer = "$" + strings.ReplaceAll(pointer[1:], "/", ".")
	return pointer, nil
}

func SplitJSONRef(refStr string) (location string, jsonPath string, err error) {
	parsedUrl, err := url.Parse(refStr)
	if err != nil {
		return "", "", errors.New(err)
	}

	if parsedUrl.Scheme != "" {
		return "", "", errors.Errorf("JSONRef %s is not supported", refStr)
	}

	jsonPath, err = JSONPointer2JSONPath("#" + parsedUrl.Fragment)
	if err != nil {
		return "", "", err
	}

	return parsedUrl.Path, jsonPath, nil
}

func ResolveYAMLRef(node *yaml.Node, nodePath string) (*yaml.Node, error) {
	var err error

	jsonRef := node.Content[1].Value
	refFilePath, refJSONPath, err := SplitJSONRef(node.Content[1].Value)
	if err != nil {
		return nil, err
	}

	if refFilePath == "" {
		return nil, errors.Errorf("self referncing JSONRef '%s' is not supported", jsonRef)
	}
	var fileRootNode *yaml.Node
	if !filepath.IsAbs(refFilePath) {
		refFilePath = filepath.Join(nodePath, refFilePath)
	}

	if fileRootNode, err = ParseYAMLFile(refFilePath); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	var yamlPath *yamlpath.Path
	if yamlPath, err = yamlpath.NewPath(refJSONPath); err != nil {
		return nil, errors.New(err)
	}

	var yamlNodes []*yaml.Node
	if yamlNodes, err = yamlPath.Find(fileRootNode); err != nil {
		return nil, errors.New(err)
	}

	if len(yamlNodes) == 0 {
		return nil, errors.Errorf("no node found at JSONRef '%s'", jsonRef)
	}

	if len(yamlNodes) > 1 {
		return nil, errors.Errorf("more than one node found at JSONRef '%s'", jsonRef)
	}

	return yamlNodes[0], nil
}

func ResolveYAMLRefs(node *yaml.Node, nodePath string) (*yaml.Node, error) {
	if node == nil {
		return nil, nil
	}

	var resolvedNode *yaml.Node
	var err error

	if node.Kind == yaml.MappingNode && isYAMLRef(node) {
		if resolvedNode, err = ResolveYAMLRef(node, nodePath); err != nil {
			return nil, err
		}
		return resolvedNode, nil
	} else if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			if resolvedNode, err = ResolveYAMLRefs(node.Content[i+1], nodePath); err != nil {
				return nil, err
			}
			node.Content[i+1] = resolvedNode
		}
		return node, nil
	} else if node.Kind == yaml.SequenceNode || node.Kind == yaml.DocumentNode {
		for i := 0; i < len(node.Content); i += 1 {
			if resolvedNode, err = ResolveYAMLRefs(node.Content[i], nodePath); err != nil {
				return nil, err
			}
			node.Content[i] = resolvedNode
		}
		return node, nil
	}

	return node, nil
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

var ParsedYAMLFiles map[string]*yaml.Node

func init() {
	ParsedYAMLFiles = make(map[string]*yaml.Node)
}
