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

package render

import (
	"bytes"
	"fmt"
	"github.com/apigee/apigee-go-gen/pkg/apigee/v1"
	"github.com/apigee/apigee-go-gen/pkg/flags"
	"github.com/apigee/apigee-go-gen/pkg/git"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/go-errors/errors"
	"os"
	"path/filepath"
)

func GenerateBundle(createModelFunc func(string) (v1.Model, error), cFlags *CommonFlags, validate bool, dryRun string, debug bool) error {
	var err error

	bundleOutputFile := cFlags.OutputFile

	if git.IsGitURI(string(cFlags.TemplateFile)) {
		var templateFileFromGit string
		var templateDirFromGit string
		if templateFileFromGit, templateDirFromGit, err = git.FetchFile(string(cFlags.TemplateFile)); err != nil {
			return err
		}
		defer utils.MustRemoveAll(templateDirFromGit)
		cFlags.TemplateFile = flags.String(templateFileFromGit)
		fileRelative, _ := filepath.Rel(templateDirFromGit, templateFileFromGit)
		cFlags.TemplateFileAlias = flags.String(fileRelative)
	}

	templateDir := filepath.Dir(string(cFlags.TemplateFile))

	var tmpDir string
	//copy the template directory into temporary location for rendering into
	if tmpDir, err = os.MkdirTemp("", "render-*"); err != nil {
		return errors.New(err)
	}
	defer utils.MustRemoveAll(tmpDir)

	if err = utils.CopyDir(tmpDir, templateDir); err != nil {
		return errors.New(err)
	}

	var tempRenderedFile *os.File
	if tempRenderedFile, err = os.CreateTemp(tmpDir, fmt.Sprintf("rendered-*-template.yaml")); err != nil {
		return errors.New(err)
	}

	// render the template to a temporary location
	cFlags.OutputFile = flags.String(tempRenderedFile.Name())
	if err = RenderGenericTemplateLocal(cFlags, false); err != nil {
		return err
	}

	rendered, err := os.ReadFile(tempRenderedFile.Name())
	if err != nil {
		return errors.New(err)
	}

	if debug {
		fmt.Println(string(rendered))
		return nil
	}

	rendered, err = ResolveYAML(rendered, tempRenderedFile.Name())
	if err != nil {
		return utils.MultiError{
			Errors: []error{
				err,
				errors.New("rendered template appears to not be valid YAML. Use --debug=true flag to inspect rendered output")}}
	}

	err = os.WriteFile(tempRenderedFile.Name(), rendered, os.ModePerm)
	if err != nil {
		return errors.New(err)
	}

	// create apiproxy from rendered template
	model, err := createModelFunc(tempRenderedFile.Name())
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
	popd := utils.PushDir(filepath.Dir(filePath))
	defer popd()

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
