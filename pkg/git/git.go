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

package git

import (
	"fmt"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/go-errors/errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ParseURI attempts to split a git-style URI into its three core components:
// the Repository URL, the Revision Reference (ref), and the Resource Path
// within the repository.
//
// 1. The function primarily recognizes common web-view formats:
//
//	a. GitHub web-view: <GIT_REPO>/blob/<REF>/<RESOURCE>
//	b. BitBucket web-view: <GIT_REPO>/src/<REF>/<RESOURCE>
//	c. GitLab web-view: <GIT_REPO>/-/blob/<REF>/<RESOURCE>
//
// Web-view URI Examples:
//   - https://github.com/org/repo/blob/main/path/to/file.yaml
//   - https://bitbucket.org/org/repo/src/commit-hash/path/to/file.yaml
//   - https://gitlab.com/org/repo/-/blob/v1.0.0/path/to/file.yaml
//
// 2. The function also accepts more generic style formats:
//
//	a. <GIT_REPO>/-/<REF>/<RESOURCE>
//	b. <GIT_REPO>/-/<REF>#<RESOURCE> (Use '#' when REF contains a forward slash '/')
//
// Generic Style URI Examples:
//   - ssh://git@gitlab.com/org/repo/-/main/path/to/resource.yaml
//   - git@github.com:org/repo/-/v2.0/path/to/resource.yaml
//   - https://git.example.com/project/repo/-/feature/new-api#path/to/resource.yaml
//
// It returns: (repoURL, ref, resourcePath, error)
func ParseURI(uri string) (string, string, string, error) {

	gitHubStyle := "/blob/"
	bitBucketStyle := "/src/"
	gitLabStyle := "/-/blob"
	genericStyle := "/-/"

	repoSeparators := []string{gitLabStyle, gitHubStyle, bitBucketStyle, genericStyle}
	var parts []string
	for _, sep := range repoSeparators {
		parts = []string{}
		reSep := regexp.MustCompile(sep)
		parts = reSep.Split(uri, 2)
		if len(parts) == 2 {
			break
		}
	}

	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("unknown git URI format: %s", uri)
	}

	repoURL := parts[0]
	refAndPath := parts[1] // Contains <ref>#<resource> or <ref>/<resource>
	refAndPath = strings.Trim(refAndPath, "#/")

	if !(strings.HasPrefix(repoURL, "https://") ||
		strings.HasPrefix(repoURL, "ssh://") ||
		strings.HasPrefix(uri, "git@")) {
		return "", "", "", errors.Errorf("git URI must start with a scheme (https://, ssh://, or with git@")
	}

	refSeparators := []string{"#", "/"}
	for _, sep := range refSeparators {
		parts = []string{}
		reSep := regexp.MustCompile(sep)
		parts = reSep.Split(refAndPath, 2)
		if len(parts) == 2 {
			break
		}
	}

	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("git URI must contain ref, and resource: %s", uri)
	}

	ref := parts[0]
	resource := parts[1]

	return repoURL, ref, resource, nil
}

// IsGitURI checks if a given URI string follows one of the recognized
// git-style URL formats that can be parsed by ParseURI.
func IsGitURI(uri string) bool {
	//if it's a local file, it can't possibly be a Git URI
	if _, err := os.Stat(uri); err == nil {
		return false
	}

	//if it parses as Git URI, then it must be one
	_, _, _, err := ParseURI(uri)
	return err == nil
}

// RequireGit checks for the presence of the 'git' executable in the system's PATH.
// It returns an error if the command is not found.
func RequireGit() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return errors.Errorf("git command not found in $PATH. error: %s", err)
	}
	return nil
}

// RequireGitVersion checks that the installed Git client meets or exceeds
// the specified major and minor version numbers.
// This is critical for ensuring features like sparse-checkout are available.
func RequireGitVersion(requiredMajor int, requiredMinor int) error {
	const versionRegex = `(\d+)\.(\d+)`
	re := regexp.MustCompile(versionRegex)

	out, err := utils.Run("git", "version")
	if err != nil {
		return errors.Errorf("failed to run 'git version': %s. %s", err, out)
	}

	matches := re.FindStringSubmatch(out)
	if len(matches) < 3 {
		return errors.Errorf("could find git version pattern in output: %s", strings.TrimSpace(out))
	}

	fullVersionMatch := matches[0]

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return errors.Errorf("could not parse git major version: %s", matches[1])
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return errors.Errorf("could not parse git minor version: %s", matches[2])
	}

	if !((major > requiredMajor) || (major == requiredMajor && minor >= requiredMinor)) {
		return errors.Errorf("installed git (%s) is too old. Minimum required version is %d.%d.", fullVersionMatch, requiredMajor, requiredMinor)
	}

	return nil
}

