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

package values

import (
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/go-errors/errors"
	"regexp"
	"strconv"
	"strings"
)

type Map map[string]any
type Slice []any

func (m *Map) Set(key string, value any) {
	regex := regexp.MustCompile(`(?:[^.\[\]]+|(?:\[\d+\]))`)
	keyParts := regex.FindAllString(key, -1)
	utils.Must(set(m, keyParts, 0, value))

}

func isNamedSlice(key string) bool {
	return strings.Index(key, "[") > 0
}

func isUnnamedSlice(key string) bool {
	return strings.Index(key, "[") == 0
}

func sliceGrow(parent *Slice, index int) *Slice {
	need := index - len(*parent) + 1
	if need < 0 {
		return parent
	}

	for i := 0; i < need; i++ {
		*parent = append(*parent, nil)
	}

	return parent
}

func getSliceKeyParts(key string) (string, int, error) {
	regex := regexp.MustCompile(`([^\[]+)?\[(\d+)\]`)

	submatch := regex.FindStringSubmatch(key)

	if len(submatch) == 0 {
		return "", 0, errors.Errorf("%s is not a valid slice key", key)
	}

	namePart := ""
	indexPart := ""

	if len(submatch) == 3 {
		namePart = submatch[1]
		indexPart = submatch[2]
	} else {
		return "", 0, errors.Errorf("%s is not a valid slice key", key)
	}

	parsedIndex, err := strconv.ParseInt(indexPart, 10, 64)
	if err != nil {
		return "", 0, errors.Errorf("%s is not a valid slice key", key)
	}

	return namePart, int(parsedIndex), nil
}

func set(parent any, keyParts []string, keyIndex int, value any) error {

	cKey := keyParts[keyIndex]

	if keyIndex == len(keyParts)-1 {
		//the end
		switch typedParent := parent.(type) {
		case map[string]any:
			return set(&typedParent, keyParts, keyIndex, value)
		case Map:
			return set(&typedParent, keyParts, keyIndex, value)
		case []any:
			return set(&typedParent, keyParts, keyIndex, value)
		case Slice:
			return set(&typedParent, keyParts, keyIndex, value)
		case *map[string]any:
			return set((*Map)(typedParent), keyParts, keyIndex, value)
		case *Map:
			if isUnnamedSlice(cKey) {
				return errors.Errorf("cannot set key %s on map type", cKey)
			}
			(*typedParent)[cKey] = value
		case *[]any:
			return set((*Slice)(typedParent), keyParts, keyIndex, value)
		case *Slice:
			if !isUnnamedSlice(cKey) {
				return errors.Errorf("cannot key set %s on slice type", cKey)
			}

			_, index, err := getSliceKeyParts(cKey)
			if err != nil {
				return err
			}

			typedParent = sliceGrow(typedParent, index)
			(*typedParent)[index] = value
		default:
			return errors.Errorf("cannot set value on type %T", parent)
		}
	} else {
		//walk down
		switch typedParent := parent.(type) {
		case map[string]any:
			return set(&typedParent, keyParts, keyIndex, value)
		case Map:
			return set(&typedParent, keyParts, keyIndex, value)
		case *map[string]any:
			return set((*Map)(typedParent), keyParts, keyIndex, value)
		case *Map:
			if isUnnamedSlice(cKey) {
				return errors.Errorf("cannot set key %s on map type", cKey)
			}
			newParent, hasChild := (*typedParent)[cKey]
			if !hasChild {
				nextKey := keyParts[keyIndex+1]
				if isUnnamedSlice(nextKey) {
					//create a slice at this location
					newParent = []any{}
				} else {
					//create a map at this location
					newParent = map[string]any{}
				}
				(*typedParent)[cKey] = newParent
			}
			return set(newParent, keyParts, keyIndex+1, value)
		case []any:
			return set(&typedParent, keyParts, keyIndex, value)
		case Slice:
			return set(&typedParent, keyParts, keyIndex, value)
		case *[]any:
			return set((*Slice)(typedParent), keyParts, keyIndex, value)
		case *Slice:
			if !isUnnamedSlice(cKey) {
				return errors.Errorf("cannot key set %s on slice type", cKey)
			}

			_, index, err := getSliceKeyParts(cKey)
			if err != nil {
				return err
			}

			typedParent = sliceGrow(typedParent, index)
			nextKey := keyParts[keyIndex+1]
			var newParent any
			if isUnnamedSlice(nextKey) {
				//create a slice at this location
				newParent = []any{}
			} else {
				//create a map at this location
				newParent = map[string]any{}
			}

			(*typedParent)[index] = newParent

			return set(newParent, keyParts, keyIndex+1, value)
		default:
			return errors.Errorf("cannot set value on type %T", parent)
		}

	}

	return nil

}
