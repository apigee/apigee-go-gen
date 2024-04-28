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
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"slices"
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
	resultNode, err := ResolveYAMLRefs(&yamlNode, filePath, false)
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

	return node.Kind == yaml.MappingNode && slices.IndexFunc(node.Content, func(n *yaml.Node) bool {
		return n.Value == "$ref"
	}) >= 0

}

func getYAMLRefString(node *yaml.Node) string {
	index := slices.IndexFunc(node.Content, func(n *yaml.Node) bool {
		return n.Value == "$ref"
	})
	if index < 0 {
		return ""
	}
	return node.Content[index+1].Value
}

func ParseYAMLFile(filePath string) (*yaml.Node, error) {
	var err error

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, errors.New(err)
	}

	//check if the file has already been parsed
	rootNode, ok := ParsedYAMLFiles[absFilePath]
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

	ParsedYAMLFiles[filePath] = &yamlNode
	return &yamlNode, nil
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

func ResolveYAMLRef(root *yaml.Node,
	node *yaml.Node, nodePath string, nodePaths []string,
	filePath string, filePaths []string,
	allowCycles bool, resolveCyclesOnly bool) (*yaml.Node, bool, error) {
	var err error

	jsonRef := getYAMLRefString(node)
	if jsonRef == "" {
		return nil, false, errors.Errorf("JSONRef at %s is not valid", nodePath)
	}

	refFilePath, refJSONPath, err := SplitJSONRef(jsonRef)
	if err != nil {
		return nil, false, err
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, false, errors.New(err)
	}

	if refFilePath == "" {
		refFilePath = filepath.Base(filePath)
	}

	var absRefFilePath string
	if filepath.IsAbs(refFilePath) {
		absRefFilePath = refFilePath
	} else {
		absRefFilePath, err = filepath.Abs(filepath.Join(filepath.Dir(absFilePath), refFilePath))
	}

	absJSONRef := fmt.Sprintf("%s%s", absRefFilePath, refJSONPath)
	if _, ok := ResolvedRefs[absJSONRef]; ok {
		//return cached result
		return ResolvedRefs[absJSONRef], false, nil
	}

	//detect cycles
	if index := slices.Index(nodePaths, absJSONRef); index >= 0 {
		if !allowCycles {
			return nil, false, errors.New(NewCyclicJSONRefError(nodePaths))
		}
		ResolvedRefs[absJSONRef] = MakeCyclicRefPlaceholder(refJSONPath)
		return ResolvedRefs[absJSONRef], true, nil
	}

	isSelfRef := absRefFilePath == absFilePath

	if isSelfRef && len(filePaths) == 1 {
		//self ref at the first level, no need to de-reference it
		ResolvedRefs[absJSONRef] = node
		return ResolvedRefs[absJSONRef], false, nil
	} else if isSelfRef && len(filePaths) > 1 {
		//self ref below first level, need to de-reference it
		yamlNode, err := LocateRef(root, refJSONPath, jsonRef)

		newNodePaths := append([]string{}, nodePaths...)
		newNodePaths = append(newNodePaths, absJSONRef)

		resolvedNode, cycleDetected, err := ResolveYAMLRefsRecursive(root, yamlNode, nodePath, newNodePaths, absRefFilePath, filePaths, allowCycles, resolveCyclesOnly)
		if err != nil {
			return nil, cycleDetected, err
		}

		if resolveCyclesOnly && !cycleDetected {
			resolvedNode = node
		}

		ResolvedRefs[absJSONRef] = resolvedNode
		return ResolvedRefs[absJSONRef], cycleDetected, nil
	}

	//ref to a different file, need to de-reference it
	var refFileNode *yaml.Node
	refFileNode, err = ParseYAMLFile(absRefFilePath)
	if err != nil {
		return nil, false, err
	}

	yamlNode, err := LocateRef(refFileNode, refJSONPath, jsonRef)
	if err != nil {
		return nil, false, err
	}

	newFilePaths := append([]string{}, filePaths...)
	newFilePaths = append(newFilePaths, absRefFilePath)

	newNodePaths := append([]string{}, nodePaths...)
	newNodePaths = append(newNodePaths, absJSONRef)

	resolvedNode, cycleDetected, err := ResolveYAMLRefsRecursive(refFileNode, yamlNode, nodePath, newNodePaths, absRefFilePath, newFilePaths, allowCycles, resolveCyclesOnly)
	if err != nil {
		return nil, cycleDetected, err
	}

	if resolveCyclesOnly && !cycleDetected {
		resolvedNode = node
	}

	ResolvedRefs[absJSONRef] = resolvedNode
	return ResolvedRefs[absJSONRef], cycleDetected, nil
}

func MakeCyclicRefPlaceholder(refJSONPath string) *yaml.Node {
	cyclicRef := &yaml.Node{Kind: yaml.MappingNode}
	key := yaml.Node{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle}
	key.SetString("description")

	value := yaml.Node{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle}
	value.SetString(fmt.Sprintf("cyclic JSONRef to %s", refJSONPath))

	cyclicRef.Content = append(cyclicRef.Content, &key, &value)
	return cyclicRef
}

