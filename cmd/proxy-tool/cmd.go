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

package main

import (
	"github.com/micovery/apigee-yaml-toolkit/cmd/proxy-tool/render"
	"github.com/micovery/apigee-yaml-toolkit/cmd/proxy-tool/transform"
	"github.com/micovery/apigee-yaml-toolkit/cmd/proxy-tool/version"
	"github.com/micovery/apigee-yaml-toolkit/pkg/flags"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var showStack = flags.NewBool(false)

var RootCmd = &cobra.Command{
	Use: filepath.Base(os.Args[0]),
}

func init() {
	RootCmd.AddCommand(render.Cmd)
	RootCmd.AddCommand(transform.Cmd)
	RootCmd.AddCommand(version.Cmd)

	RootCmd.PersistentFlags().Var(&showStack, "show-stack", "show stack trace for errors")
}
