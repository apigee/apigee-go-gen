//  Copyright 2025 Google LLC
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
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/apigee/apigee-go-gen/pkg/values"
	"github.com/go-errors/errors"
	"os"
	"strings"
)

type SetTF struct {
	Data *values.Map
}

func NewSetTF(data *values.Map) SetTF {
	return SetTF{Data: data}
}

func (v *SetTF) Type() string {
	return "string"
}

func (v *SetTF) String() string {
	return ""
}

func (v *SetTF) Set(entry string) error {
	key, filePath, found := strings.Cut(entry, "=")
	if !found {
		return errors.Errorf("missing file path in set-tf for key=%s", key)
	}

	tfFileText, err := os.ReadFile(filePath)
	if err != nil {
		return errors.New(err)
	}

	tfFileMap := make(map[string]any)
	if tfFileMap, err = utils.TFTextToMap(tfFileText, filePath); err != nil {
		return err
	}

	v.Data.Set(key, tfFileMap)
	v.Data.Set(fmt.Sprintf("%s_string", key), string(tfFileText))
	v.Data.Set(fmt.Sprintf("%s_file", key), filePath)

	return nil
}
