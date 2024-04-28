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
	"bytes"
	"github.com/go-errors/errors"
	"gopkg.in/yaml.v3"
	"io"
)

func JSONText2YAMLText(reader io.Reader) ([]byte, error) {
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

	//convert back to YAML text
	return YAML2Text(UnFlowYAMLNode(yamlNode), 2)
}

func JSONFile2YAMLFile(input string, output string) error {
	text, err := ReadInputText(input)
	if err != nil {
		return err
	}

	outputText, err := JSONText2YAMLText(bytes.NewReader(text))
	if err != nil {
		return err
	}

	return WriteOutputText(output, outputText)
}
