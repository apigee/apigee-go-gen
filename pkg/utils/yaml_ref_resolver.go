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
	"net/url"
	"path/filepath"
	"slices"
	"strings"
)

func YAMLResolveAllRefs(root *yaml.Node, filePath string, allowCycles bool) (*yaml.Node, error) {
	fileDir := filepath.Dir(filePath)
	filePath = filepath.Base(filePath)

	PopD := PushDir(fileDir)
	defer PopD()

	cycles := [][]string{}
	resolved, err := YAMLResolveRefsRecursive(root, "", filePath, []string{}, &map[string]*yaml.Node{}, &cycles, true)
	if err != nil {
		return nil, err
	}

	if len(cycles) > 0 && allowCycles == false {
		var multiError MultiError
		for _, cycle := range cycles {
			multiError.Errors = append(multiError.Errors, errors.Errorf("cyclic ref at %s", strings.Join(cycle, ":")))
		}
		return nil, errors.New(multiError)
	}

	return resolved, nil
}

func YAMLResolveRefs(root *yaml.Node, filePath string, allowCycles bool) (*yaml.Node, error) {
	fileDir := filepath.Dir(filePath)
	filePath = filepath.Base(filePath)

	PopD := PushDir(fileDir)
	defer PopD()

	cycles := [][]string{}
	resolved, err := YAMLResolveRefsRecursive(root, "", filePath, []string{}, &map[string]*yaml.Node{}, &cycles, false)
	if err != nil {
		return nil, err
	}

	if len(cycles) > 0 && allowCycles == false {
		var multiError MultiError
		for _, cycle := range cycles {
			multiError.Errors = append(multiError.Errors, errors.Errorf("cyclic ref at %s", strings.Join(cycle, ":")))
		}
		return nil, errors.New(multiError)
	}

	return resolved, nil
}

func YAMLResolveRefsRecursive(node *yaml.Node, relParentPath string, parentFile string, activePaths []string, loaded *map[string]*yaml.Node, cycles *[][]string, resolveMainRefs bool) (*yaml.Node, error) {
	var err error

	if node == nil {
		return nil, nil
	}

	absFilePath, err := filepath.Abs(parentFile)
	if err != nil {
		return nil, errors.Errorf("could not process %s:%s. %s", parentFile, relParentPath, err.Error())
	}

	activePath := fmt.Sprintf("%s:%s", absFilePath, relParentPath)
	if slices.Contains(activePaths, activePath) {
		rootPath, _, _ := strings.Cut(activePaths[0], ":")

		lastPath := activePaths[len(activePaths)-1]
		lastFilePath, lastNodePath, _ := strings.Cut(lastPath, ":")

		rel, _ := filepath.Rel(filepath.Dir(rootPath), lastFilePath)
		*cycles = append(*cycles, []string{rel, lastNodePath})

		return MakeCyclicRefPlaceholder(relParentPath), nil
	}

	activePaths = append(activePaths, activePath)

	if node.Kind == yaml.MappingNode && isYAMLRef(node) {
		jsonRef := getYAMLRefString(node)

		if jsonRef == "" {
			return nil, errors.Errorf("JSONRef %s at %s is not valid", jsonRef, parentFile)
		}

		refFilePath, refJSONPath, err := SplitJSONRef(jsonRef)
		if err != nil {
			return nil, errors.Errorf("could not process JSONRef %s at %s. %s", jsonRef, parentFile, err.Error())
		}

		if refFilePath == "" {
			refFilePath = parentFile
		}

		//do not resolve refs that point back to the main file back
		absRefFilePath, _ := filepath.Abs(refFilePath)
		rootPath, _, _ := strings.Cut(activePaths[0], ":")
		if absRefFilePath == rootPath && resolveMainRefs == false {
			return node, nil
		}

		refFileNode, err := loadYAMLFile(refFilePath, loaded)

		if err != nil {
			return nil, errors.Errorf("could not process JSONRef %s at %s. %s", jsonRef, parentFile, err.Error())
		}

		yamlNode, err := LocateRef(refFileNode, refJSONPath, jsonRef)
		if err != nil {
			return nil, errors.Errorf("could not process JSONRef %s at %s, %s", jsonRef, parentFile, err.Error())
		}

		//switch-dir
		popd := PushDir(filepath.Dir(refFilePath))
		defer popd()

		resolved, err := YAMLResolveRefsRecursive(yamlNode, refJSONPath, filepath.Base(refFilePath), activePaths, loaded, cycles, resolveMainRefs)
		if err != nil {
			return nil, err
		}
		return resolved, nil

	} else if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			curPath := fmt.Sprintf("%s.%s", relParentPath, node.Content[i].Value)

			resolved, err := YAMLResolveRefsRecursive(node.Content[i+1], curPath, parentFile, activePaths, loaded, cycles, resolveMainRefs)
			if err != nil {
				return nil, err
			}
			node.Content[i+1] = resolved
		}
		return node, nil
	} else if node.Kind == yaml.DocumentNode {
		resolved, err := YAMLResolveRefsRecursive(node.Content[0], "$", parentFile, activePaths, loaded, cycles, resolveMainRefs)
		if err != nil {
			return nil, err
		}
		node.Content[0] = resolved
		return node, nil
	} else if node.Kind == yaml.SequenceNode {
		for i := 0; i < len(node.Content); i += 1 {
			curPath := fmt.Sprintf("%s.%d", relParentPath, i)

			resolved, err := YAMLResolveRefsRecursive(node.Content[i], curPath, parentFile, activePaths, loaded, cycles, resolveMainRefs)
			if err != nil {
				return nil, err
			}
			node.Content[i] = resolved
		}
		return node, nil
	}

	return node, nil
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

func setYAMLRefString(node *yaml.Node, value string) {
	index := slices.IndexFunc(node.Content, func(n *yaml.Node) bool {
		return n.Value == "$ref"
	})
	if index < 0 {
		return
	}
	node.Content[index+1].Value = value
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
