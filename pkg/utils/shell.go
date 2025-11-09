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
	"os/exec"
	"strings"
)

func Run(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	var out strings.Builder
	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	done := make(chan struct{})
	scanner := bufio.NewScanner(r)
	go func() {
		for scanner.Scan() {
			line := scanner.Bytes()
			out.Write(line)
		}
		done <- struct{}{}
	}()

	err := cmd.Start()
	if err != nil {
		return out.String(), err
	}

	<-done
	err = cmd.Wait()
	if err != nil {
		return out.String(), err
	}

	return out.String(), nil
}
