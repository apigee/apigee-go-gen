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
	"github.com/apigee/apigee-go-gen/pkg/values"
	"github.com/go-errors/errors"
	"os"
	"strings"
)

type SetFile struct {
	Data *values.Map
}

func NewSetFile(data *values.Map) SetFile {
	return SetFile{Data: data}
}

func (v *SetFile) Type() string {
	return "string"
}

func (v *SetFile) String() string {
	return ""
}

func (v *SetFile) Set(entry string) error {
	key, filePath, found := strings.Cut(entry, "=")
	if !found {
		return errors.Errorf("missing file path in set-file for key=%s", key)
	}

	fileText, err := os.ReadFile(filePath)
	if err != nil {
		return errors.New(err)
	}

	v.Data.Set(key, string(fileText))
	return nil
}
