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

package flags

import (
	"fmt"
	"github.com/apigee/apigee-go-gen/pkg/parser"
	"github.com/apigee/apigee-go-gen/pkg/values"
	"github.com/go-errors/errors"
	"strings"
)

type SetGraphQL struct {
	Data *values.Map
}

func NewSetGraphQL(data *values.Map) SetGraphQL {
	return SetGraphQL{Data: data}
}

func (v *SetGraphQL) Type() string {
	return "string"
}

func (v *SetGraphQL) String() string {
	return ""
}

func (v *SetGraphQL) Set(entry string) error {
	key, filePath, found := strings.Cut(entry, "=")
	if !found {
		return errors.Errorf("missing file path in set-graphql for key=%s", key)
	}

	parserResult, schemaBytes, err := parser.ParseGraphQLSchema(filePath)
	if err != nil {
		return err
	}

	schema := *parserResult
	schemaFileText := string(schemaBytes)

	v.Data.Set(key, schema)
	v.Data.Set(fmt.Sprintf("%s_string", key), schemaFileText)

	return nil
}
