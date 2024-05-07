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
	"fmt"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/go-errors/errors"
	"os"
)

func main() {
	err := RootCmd.Execute()

	if err != nil {
		var multiErrors utils.MultiError
		isMultiErrors := errors.As(err, &multiErrors)

		if isMultiErrors {
			for i := 0; i < len(multiErrors.Errors) && i < 10; i++ {
				_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", multiErrors.Errors[i].Error())
			}
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		}

		if showStack {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, 0).Stack())
		}
		os.Exit(1)
	}

	return

}
