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
