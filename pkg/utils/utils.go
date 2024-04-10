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
	"flag"
	"fmt"
	"github.com/go-errors/errors"
	"os"
	"path/filepath"
	"runtime/debug"
)

func RequireParamAndExit(param string) {
	fmt.Fprintf(os.Stderr, "error: -%s parameter is required\n", param)
	flag.PrintDefaults()
	os.Exit(1)
}

func PrintErrorWithStackAndExit(err error) {
	fmt.Printf("error: %s\n", err.Error())
	fmt.Printf("%s\n", errors.Wrap(err, 0).Stack())
	os.Exit(1)
}

func PrintVersion() {
	program := filepath.Base(os.Args[0])
	info, _ := debug.ReadBuildInfo()

	fmt.Printf("%s %s\n", program, info.Main.Version)
}
