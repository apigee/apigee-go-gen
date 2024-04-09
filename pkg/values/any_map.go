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

package values

import (
	"fmt"
	"github.com/go-errors/errors"
	"strconv"
	"strings"
)

type AnyMap struct {
	Data *Map
}

func NewValues(data *Map) AnyMap {
	return AnyMap{Data: data}
}

func (v *AnyMap) String() string {
	return fmt.Sprintf("%v", v.Data)
}

func (v *AnyMap) Set(entry string) error {
	key, value, found := strings.Cut(entry, "=")
	if !found {
		return errors.Errorf("missing value in set for key=%s", key)
	}

	parsedValue := ParseValue(value)
	v.Data.Set(key, parsedValue)

	return nil
}

func ParseValue(value string) any {
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return intValue
	}

	floatValue, err := strconv.ParseFloat(value, 10)
	if err != nil {
		return floatValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return boolValue
	}

	return value
}
