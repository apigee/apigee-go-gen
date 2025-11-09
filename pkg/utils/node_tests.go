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

package utils

import (
	"bufio"
	"os"
	"os/exec"
	"testing"
)

// RunNPMTTest checks for 'npm' and 'node' in PATH and skips the test if either is not found.
// If available, it changes the working directory to targetDir, runs 'npm install',
// and then executes 'npm run unit-test', streaming the output directly.
func RunNPMTTest(t *testing.T, targetDir string) {

	defer HandleTestPanic(t)

	// 1. Check for external dependencies
	if _, err := exec.LookPath("node"); err != nil {
		t.Skipf("Skipping NPM test: 'node' command not found in PATH. Error: %v", err)
		return
	}
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skipf("Skipping NPM test: 'npm' command not found in PATH. Error: %v", err)
		return
	}

	// 2. Check if the target directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		t.Fatalf("Target directory does not exist: %s. Please create it or update 'targetDir'.", targetDir)
	}

	var cmd *exec.Cmd
	t.Logf("Running 'npm install' in directory: %s", targetDir)
	cmd = exec.Command("npm", "install", "deploy")
	cmd.Dir = targetDir
	RunT(cmd, t)
	t.Logf("npm install successful.")

	t.Logf("Running 'npm test' in directory: %s", targetDir)
	cmd = exec.Command("npm", "test")
	cmd.Dir = targetDir
	RunT(cmd, t)
	t.Logf("All npm processes completed successfully.")

}

func HandleTestPanic(t *testing.T) {
	v := recover()
	if v != nil {
		t.Fatalf("Error: %v", v)
	}
}

func RunT(cmd *exec.Cmd, t *testing.T) {
	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	done := make(chan struct{})
	scanner := bufio.NewScanner(r)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			t.Logf(line)
		}
		done <- struct{}{}
	}()

	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	<-done
	err = cmd.Wait()
	if err != nil {
		panic(err)
	}
}