func LocateRef(refFileNode *yaml.Node, refJSONPath string, jsonRef string) (*yaml.Node, error) {
	var err error
	var yamlPath *yamlpath.Path

	yamlPath, err = yamlpath.NewPath(refJSONPath)
	if err != nil {
		return nil, errors.New(err)
	}

	var yamlNodes []*yaml.Node
	yamlNodes, err = yamlPath.Find(refFileNode)
	if err != nil {
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

func ResolveYAMLRefs(node *yaml.Node, filePath string, allowCycles bool) (*yaml.Node, error) {
	ResetYAMLRefs()
	node, _, err := ResolveYAMLRefsRecursive(node, node, "$", nil, filePath, nil, allowCycles, false)
	return node, err
}

func DetectCycle(node *yaml.Node, filePath string) (bool, error) {
	ResetYAMLRefs()
	_, cycleDetected, err := ResolveYAMLRefsRecursive(node, node, "$", nil, filePath, nil, false, true)
	return cycleDetected, err
}

func ResolveCycles(node *yaml.Node, filePath string) (*yaml.Node, error) {
	ResetYAMLRefs()
	node, _, err := ResolveYAMLRefsRecursive(node, node, "$", nil, filePath, nil, true, true)
	return node, err
}

func ResolveYAMLRefsRecursive(root *yaml.Node, node *yaml.Node,
	nodePath string, nodePaths []string,
	filePath string, filePaths []string,
	allowCycles bool,
	resolveCyclesOnly bool) (*yaml.Node, bool, error) {
	var err error

	if !filepath.IsAbs(filePath) {
		filePath, err = filepath.Abs(filePath)
		if err != nil {
			return nil, false, errors.New(err)
		}
	}

	if len(filePaths) == 0 {
		filePaths = append(filePaths, filePath)
	}

	if len(nodePaths) == 0 {
		nodePaths = append(nodePaths, nodePath)
	}

	if node == nil {
		return nil, false, nil
	}

	var resolvedNode *yaml.Node
	var cycleDetected bool

	if node.Kind == yaml.MappingNode && isYAMLRef(node) {
		if resolvedNode, cycleDetected, err = ResolveYAMLRef(root, node, nodePath, nodePaths, filePath, filePaths, allowCycles, resolveCyclesOnly); err != nil {
			return nil, cycleDetected, err
		}
		return resolvedNode, cycleDetected, nil
	} else if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			subPath := fmt.Sprintf("%s.%s", nodePath, node.Content[i].Value)
			subContent := node.Content[i+1]

			if slices.Index(nodePaths, subPath) >= 0 {
				return nil, true, errors.Errorf("cycle detected at %v", nodePaths)
			}

			newNodePaths := append([]string{}, nodePaths...)
			newNodePaths = append(newNodePaths, subPath)
			if resolvedNode, cycleDetected, err = ResolveYAMLRefsRecursive(root, subContent, subPath, newNodePaths, filePath, filePaths, allowCycles, resolveCyclesOnly); err != nil {
				return nil, cycleDetected, err
			}
			node.Content[i+1] = resolvedNode
		}
		return node, cycleDetected, nil
	} else if node.Kind == yaml.DocumentNode {
		subPath := nodePath
		subContent := node.Content[0]
		newNodePaths := nodePaths
		if resolvedNode, cycleDetected, err = ResolveYAMLRefsRecursive(root, subContent, subPath, newNodePaths, filePath, filePaths, allowCycles, resolveCyclesOnly); err != nil {
			return nil, cycleDetected, err
		}
		node.Content[0] = resolvedNode
		return node, cycleDetected, nil

	} else if node.Kind == yaml.SequenceNode {
		for i := 0; i < len(node.Content); i += 1 {
			subPath := fmt.Sprintf("%s.%v", nodePath, i)
			subContent := node.Content[i]

			if slices.Index(nodePaths, subPath) >= 0 {
				return nil, true, errors.Errorf("cycle detected at %v", nodePaths)
			}

			newNodePaths := append([]string{}, nodePaths...)
			newNodePaths = append(newNodePaths, subPath)
			if resolvedNode, cycleDetected, err = ResolveYAMLRefsRecursive(root, subContent, subPath, newNodePaths, filePath, filePaths, allowCycles, resolveCyclesOnly); err != nil {
				return nil, cycleDetected, err
			}
			node.Content[i] = resolvedNode
		}
		return node, cycleDetected, nil
	}

	return node, cycleDetected, nil
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

var ResolvedRefs map[string]*yaml.Node
var ParsedYAMLFiles map[string]*yaml.Node

func init() {
	ResetYAMLRefs()
}

func ResetYAMLRefs() {
	ParsedYAMLFiles = make(map[string]*yaml.Node)
	ResolvedRefs = make(map[string]*yaml.Node)
}

func YAMLFile2YAML(filePath string) (*yaml.Node, error) {
	var file *os.File
	var err error
	if file, err = os.Open(filePath); err != nil {
		return nil, errors.New(err)
	}
	defer func() { MustClose(file) }()

	//switch to directory relative to the YAML file so that JSON $refs are valid
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.New(err)
	}
	defer func() { Must(os.Chdir(wd)) }()

	err = os.Chdir(filepath.Dir(filePath))
	if err != nil {
		return nil, errors.New(err)
	}

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
		fallthrough
	case yaml.MappingNode:
		node.Style = 0
		for _, v := range node.Content {
			v.Style = 0
			UnFlowYAMLNode(v)
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
