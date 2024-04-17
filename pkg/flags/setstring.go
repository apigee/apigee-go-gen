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
	"github.com/go-errors/errors"
	"github.com/micovery/apigee-go-gen/pkg/values"
	"strings"
)

type SetString struct {
	Data *values.Map
}

func NewSetString(data *values.Map) SetString {
	return SetString{Data: data}
}

func (v *SetString) Type() string {
	return "string"
}

func (v *SetString) String() string {
	return ""
}

func (v *SetString) Set(entry string) error {
	key, value, found := strings.Cut(entry, "=")
	if !found {
		return errors.Errorf("missing value in set for key=%s", key)
	}

	v.Data.Set(key, value)
	return nil
}
