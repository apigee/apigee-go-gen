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
)

func ResolveDollarRefs(input string, output string, allowCycles bool) error {
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

	//resolve references
	yamlNode, err = ResolveYAMLRefs(yamlNode, input, allowCycles)
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
