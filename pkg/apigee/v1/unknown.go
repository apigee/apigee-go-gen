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

package v1

import (
	"encoding/xml"
	"fmt"
)

type Deprecated AnyNode

type AnyList []*AnyNode

type AnyNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	CharData []byte     `xml:",chardata"`
	Children []AnyNode  `xml:",any"`
}

type UnknownNodeError struct {
	Location string
	Node     *AnyNode
}

func (e *UnknownNodeError) Error() string {
	return fmt.Sprintf(`unknown node "%s" found at "%s"`, e.Node.XMLName.Local, e.Location)
}

type ValidationErrors struct {
	Errors []error
}

func (e ValidationErrors) Error() string {
	return e.Errors[0].Error()
}
