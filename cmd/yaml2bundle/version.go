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

package main

import (
	"flag"
	"fmt"
	"github.com/micovery/apigee-yaml-toolkit/cmd/yaml2bundle/resources"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var BuildVersion = "0.0.0"
var BuildTimestamp = "0"
var BuildGitHash = "unset"

func PrintVersion() {
	program := filepath.Base(os.Args[0])
	ts, _ := strconv.Atoi(BuildTimestamp)
	text, _ := time.Unix(int64(ts), 0).UTC().MarshalText()
	fmt.Printf("%s version=%s date=%s commit=%s\n", program, BuildVersion, text, BuildGitHash)
}

func PrintUsage() {
	program := filepath.Base(os.Args[0])
	usageText, err := resources.FS.ReadFile("usage.txt")

	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, string(usageText), program)

	flag.PrintDefaults()
}
