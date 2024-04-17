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
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
	"runtime/debug"
	"slices"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print tool version",
	RunE: func(cmd *cobra.Command, args []string) error {

		var versionInfo any
		goReleaserInfo := GetGoReleaserInfo()
		goInstallInfo := GetGoInstallInfo()
		goBuildVCSInfo := GetGoBuildVCSInfo()

		if goReleaserInfo != nil {
			//package built with go releaser
			versionInfo = goReleaserInfo
		} else if goInstallInfo != nil {
			//package built with go install
			versionInfo = goInstallInfo
		} else if goBuildVCSInfo != nil {
			//other builds
			versionInfo = goBuildVCSInfo
		}

		versionText, err := json.Marshal(versionInfo)
		if err != nil {
			return errors.New(err)
		}
		fmt.Printf("%s\n", string(versionText))
		return nil
	},
}

func GetGoBuildVCSInfo() any {
	info, _ := debug.ReadBuildInfo()
	getBuildSettingValue := func(key string, settings []debug.BuildSetting) string {
		index := slices.IndexFunc(settings, func(setting debug.BuildSetting) bool {
			return setting.Key == key
		})

		if index < 0 {
			return ""
		}

		return settings[index].Value
	}

	type GoBuildVCSInfo struct {
		VCSInfo struct {
			Revision string `json:"vcs.revision"`
			Time     string `json:"vcs.time"`
			Modified string `json:"modified"`
		} `json:"VCSInfo"`
	}

	revision := getBuildSettingValue("vcs.revision", info.Settings)
	if revision == "" {
		return nil
	}

	result := &GoBuildVCSInfo{}
	result.VCSInfo.Revision = revision
	result.VCSInfo.Time = getBuildSettingValue("vcs.time", info.Settings)
	result.VCSInfo.Modified = getBuildSettingValue("vcs.modified", info.Settings)

	return result
}

func GetGoInstallInfo() any {
	info, _ := debug.ReadBuildInfo()
	if info.Main.Version == "(devel)" {
		return nil
	}

	type GoInstallInfo struct {
		InstallInfo struct {
			Version string `json:"Version"`
			Sum     string `json:"Sum"`
		} `json:"InstallInfo"`
	}

	result := &GoInstallInfo{}
	result.InstallInfo.Version = info.Main.Version
	result.InstallInfo.Sum = info.Main.Sum

	return result

}

// GitTag this is populated through -ldflags during build
var GitTag = ""
var GitCommit = ""
var BuildTimestamp = ""

func GetGoReleaserInfo() any {
	if GitTag == "" {
		return nil
	}

	type GoReleaserInfo struct {
		ReleaseInfo struct {
			GitCommit string `json:"GitCommit"`
			GitTag    string `json:"GitTag"`
			Timestamp string `json:"Timestamp"`
		} `json:"ReleaseInfo"`
	}

	result := &GoReleaserInfo{}
	result.ReleaseInfo.GitTag = GitTag
	result.ReleaseInfo.GitCommit = GitCommit
	result.ReleaseInfo.Timestamp = BuildTimestamp

	return result
}
