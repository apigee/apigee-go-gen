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
	"fmt"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-errors/errors"
	libopenapijson "github.com/pb33f/libopenapi/json"
	"gopkg.in/yaml.v3"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
)

func OAS2YAMLtoOAS3YAML(oasNode *yaml.Node) (*yaml.Node, error) {
	//convert it to JSON, since the converter library depends on JSON text
	jsonText, err := libopenapijson.YAMLNodeToJSON(oasNode, "  ")
	if err != nil {
		return nil, errors.New(err)
	}

	//then, convert it to the OAS2 data model
	var oas2doc openapi2.T
	err = json.Unmarshal(jsonText, &oas2doc)
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

	oas3doc, err := ToV3(&oas2doc)
	if err != nil {
		return nil, errors.New(err)
	}

	//and back to YAML node
	return OAS3ToYAML(oas3doc)
}

func OAS2FileToOAS3File(input string, output string, allowCycles bool) error {
	text, err := ReadInputText(input)
	if err != nil {
		return err
	}

	//first, use the YAML library to parse it (regardless if it's JSON or YAML)
	var oas2node *yaml.Node
	oas2node = &yaml.Node{}
	err = yaml.Unmarshal(text, oas2node)
	if err != nil {
		return errors.New(err)
	}

	//verify we are actually working with OAS2
	if slices.IndexFunc(oas2node.Content[0].Content, func(n *yaml.Node) bool {
		return n.Value == "swagger"
	}) < 0 {
		return errors.Errorf("input %s is not an OpenAPI 2.0 spec", input)
	}

	cycles, err := YAMLDetectRefCycles(oas2node, input)
	if err != nil {
		return err
	}

	if len(cycles) > 0 && allowCycles == false {
		var multiError MultiError
		for _, cycle := range cycles {
			multiError.Errors = append(multiError.Errors, errors.Errorf("cyclic ref at %s", strings.Join(cycle, ":")))
		}
		return errors.New(multiError)
	}

	oas3node, err := OAS2YAMLtoOAS3YAML(oas2node)
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
		outputText, err = libopenapijson.YAMLNodeToJSON(oas3node, "  ")
		if err != nil {
			return errors.New(err)
		}
	} else {
		outputText, err = YAML2Text(UnFlowYAMLNode(oas3node), 2)
		if err != nil {
			return err
		}
	}

	return WriteOutputText(output, outputText)
}

func ToV3(doc2 *openapi2.T) (*openapi3.T, error) {
	doc3 := &openapi3.T{
		OpenAPI:      "3.0.3",
		Info:         &doc2.Info,
		Components:   &openapi3.Components{},
		Tags:         doc2.Tags,
		Extensions:   doc2.Extensions,
		ExternalDocs: doc2.ExternalDocs,
	}

	if host := doc2.Host; host != "" {
		if strings.Contains(host, "://") {
			err := fmt.Errorf("%s host is not valid", host)
			return nil, err
		}
		schemes := doc2.Schemes
		if len(schemes) == 0 {
			schemes = []string{"https"}
		}
		basePath := doc2.BasePath
		if basePath == "" {
			basePath = "/"
		}
		for _, scheme := range schemes {
			u := url.URL{
				Scheme: scheme,
				Host:   host,
				Path:   basePath,
			}
			doc3.AddServer(&openapi3.Server{URL: u.String()})
		}
	}

	doc3.Components.Schemas = make(map[string]*openapi3.SchemaRef)
	if parameters := doc2.Parameters; len(parameters) != 0 {
		doc3.Components.Parameters = make(map[string]*openapi3.ParameterRef)
		doc3.Components.RequestBodies = make(map[string]*openapi3.RequestBodyRef)
		for k, parameter := range parameters {
			v3Parameter, v3RequestBody, v3SchemaMap, err := openapi2conv.ToV3Parameter(doc3.Components, parameter, doc2.Consumes)
			switch {
			case err != nil:
				return nil, err
			case v3RequestBody != nil:
				doc3.Components.RequestBodies[k] = v3RequestBody
			case v3SchemaMap != nil:
				for _, v3Schema := range v3SchemaMap {
					doc3.Components.Schemas[k] = v3Schema
				}
			default:
				doc3.Components.Parameters[k] = v3Parameter
			}
		}
	}

	if paths := doc2.Paths; len(paths) != 0 {
		doc3.Paths = openapi3.NewPathsWithCapacity(len(paths))
		for path, pathItem := range paths {
			r, err := openapi2conv.ToV3PathItem(doc2, doc3.Components, pathItem, doc2.Consumes)
			if err != nil {
				return nil, err
			}
			doc3.Paths.Set(path, r)
		}
	}

	if responses := doc2.Responses; len(responses) != 0 {
		doc3.Components.Responses = make(openapi3.ResponseBodies, len(responses))
		for k, response := range responses {
			r, err := openapi2conv.ToV3Response(response, doc2.Produces)
			if err != nil {
				return nil, err
			}
			doc3.Components.Responses[k] = r
		}
	}

	for key, schema := range openapi2conv.ToV3Schemas(doc2.Definitions) {
		doc3.Components.Schemas[key] = schema
	}

	if m := doc2.SecurityDefinitions; len(m) != 0 {
		doc3SecuritySchemes := make(map[string]*openapi3.SecuritySchemeRef)
		for k, v := range m {
			r, err := openapi2conv.ToV3SecurityScheme(v)
			if err != nil {
				return nil, err
			}
			doc3SecuritySchemes[k] = r
		}
		doc3.Components.SecuritySchemes = doc3SecuritySchemes
	}

	doc3.Security = openapi2conv.ToV3SecurityRequirements(doc2.Security)
	return doc3, nil
}
