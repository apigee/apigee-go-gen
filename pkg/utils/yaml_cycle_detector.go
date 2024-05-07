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
	"fmt"
	"github.com/go-errors/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func YAMLDetectRefCycles(root *yaml.Node, filePath string) (cycles [][]string, err error) {
	fileDir := filepath.Dir(filePath)
	filePath = filepath.Base(filePath)

	PopD := PushDir(fileDir)
	defer PopD()

	err = YAMLDetectRefCyclesRecursive(root, "", filePath, []string{}, &map[string]*yaml.Node{}, &cycles)

	return cycles, err
}

func YAMLDetectRefCyclesRecursive(node *yaml.Node, relParentPath string, parentFile string, activePaths []string, loaded *map[string]*yaml.Node, cycles *[][]string) error {
	var err error

	if node == nil {
		return nil
	}

	absFilePath, err := filepath.Abs(parentFile)
	if err != nil {
		return errors.Errorf("could not process %s:%s. %s", parentFile, relParentPath, err.Error())
	}

	activePath := fmt.Sprintf("%s:%s", absFilePath, relParentPath)
	if slices.Contains(activePaths, activePath) {
		rootPath, _, _ := strings.Cut(activePaths[0], ":")

		lastPath := activePaths[len(activePaths)-1]
		lastFilePath, lastNodePath, _ := strings.Cut(lastPath, ":")

		rel, _ := filepath.Rel(filepath.Dir(rootPath), lastFilePath)
		*cycles = append(*cycles, []string{rel, lastNodePath})
		return nil
	}

	activePaths = append(activePaths, activePath)

	if node.Kind == yaml.MappingNode && isYAMLRef(node) {
		jsonRef := getYAMLRefString(node)

		if jsonRef == "" {
			return errors.Errorf("JSONRef %s at %s is not valid", jsonRef, parentFile)
		}

		refFilePath, refJSONPath, err := SplitJSONRef(jsonRef)
		if err != nil {
			return errors.Errorf("could not process JSONRef %s at %s. %s", jsonRef, parentFile, err.Error())
		}

		if refFilePath == "" {
			refFilePath = parentFile
		}

		refFileNode, err := loadYAMLFile(refFilePath, loaded)

		if err != nil {
			return errors.Errorf("could not process JSONRef %s at %s. %s", jsonRef, parentFile, err.Error())
		}

		yamlNode, err := LocateRef(refFileNode, refJSONPath, jsonRef)
		if err != nil {
			return errors.Errorf("could not process JSONRef %s at %s, %s", jsonRef, parentFile, err.Error())
		}

		//switch-dir
		PopD := PushDir(filepath.Dir(refFilePath))
		defer PopD()

		err = YAMLDetectRefCyclesRecursive(yamlNode, refJSONPath, filepath.Base(refFilePath), activePaths, loaded, cycles)
		if err != nil {
			return err
		}

		return nil
	} else if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			curPath := fmt.Sprintf("%s.%s", relParentPath, node.Content[i].Value)

			err = YAMLDetectRefCyclesRecursive(node.Content[i+1], curPath, parentFile, activePaths, loaded, cycles)
			if err != nil {
				return err
			}
		}
	} else if node.Kind == yaml.DocumentNode {
		err = YAMLDetectRefCyclesRecursive(node.Content[0], "$", parentFile, activePaths, loaded, cycles)
		if err != nil {
			return err
		}
	} else if node.Kind == yaml.SequenceNode {
		for i := 0; i < len(node.Content); i += 1 {
			curPath := fmt.Sprintf("%s.%d", relParentPath, i)

			err = YAMLDetectRefCyclesRecursive(node.Content[i], curPath, parentFile, activePaths, loaded, cycles)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func loadYAMLFile(filePath string, loaded *map[string]*yaml.Node) (*yaml.Node, error) {
	var err error

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, errors.New(err)
	}

	//check if the file has already been parsed
	cachedNode, ok := (*loaded)[absFilePath]
	if ok {
		return cachedNode, nil
	}

	file, err := os.Open(absFilePath)
	if err != nil {
		return nil, errors.New(err)
	}

	decoder := yaml.NewDecoder(file)
	yamlNode := yaml.Node{}
	if err = decoder.Decode(&yamlNode); err != nil {
		return nil, errors.Errorf("could not load %s. %s", filePath, err.Error())
	}

	(*loaded)[absFilePath] = &yamlNode
	return &yamlNode, nil
}
