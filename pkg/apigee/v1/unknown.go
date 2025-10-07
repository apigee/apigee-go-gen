// Copyright 2025 Google LLC
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
	"runtime/debug" // Import the debug package
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
	Stack    []byte // Add a field to hold the stack trace
}

func (e *UnknownNodeError) Error() string {
	return fmt.Sprintf(`unknown node "%s" found at "%s"`, e.Node.XMLName.Local, e.Location)
}

func (e *UnknownNodeError) String() string {
	return fmt.Sprintf("%s\n%s", e.Error(), e.Stack)
}

func NewUnknownNodeError(location string, node *AnyNode) *UnknownNodeError {
	return &UnknownNodeError{
		Location: location,
		Node:     node,
		Stack:    debug.Stack(),
	}
}
