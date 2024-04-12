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

package utils

import (
	"bytes"
	"encoding/xml"
	"github.com/beevik/etree"
	"github.com/go-errors/errors"
	"gopkg.in/yaml.v3"
	"io"
	"slices"
	"strings"
)

func XMLText2YAMLText(reader io.Reader) ([]byte, error) {
	var err error
	var doc *etree.Document
	if doc, err = Text2XML(reader); err != nil {
		PrintErrorWithStackAndExit(err)
		return nil, err
	}

	var yamlText []byte
	if yamlText, err = XML2YAMLText(doc, 2); err != nil {
		PrintErrorWithStackAndExit(err)
		return nil, err
	}
	return yamlText, nil
}

func XML2YAML(doc *etree.Document) (*yaml.Node, error) {
	var err error
	var res *yaml.Node
	if _, res, err = XML2YAMLRecursive(&doc.Element); err != nil {
		return nil, err
	}

	return res, nil
}

func XML2Text(doc *etree.Document) ([]byte, error) {
	var err error
	var bytes []byte
	doc.Indent(2)
	doc.WriteSettings = etree.WriteSettings{
		CanonicalEndTags: false,
		CanonicalText:    true,
		CanonicalAttrVal: false,
		AttrSingleQuote:  false,
		UseCRLF:          false,
	}
	if bytes, err = doc.WriteToBytes(); err != nil {
		return nil, errors.New(err)
	}

	return bytes, nil
}

func XMLTextFormat(reader io.Reader) ([]byte, error) {

	xmlDoc, err := Text2XML(reader)
	if err != nil {
		return nil, err
	}

	return XML2Text(xmlDoc)

}
func Text2XML(reader io.Reader) (*etree.Document, error) {
	var err error
	doc := etree.NewDocument()
	if _, err = doc.ReadFrom(reader); err != nil {
		return nil, errors.New(err)
	}
	return doc, nil
}

func XML2YAMLText(doc *etree.Document, indent int) ([]byte, error) {
	var err error
	var yamlNode *yaml.Node
	if yamlNode, err = XML2YAML(doc); err != nil {
		return nil, err
	}

	return YAML2Text(yamlNode, indent)
}

func XMLText2YAML(reader io.Reader) (*yaml.Node, error) {
	var err error
	doc := etree.NewDocument()

	if _, err = doc.ReadFrom(reader); err != nil {
		return nil, errors.New(err)
	}

	return XML2YAML(doc)
}

func XMLText2XML(reader io.Reader) (*etree.Document, error) {
	var err error
	doc := etree.NewDocument()

	if _, err = doc.ReadFrom(reader); err != nil {
		return nil, errors.New(err)
	}

	return doc, nil
}

func Struct2XMLDocText(p any) ([]byte, error) {
	var err error
	outXML := bytes.Buffer{}
	encoder := xml.NewEncoder(&outXML)
	encoder.Indent("", " ")
	if err = encoder.Encode(p); err != nil {
		return nil, errors.New(err)
	}

	fmtXML, err := XMLTextFormat(bytes.NewReader(outXML.Bytes()))
	if err != nil {
		return nil, err
	}

	w := &bytes.Buffer{}
	w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"))
	w.Write(fmtXML)

	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func XML2YAMLRecursive(ele *etree.Element) (key *yaml.Node, value *yaml.Node, err error) {
	if ele == nil {
		return nil, nil, nil
	}

	nodeKey := &yaml.Node{Kind: yaml.ScalarNode, Value: ele.Tag}
	nodeVal := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{}}

	for i := 0; i < len(ele.Attr); i++ {
		attr := ele.Attr[i]
		nodeVal.Content = append(nodeVal.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: "." + attr.Key}, &yaml.Node{Kind: yaml.ScalarNode, Value: attr.Value})
	}

	getCharElement := func(ele *etree.Element) (*etree.CharData, bool) {
		var charData *etree.CharData
		var ok bool

		data := ""
		for i := 0; i < len(ele.Child); i++ {
			if charData, ok = (ele.Child[i]).(*etree.CharData); !ok {
				return nil, false
			}
			data += charData.Data
		}

		return &etree.CharData{Data: strings.TrimSpace(data)}, false
	}

	var children *yaml.Node
	charEle, _ := getCharElement(ele)
	const cData = "-Data"
	if charEle != nil && len(ele.Attr) > 0 {
		if len(charEle.Data) > 0 {
			nodeVal.Content = append(nodeVal.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: cData},
				&yaml.Node{Kind: yaml.ScalarNode, Value: charEle.Data})
		}
		return nodeKey, nodeVal, nil
	} else if charEle != nil && len(ele.Attr) == 0 {
		nodeKey = &yaml.Node{Kind: yaml.ScalarNode, Value: ele.Tag}
		if len(charEle.Data) > 0 {
			nodeVal = &yaml.Node{Kind: yaml.ScalarNode, Value: charEle.Data}
		} else {
			nodeVal = &yaml.Node{Kind: yaml.MappingNode}
		}
		return nodeKey, nodeVal, nil
	} else if len(ele.Attr) > 0 {
		children = &yaml.Node{Kind: yaml.MappingNode}
		children.Content = []*yaml.Node{}
		nodeVal.Content = append(nodeVal.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: cData},
			children)
	} else {
		children = nodeVal
	}

	uniqueKeysLookup := map[string]bool{}
	allUnique := true
	for i := 0; i < len(ele.Child); i++ {
		ele := ele.Child[i]
		switch v := (ele).(type) {
		case *etree.Element:
			var childKey *yaml.Node
			var childValue *yaml.Node
			if childKey, childValue, err = XML2YAMLRecursive(v); err != nil {
				return nil, nil, err
			}

			if childKey == nil || childValue == nil {
				continue
			}

			children.Content = append(children.Content, childKey, childValue)
			if _, ok := uniqueKeysLookup[childKey.Value]; !ok {
				uniqueKeysLookup[childKey.Value] = true
			} else {
				uniqueKeysLookup[childKey.Value] = false
				allUnique = false
			}
		}
	}

	if !allUnique {
		//it's a sequence
		sequence := &yaml.Node{Kind: yaml.SequenceNode}
		sequence.Content = []*yaml.Node{}

		for i := 0; i < len(children.Content); i += 2 {
			childKey := children.Content[i]
			childValue := children.Content[i+1]
			newChild := &yaml.Node{Kind: yaml.MappingNode}
			newChild.Content = append([]*yaml.Node{}, childKey, childValue)
			sequence.Content = append(sequence.Content, newChild)
		}

		if nodeVal == children {
			// parent has no attributes
			nodeVal = sequence
		} else {
			// parent has
			children.Kind = sequence.Kind
			children.Content = sequence.Content
		}

		return nodeKey, nodeVal, nil
	}

	childrenIndex := slices.IndexFunc(nodeVal.Content, func(e *yaml.Node) bool {
		return e.Value == cData
	})

	if childrenIndex > 0 && children != nodeVal &&
		children.Kind == yaml.MappingNode &&
		nodeVal.Kind == yaml.MappingNode &&
		allUnique {
		//remove unnecessary nesting
		newContent := append([]*yaml.Node{}, nodeVal.Content[0:childrenIndex]...)
		newContent = append(newContent, nodeVal.Content[childrenIndex+2:]...)
		newContent = append(newContent, children.Content...)
		nodeVal.Content = newContent
		return nodeKey, nodeVal, nil
	}

	return nodeKey, nodeVal, nil

}
