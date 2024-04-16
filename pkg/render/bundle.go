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
	"github.com/go-errors/errors"
	"github.com/micovery/apigee-yaml-toolkit/pkg/apigee/v1"
	"os"
	"path/filepath"
)

func RenderBundle(cFlags *CommonFlags, validate bool, dryRun string) (error, []error) {
	var err error

	bundleOutputFile := cFlags.OutputFile

	//create a temporary location for rendering into
	tmpDir, err := os.MkdirTemp("", "render-*")
	if err != nil {
		return errors.New(err), nil
	}

	// render the template to a temporary location
	cFlags.OutputFile = filepath.Join(tmpDir, "apiproxy.yaml")
	err = RenderGenericTemplate(cFlags, false)
	if err != nil {
		return errors.New(err), nil
	}

	// create bundle from rendered template
	return v1.APIProxyModelYAML2Bundle(cFlags.OutputFile, bundleOutputFile, validate, dryRun)
}
