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

package render

import (
	"github.com/apigee/apigee-go-gen/pkg/flags"
	"github.com/apigee/apigee-go-gen/pkg/git"
	"github.com/apigee/apigee-go-gen/pkg/utils"
)

func RenderGenericTemplate(cFlags *CommonFlags, dryRun bool) error {
	var err error
	if git.IsGitURI(string(cFlags.TemplateFile)) {
		var templateFileFromGit string
		var templateDirFromGit string
		if templateFileFromGit, templateDirFromGit, err = git.FetchFile(string(cFlags.TemplateFile)); err != nil {
			return err
		}
		defer utils.LenientRemoveAll(templateDirFromGit)
		cFlags.TemplateFile = flags.String(templateFileFromGit)
	}

	return RenderGenericTemplateLocal(cFlags, dryRun)
}

func RenderGenericTemplateLocal(cFlags *CommonFlags, dryRun bool) error {
	type TemplateContext struct {
		Values map[string]any
	}

	context := &TemplateContext{
		Values: *cFlags.Values,
	}

	return RenderGeneric(context, cFlags, dryRun)
}