// FetchFile fetches a single file from a remote Git repository based on a
// git-style URI.
//
// It performs the following steps:
//  1. Checks if the 'git' executable is available (RequireGit).
//  2. Checks if the 'git' version meets the minimum requirement (v2.25 is required
//     for modern sparse-checkout features) (RequireGitVersion).
//  3. Parses the URI to extract the repo URL, ref, and resource path (ParseURI).
//  4. Creates a temporary directory.
//  5. Initializes a local git repository in the temporary directory.
//  6. Adds the remote repository.
//  7. Attempts to fetch the specified ref (branch, tag, or commit hash).
//  8. Initializes and sets up a **sparse-checkout** to fetch only the directory
//     containing the resource (to minimize fetch size and time). **This means
//     the contents of the directory where the resource resides are also
//     downloaded.**
//  9. Checks out the fetched 'local' branch.
//
// It returns the full local path to the fetched file, the path to the temporary
// directory, and an error if the operation fails at any step.
// The caller is responsible for cleaning up the returned temporary directory.
func FetchFile(file string) (string, string, error) {
	var err error

	if err = RequireGit(); err != nil {
		return "", "", err
	}
	if err = RequireGitVersion(2, 25); err != nil {
		return "", "", err
	}

	var templateDir string
	if templateDir, err = os.MkdirTemp("", "apigee-go-gen-"); err != nil {
		return "", "", errors.New(err)
	}

	popd := utils.PushDir(templateDir)
	defer popd()

	var repoURL string
	var ref string
	var templatePath string

	if repoURL, ref, templatePath, err = ParseURI(file); err != nil {
		return "", "", err
	}

	var out string

	if out, err = utils.Run("git", "init"); err != nil {
		return "", "", errors.Errorf("git: could not init repo (%s). stdout/stderr: %s", err, out)
	}

	if out, err = utils.Run("git", "remote", "add", "origin", repoURL); err != nil {
		return "", "", errors.Errorf("git: could not add remote '%s' (%s). stdout/stderr: %s", repoURL, err, out)
	}

	refSpecs := []string{
		fmt.Sprintf("refs/heads/%s:refs/heads/local", ref), //branch
		fmt.Sprintf("refs/tags/%s:refs/heads/local", ref),  //tag
		fmt.Sprintf("%s:refs/heads/local", ref),            //hash
	}

	var lastOut string
	var lastErr error
	for _, refSpec := range refSpecs {
		lastOut, lastErr = utils.Run("git", "fetch", "--depth", "1", "origin", refSpec)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return "", "", errors.Errorf("git: could not fetch ref to '%s' (%s). stdout/stderr: %s", ref, lastErr, lastOut)
	}

	if out, err = utils.Run("git", "sparse-checkout", "init"); err != nil {
		return "", "", errors.Errorf("could not init sparse checkout on local git repo (%s). stdout/stderr: %s", err, out)
	}

	var checkoutDir = filepath.Dir(templatePath)
	if checkoutDir == "." {
		checkoutDir = "/"
	}

	if out, err = utils.Run("git", "sparse-checkout", "set", "--no-cone", checkoutDir); err != nil {
		return "", "", errors.Errorf("could not set sparse-checkout on local git repo (%s). stdout/stderr: %s", err, out)
	}

	if out, err = utils.Run("git", "checkout", "local"); err != nil {
		return "", "", errors.Errorf("could not checkout branch on local git repo (%s). stdout/stderr: %s", err, out)
	}

	//for security reasons, let's remove the .git directory right away
	if err = os.RemoveAll(filepath.Join(templateDir, ".git")); err != nil {
		return "", "", errors.Errorf("could not remove tempoary .git directory. %s", err)
	}

	file = filepath.Join(templateDir, templatePath)
	return file, templateDir, nil
}
