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

package values

import (
	"fmt"
	"github.com/go-errors/errors"
	"gopkg.in/yaml.v3"
	"os"
)

type FileMap struct {
	Data *Map
}

func NewValuesFile(data *Map) FileMap {
	return FileMap{Data: data}
}

func (v *FileMap) String() string {
	return fmt.Sprintf("%v", v.Data)
}

func (v *FileMap) Set(filePath string) error {
	yamlText, err := os.ReadFile(filePath)
	if err != nil {
		return errors.New(err)
	}

	err = yaml.Unmarshal(yamlText, v.Data)
	if err != nil {
		return errors.New(err)
	}

	return nil
}
