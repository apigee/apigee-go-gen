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
	"github.com/go-errors/errors"
	"github.com/micovery/apigee-yaml-toolkit/pkg/parser"
	"github.com/micovery/apigee-yaml-toolkit/pkg/values"
	"strings"
)

type SetGRPC struct {
	Data *values.Map
}

func NewSetGRPC(data *values.Map) SetGRPC {
	return SetGRPC{Data: data}
}

func (v *SetGRPC) Type() string {
	return "string"
}

func (v *SetGRPC) String() string {
	return ""
}

func (v *SetGRPC) Set(entry string) error {
	key, filePath, found := strings.Cut(entry, "=")
	if !found {
		return errors.Errorf("missing file path in set-grpc for key=%s", key)
	}

	parserResult, protoBytes, err := parser.ParseGRPCProto(filePath)
	if err != nil {
		return err
	}

	proto := *parserResult.FileDescriptorProto()
	protoFileText := string(protoBytes)

	v.Data.Set(key, proto)
	v.Data.Set(fmt.Sprintf("%s_string", key), protoFileText)

	return nil
}
