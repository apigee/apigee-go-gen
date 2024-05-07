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
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

type HelperFunc func(args ...any) string

func getOSCopyFileFunc(templateFile string, outputFile string, dryRun bool) HelperFunc {
	_osCopyFileFunc := func(args ...any) string {
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
