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

package render

import (
	"bytes"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/micovery/apigee-go-gen/pkg/apigee/v1"
	"github.com/micovery/apigee-go-gen/pkg/flags"
	"github.com/micovery/apigee-go-gen/pkg/utils"
	"os"
	"path/filepath"
)

func GenerateBundle(createModelFunc func(string) (v1.Model, error), cFlags *CommonFlags, validate bool, dryRun string) error {
	var err error

	bundleOutputFile := cFlags.OutputFile

	//create a temporary location for rendering into
	tmpDir, err := os.MkdirTemp("", "render-*")
	if err != nil {
		return errors.New(err)
	}

	// render the template to a temporary location
	cFlags.OutputFile = flags.String(filepath.Join(tmpDir, "model.yaml"))
	err = RenderGenericTemplate(cFlags, false)
	if err != nil {
		return errors.New(err)
	}

	templateFile := string(cFlags.TemplateFile)

	//resolve YAML Refs immediately after rendering to avoid having copy files in $ref
	rendered, err := os.ReadFile(string(cFlags.OutputFile))
	if err != nil {
		return errors.New(err)
	}

	rendered, err = ResolveYAML(rendered, templateFile)
	if err != nil {
		return err
	}

	err = os.WriteFile(string(cFlags.OutputFile), rendered, os.ModePerm)
	if err != nil {
		return errors.New(err)
	}

	// create apiproxy from rendered template
	model, err := createModelFunc(string(cFlags.OutputFile))
	if err != nil {
		return err
	}
	return CreateBundle(model, string(bundleOutputFile), validate, dryRun)
}

func CreateBundle(model v1.Model, output string, validate bool, dryRun string) (err error) {
	if err != nil {
		return err
	}

	if dryRun == "xml" {
		xmlText, err := model.XML()
		if err != nil {
			return err
		}
		fmt.Println(string(xmlText))
	} else if dryRun == "yaml" {
		yamlText, err := model.YAML()
		if err != nil {
			return err
		}
		fmt.Println(string(yamlText))
	}

	if validate {
		err = model.Validate()
		if err != nil {
			return err
		}
	}

	if dryRun != "" {
		return nil
	}

	err = v1.Model2Bundle(model, output)
	if err != nil {
		return err
	}

	return nil
}

func ResolveYAML(text []byte, filePath string) ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.New(err)
	}
	defer func() { utils.Must(os.Chdir(wd)) }()

	err = os.Chdir(filepath.Dir(filePath))
	if err != nil {
		return nil, errors.New(err)
	}

	yaml, err := utils.Text2YAML(bytes.NewReader(text))
	if err != nil {
		return nil, err
	}
	newText, err := utils.YAML2Text(yaml, 2)
	if err != nil {
		return nil, err
	}

	return newText, nil
}
