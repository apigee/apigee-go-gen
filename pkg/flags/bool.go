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

package flags

import (
	"fmt"
	"github.com/go-errors/errors"
	"strconv"
)

type Bool bool

func (f *Bool) String() string {
	return fmt.Sprintf("%v", bool(*f))
}

func (f *Bool) Set(input string) error {

	value, err := strconv.ParseBool(input)
	if err != nil {
		return errors.New(err)
	}

	*f = Bool(value)
	return nil
}
