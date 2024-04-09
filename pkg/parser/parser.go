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

package parser

import (
	"bytes"
	"github.com/bufbuild/protocompile/ast"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/go-errors/errors"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	"github.com/vektah/gqlparser/v2"
	ast2 "github.com/vektah/gqlparser/v2/ast"
	"os"
)

func ParseOAS(specFile string) (libopenapi.Document, error) {
	var specBytes []byte
	var err error
	if specBytes, err = os.ReadFile(specFile); err != nil {
		return nil, errors.New(err)
	}

	config := datamodel.DocumentConfiguration{
		BasePath:            ".",
		AllowFileReferences: true,
	}

	var specDoc libopenapi.Document
	if specDoc, err = libopenapi.NewDocumentWithConfiguration(specBytes, &config); err != nil {
		return nil, errors.New(err)
	}

	return specDoc, nil
}

func ParseGRPCProto(protoFile string) (parser.Result, []byte, error) {

	var protoBytes []byte
	var err error
	if protoBytes, err = os.ReadFile(protoFile); err != nil {
		return nil, nil, errors.New(err)
	}
	errs := func(err reporter.ErrorWithPos) error {
		return err
	}
	warnings := func(pos reporter.ErrorWithPos) {
		return
	}
	newReporter := reporter.NewReporter(errs, warnings)
	handler := reporter.NewHandler(newReporter)
	var protoAst *ast.FileNode
	if protoAst, err = parser.Parse(protoFile, bytes.NewReader(protoBytes), handler); err != nil {
		return nil, nil, errors.New(err)
	}

	fromAST, err := parser.ResultFromAST(protoAst, false, handler)
	if err != nil {
		return nil, nil, errors.New(err)
	}

	return fromAST, protoBytes, nil

}

func ParseGraphQLSchema(schemaFile string) (*ast2.Schema, []byte, error) {

	var schemaBytes []byte
	var err error
	if schemaBytes, err = os.ReadFile(schemaFile); err != nil {
		return nil, nil, errors.New(err)
	}

	source := ast2.Source{
		Name:    "schema.graphql",
		Input:   string(schemaBytes),
		BuiltIn: false,
	}
	var schema *ast2.Schema

	if schema, err = gqlparser.LoadSchema(&source); err != nil {
		return nil, nil, errors.New(err)
	}

	return schema, schemaBytes, nil
}
