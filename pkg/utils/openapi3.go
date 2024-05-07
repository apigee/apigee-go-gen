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
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

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
