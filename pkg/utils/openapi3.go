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
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-errors/errors"
	libopenapijson "github.com/pb33f/libopenapi/json"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"slices"
)

func OAS3YAMLtoOAS2YAML(oasNode *yaml.Node) (*yaml.Node, error) {
	//convert it to JSON, since the converter library depends on JSON text
	jsonText, err := libopenapijson.YAMLNodeToJSON(oasNode, "  ")
	if err != nil {
		return nil, errors.New(err)
	}

	//then, convert it to the OAS2 data model
	var oas3doc openapi3.T
	err = json.Unmarshal(jsonText, &oas3doc)
	if err != nil {
		return nil, errors.New(err)
	}

	//finally, convert it to the OAS3 data model
	openapi3.CircularReferenceCounter = 5
	openapi3.DisableSchemaDefaultsValidation()
	openapi3.DisablePatternValidation()
	openapi3.DisableExamplesValidation()
	openapi3.DisableSchemaPatternValidation()
	openapi3.DisableSchemaFormatValidation()
	openapi3.DisableReadOnlyValidation()
	openapi3.DisableWriteOnlyValidation()

	//loader := openapi3.NewLoader()
	//loader.IsExternalRefsAllowed = true
	oas2doc, err := openapi2conv.FromV3(&oas3doc)
	if err != nil {
		return nil, errors.New(err)
	}

	//and back to YAML node
	return OAS2ToYAML(oas2doc)
}

func OAS3FileToOAS2File(input string, output string, allowCycles bool) error {
	text, err := ReadInputText(input)
	if err != nil {
		return err
	}

	//first, use the YAML library to parse it (regardless if it's JSON or YAML)
	var oas3node *yaml.Node
	oas3node = &yaml.Node{}
	err = yaml.Unmarshal(text, oas3node)
	if err != nil {
		return errors.New(err)
	}

	//verify we are actually working with OAS2
	if slices.IndexFunc(oas3node.Content[0].Content, func(n *yaml.Node) bool {
		return n.Value == "openapi"
	}) < 0 {
		return errors.Errorf("input %s is not an OpenAPI 3.0 spec", input)
	}

	//detect JSONRef cycles
	_, err = DetectCycle(oas3node, input)
	if err != nil {
		var cyclicError CyclicJSONRefError
		isCyclicError := errors.As(err, &cyclicError)

		if isCyclicError && allowCycles == true {
			oas3node, err = ResolveCycles(oas3node, input)
		} else {
			return err
		}
	}

	oas2node, err := RunWithinDirectory(filepath.Dir(input), func() (*yaml.Node, error) {
		return OAS3YAMLtoOAS2YAML(oas3node)
	})
	if err != nil {
		return err
	}

	ext := filepath.Ext(output)
	if ext == "" {
		ext = filepath.Ext(input)
	}

	//depending on the file extension write the output as either JSON or YAML
	var outputText []byte
	if ext == ".json" {
		outputText, err = libopenapijson.YAMLNodeToJSON(oas2node, "  ")
		if err != nil {
			return errors.New(err)
		}
	} else {
		outputText, err = YAML2Text(UnFlowYAMLNode(oas2node), 2)
		if err != nil {
			return err
		}
	}

	return WriteOutputText(output, outputText)
}

func OAS3ToYAML(doc *openapi3.T) (*yaml.Node, error) {
	var err error
	oas := &yaml.Node{Kind: yaml.MappingNode}

	//required
	_, err = AddEntryToOASYAML(oas, "openapi", doc.OpenAPI, nil)
	if err != nil {
		return nil, err
	}

	//required
	_, err = AddEntryToOASYAML(oas, "info", doc.Info, &yaml.Node{Kind: yaml.MappingNode})
	if err != nil {
		return nil, err
	}

	for k, v := range doc.Extensions {
		_, err = AddEntryToOASYAML(oas, k, v, &yaml.Node{Kind: yaml.MappingNode})
		if err != nil {
			return nil, err
		}
	}

	//optional
	if len(doc.Servers) > 0 {
		_, err = AddEntryToOASYAML(oas, "servers", doc.Servers, &yaml.Node{Kind: yaml.SequenceNode})
		if err != nil {
			return nil, err
		}
	}

	//optional
	if doc.ExternalDocs != nil {
		_, err = AddEntryToOASYAML(oas, "externalDocs", doc.ExternalDocs, &yaml.Node{Kind: yaml.MappingNode})
		if err != nil {
			return nil, err
		}
	}

	//optional
	if len(doc.Tags) > 0 {
		_, err = AddEntryToOASYAML(oas, "tags", doc.Tags, &yaml.Node{Kind: yaml.SequenceNode})
		if err != nil {
			return nil, err
		}
	}

	//optional
	if len(doc.Security) > 0 {
		_, err = AddEntryToOASYAML(oas, "security", doc.Security, &yaml.Node{Kind: yaml.SequenceNode})
		if err != nil {
			return nil, err
		}
	}

	//required
	_, err = AddEntryToOASYAML(oas, "paths", doc.Paths, &yaml.Node{Kind: yaml.MappingNode})
	if err != nil {
		return nil, err
	}

	//optional
	if doc.Components != nil {
		_, err = AddEntryToOASYAML(oas, "components", doc.Components, &yaml.Node{Kind: yaml.MappingNode})
		if err != nil {
			return nil, err
		}
	}

	return oas, nil
}
