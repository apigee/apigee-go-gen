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
	"github.com/micovery/apigee-yaml-toolkit/pkg/values"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetString_Set(t *testing.T) {
	tests := []struct {
		entry string
		want  any
	}{
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"37",
			"37",
		},
		{
			"13.37",
			"13.37",
		},
		{
			"1hello2",
			"1hello2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.entry, func(t *testing.T) {
			data := values.Map{}
			v := NewSetString(&data)
			err := v.Set(fmt.Sprintf("field=%s", tt.entry))
			require.NoError(t, err)
			assert.Equal(t, tt.want, data["field"])
		})
	}
}
