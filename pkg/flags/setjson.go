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
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/micovery/apigee-go-gen/pkg/values"
	"strings"
)

type SetJSON struct {
	Data *values.Map
}

func NewSetJSON(data *values.Map) SetJSON {
	return SetJSON{Data: data}
}

func (v *SetJSON) Type() string {
	return "string"
}

func (v *SetJSON) String() string {
	return ""
}

func (v *SetJSON) Set(entry string) error {
	key, jsonText, found := strings.Cut(entry, "=")
	if !found {
		return errors.Errorf("missing JSON text in set-json for key=%s", key)
	}

	wrappedJSONText := fmt.Sprintf(`{"JSON":%s}`, jsonText)
	type Wrapper struct {
		JSON any `yaml:"Data"`
	}

	wrapper := Wrapper{JSON: nil}
	err := json.Unmarshal([]byte(wrappedJSONText), &wrapper)
	if err != nil {
		return errors.New(err)
	}

	v.Data.Set(key, wrapper.JSON)
	return nil
}
