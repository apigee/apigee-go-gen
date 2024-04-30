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
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/apigee/apigee-go-gen/pkg/values"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestValues_Set(t *testing.T) {
	tests := []struct {
		name  string
		entry string
		want  any
	}{
		{
			"simple flat map",
			`
base_path: /v1/hello
target_url: https://example.com/
api_name: hello
use_proxy: false`,
			values.Map{
				"base_path":  "/v1/hello",
				"target_url": "https://example.com/",
				"api_name":   "hello",
				"use_proxy":  false,
			},
		},
		{
			"multiple objects",
			`
proxy_info:
 name: helloworld
 base_path: /v1/hello
target_info:
  url: https://example.com
  use_proxy: false`,
			values.Map{
				"proxy_info": values.Map{
					"name":      "helloworld",
					"base_path": "/v1/hello",
				},
				"target_info": values.Map{
					"url":       "https://example.com",
					"use_proxy": false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := values.Map{}
			v := NewValues(&data)
			valuesFile, err := os.CreateTemp("", "test-values-*.yaml")
			defer func() { utils.MustClose(valuesFile) }()
			require.NoError(t, err)

			err = os.WriteFile(valuesFile.Name(), []byte(tt.entry), os.ModePerm)
			require.NoError(t, err)

			err = v.Set(valuesFile.Name())
			require.NoError(t, err)
			assert.Equal(t, tt.want, data)
		})
	}
}
