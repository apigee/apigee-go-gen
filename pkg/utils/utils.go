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
	"encoding/json"
	"github.com/go-errors/errors"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
)

func MustReadFileBytes(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return data
}

func AddEntryToOASYAML(oas *yaml.Node, key string, value any, defaultVal *yaml.Node) (*yaml.Node, error) {
	oas.Content = append(oas.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key})

	//the kin-openapi library only uses JSON tags, so we must marshall to JSON first
	jsonText, err := json.Marshal(value)
	if err != nil {
		return nil, errors.New(err)
	}

	yamlText, err := JSONText2YAMLText(bytes.NewReader(jsonText))
	if err != nil {
		return nil, err
	}

	yamlNode := yaml.Node{}
	err = yaml.Unmarshal(yamlText, &yamlNode)
	if err != nil {
		return nil, errors.New(err)
	}

	content := yamlNode.Content[0]
	if content.Kind == yaml.ScalarNode && content.Value == "null" && defaultVal != nil {
		content = defaultVal
	}
	oas.Content = append(oas.Content, content)
	return &yamlNode, nil
}

type MultiError struct {
	Errors []error
}

func (e MultiError) Error() string {
	return errors.Join(e.Errors...).Error()
}

func PushDir(dir string) func() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	popDir := func() {
		Must(os.Chdir(wd))
	}

	return popDir
}

func RemoveYAMLComments(data []byte) []byte {
	regex := regexp.MustCompile(`(?ms)^\s*#[^\n\r]*$[\r\n]*`)
	replaced := regex.ReplaceAll(data, []byte{})
	return replaced
}
