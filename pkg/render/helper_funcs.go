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
	"fmt"
	"github.com/apigee/apigee-go-gen/pkg/globals"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/apigee/apigee-go-gen/pkg/utils/mcp"
	"github.com/go-errors/errors"
	"os"
	"path/filepath"
	"strings"
)

type HelperFunc func(args ...any) string

func getOSCopyFileFunc(templateFile string, outputFile string, dryRun bool) HelperFunc {
	_osCopyFileFunc := func(args ...any) string {
		defer recoverPanic()

		if len(args) < 2 {
			panic("os_copyfile function requires two arguments")
		}

		dst := args[0].(string)
		src := args[1].(string)

		if filepath.IsAbs(dst) {
			panic("os_copyfile dst must not be absolute")
		}
		if strings.Index(dst, "..") >= 0 {
			panic("os_copyfile dst must not use ..")
		}

		if filepath.IsAbs(src) {
			panic("os_copyfile src must not be absolute")
		}
		if strings.Index(src, "..") >= 0 {
			panic("os_copyfile src must not use ..")
		}

		dstDir := filepath.Dir(filepath.Join(filepath.Dir(outputFile), dst))

		if !dryRun {
			err := os.MkdirAll(dstDir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}

		srcPath := filepath.Join(filepath.Dir(templateFile), src)
		dstPath := filepath.Join(filepath.Dir(outputFile), dst)

		if !dryRun {
			srcFileContent, err := os.ReadFile(srcPath)
			if err != nil {
				panic(err)
			}

			err = os.WriteFile(dstPath, srcFileContent, os.ModePerm)
			if err != nil {
				panic(err)
			}
		} else {
			//fmt.Printf(`os_copyfile("%s", "%s")\n`, dstPath, srcPath)
		}

		return dst
	}

	return _osCopyFileFunc
}

func getRemoveOASExtensions(templateFile string, outputFile string, dryRun bool) HelperFunc {
	_removeOASExtensionsFunc := func(args ...any) string {
		defer recoverPanic()

		if len(args) < 1 {
			panic("remove_oas_extensions function requires one argument")
		}

		//both destination and source are the same
		dst := args[0].(string)
		src := args[0].(string)

		if filepath.IsAbs(src) {
			panic("remove_oas_extensions src must not be absolute")
		}
		if strings.Index(src, "..") >= 0 {
			panic("remove_oas_extensions src must not use ..")
		}

		//both destination and source are relative to the output file
		dstPath := filepath.Join(filepath.Dir(outputFile), dst)
		srcPath := filepath.Join(filepath.Dir(outputFile), src)

		if !dryRun {
			err := utils.RemoveExtensions(srcPath, dstPath)
			if err != nil {
				panic(err)
			}
		} else {
			//fmt.Printf(`remove_oas_extensions("%s", "%s")\n`, dstPath, srcPath)
		}

		return dst
	}

	return _removeOASExtensionsFunc
}

func getRemoveOASSchemaExtensions(templateFile string, outputFile string, dryRun bool) HelperFunc {
	_removeOASExtensionsFunc := func(args ...any) string {
		defer recoverPanic()

		if len(args) < 1 {
			panic("remove_oas_schema_extensions function requires one argument")
		}

		//both destination and source are the same
		dst := args[0].(string)
		src := args[0].(string)

		if filepath.IsAbs(src) {
			panic("remove_oas_schema_extensions src must not be absolute")
		}
		if strings.Index(src, "..") >= 0 {
			panic("remove_oas_schema_extensions src must not use ..")
		}

		//both destination and source are relative to the output file
		dstPath := filepath.Join(filepath.Dir(outputFile), dst)
		srcPath := filepath.Join(filepath.Dir(outputFile), src)

		if !dryRun {
			err := utils.RemoveSchemaExtensions(srcPath, dstPath)
			if err != nil {
				panic(err)
			}
		} else {
			//fmt.Printf(`remove_oas_schema_extensions("%s", "%s")\n`, dstPath, srcPath)
		}

		return dst
	}

	return _removeOASExtensionsFunc
}

func convertOAS3ToMCPValues(args ...any) map[string]any {
	defer recoverPanic()

	if len(args) < 1 {
		panic("oas3_to_mcp function requires one arguments")
	}

	oas3file := args[0].(string)
	var mcpValuesMap map[string]any

	var err error
	if mcpValuesMap, err = mcp.OAS3ToMCPValues(oas3file); err != nil {
		panic(err)
	}

	return mcpValuesMap
}

func convertYAMLTextToJSON(args ...any) string {
	if len(args) < 1 {
		panic("yaml_to_json function requires one argument")
	}

	yamlText := args[0].(string)

	var jsonText []byte
	var err error

	if jsonText, err = utils.YAMLText2JSONText(strings.NewReader(yamlText)); err != nil {
		panic(err)
	}

	return string(jsonText)
}

func convertJSONToYAML(args ...any) string {
	if len(args) < 1 {
		panic("json_to_yaml function requires one argument")
	}

	jsonText := args[0].(string)

	var yamlText []byte
	var err error

	if yamlText, err = utils.JSONText2YAMLText(strings.NewReader(jsonText)); err != nil {
		panic(err)
	}

	return string(yamlText)
}

func recoverPanic() {
	if r := recover(); r != nil {
		err := errors.Wrap(r, 1)
		if *globals.ShowStack {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Stack())
		}
		panic(err)
	}
}
