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

package mock

import (
	"fmt"
	v1 "github.com/apigee/apigee-go-gen/pkg/apigee/v1"
	"github.com/apigee/apigee-go-gen/pkg/common/mock_apiproxy_template"
	"github.com/apigee/apigee-go-gen/pkg/flags"
	"github.com/apigee/apigee-go-gen/pkg/render"
	"github.com/go-errors/errors"
	"os"
	"path/filepath"
)

func GenerateMockProxyBundle(input string, output string, cFlags *render.CommonFlags, debug bool) error {
	var templateDir string
	var err error

	if templateDir, err = getMockProxyTemplateDir(); err != nil {
		return errors.New(err)
	}
	defer func() { _ = os.RemoveAll(templateDir) }()

	createModelFunc := func(input string) (v1.Model, error) {
		return v1.NewAPIProxyModel(input)
	}

	cFlags.OutputFile = flags.NewString(output)
	cFlags.TemplateFile = flags.NewString(filepath.Join(templateDir, "apiproxy.yaml"))
	cFlags.IncludeList = flags.NewIncludeList([]string{filepath.Join(templateDir, "*.tmpl")})

	var oas = flags.NewSetOAS(cFlags.Values)
	if err = oas.Set(fmt.Sprintf("spec=%s", input)); err != nil {
		return errors.New(err)
	}

	return render.GenerateBundle(createModelFunc, cFlags, true, "", debug)
}

func getMockProxyTemplateDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "mock_apiproxy_*")
	if err != nil {
		return "", errors.New(err)
	}

	dir, err := mock_apiproxy_template.FS.ReadDir(".")
	if err != nil {
		return "", errors.New(err)
	}

	for _, dirEntry := range dir {
		var fileBytes []byte
		if fileBytes, err = mock_apiproxy_template.FS.ReadFile(dirEntry.Name()); err != nil {
			return "", errors.New(err)
		}

		if err = os.WriteFile(filepath.Join(tmpDir, dirEntry.Name()), fileBytes, os.ModePerm); err != nil {
			return "", errors.New(err)
		}
	}

	return tmpDir, nil
}
