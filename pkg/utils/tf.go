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
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"path/filepath"
	"strings"
	// Assumes ReadInputText, WriteOutputText, JSONText2YAMLText are defined elsewhere or in another file
)

// TFFileToJSONFile converts a Terraform HCL file (.tf) to a JSON file.
func TFFileToJSONFile(input string, output string) error {
	text, err := ReadInputText(input)
	if err != nil {
		return err
	}

	fileName := input
	if input == "-" || len(input) == 0 {
		fileName = "input.tf" // Default to .tf
	}

	var jsonBytes []byte
	if jsonBytes, err = TFTextToJSONText(text, fileName); err != nil {
		return err
	}

	if output == "" || output == "-" {
		fmt.Println(string(jsonBytes))
		return nil
	}

	ext := filepath.Ext(output)
	if ext == "" {
		ext = filepath.Ext(input)
		// Ensure output has a .json extension if input was .tf
		if ext == ".tf" {
			output = strings.TrimSuffix(output, ext) + ".tf.json"
		} else if ext == "" {
			output += ".tf.json" // Default output extension
		}
	}

	//depending on the file extension write the output as either JSON or YAML (or stdout)
	var outputText []byte

	// Only handle .json output, YAML conversion might need separate logic
	if filepath.Ext(output) == ".json" || filepath.Ext(output) == ".tf.json" {
		outputText = jsonBytes
	} else if filepath.Ext(output) == ".yaml" {
		outputText, err = JSONText2YAMLText(strings.NewReader(string(jsonBytes)))
		if err != nil {
			return errors.Errorf("error converting JSON to YAML: %w", err)
		}
	} else {
		// Default to printing JSON to stdout if extension isn't recognized
		fmt.Printf(string(jsonBytes))
		return nil
	}

	return WriteOutputText(output, outputText)
}

// JSONFileToTFFile reads a Terraform JSON file (.tf.json), converts it to HCL text, and writes it.
func JSONFileToTFFile(input string, output string) error {
	text, err := ReadInputText(input)
	if err != nil {
		return err
	}

	fileName := input
	if input == "-" || len(input) == 0 {
		fileName = "input.tf.json" // Default to .tf.json
	}

	var tfBytes []byte
	if tfBytes, err = JSONTextToTFText(text, fileName); err != nil {
		return err
	}

	if output == "" || output == "-" {
		fmt.Println(string(tfBytes))
		return nil
	} else {
		ext := filepath.Ext(output)
		if ext == "" || ext != ".tf" {
			output = strings.TrimSuffix(output, ext) + ".tf"
		}
	}

	return WriteOutputText(output, tfBytes)
}

func TFText2HCLFile(fileText string, filePath string) (*hcl.File, error) {
	hclParser := hclparse.NewParser()

	fileName := filepath.Base(filePath)

	var parsed *hcl.File
	var diag hcl.Diagnostics
	if parsed, diag = hclParser.ParseHCL([]byte(fileText), fileName); diag != nil && diag.HasErrors() {
		errs := diag.Errs()
		return nil, errors.Join(errs...)
	}

	return parsed, nil
}

// TFTextToMap converts Terraform HCL text bytes to a Go map[string]any.
func TFTextToMap(text []byte, fileName string) (map[string]any, error) {
	var err error
	var hclFile *hcl.File
	// Ensure the parser handles Terraform HCL syntax
	if hclFile, err = TFText2HCLFile(string(text), fileName); err != nil {
		return nil, errors.Errorf("error parsing TF text: %w", err)
	}

	var resultMap map[string]any

	if resultMap, err = TFFileToMap(hclFile, fileName); err != nil {
		return nil, errors.Errorf("error converting TF file to map: %w", err)
	}
	return resultMap, nil
}

// TFTextToJSONText converts Terraform HCL text bytes to Terraform JSON text bytes.
func TFTextToJSONText(text []byte, fileName string) ([]byte, error) {
	var err error
	var resultMap map[string]any
	if resultMap, err = TFTextToMap(text, fileName); err != nil {
		return nil, err
	}

	var jsonBytes []byte
	jsonBytes, err = json.MarshalIndent(resultMap, "", "  ")
	if err != nil {
		return nil, errors.Errorf("error marshaling map to JSON: %w", err)
	}
	return jsonBytes, nil
}

// JSONTextToTFText converts Terraform JSON text bytes into formatted Terraform HCL text bytes.
func JSONTextToTFText(jsonBytes []byte, fileName string) ([]byte, error) {
	var data map[string]any
	err := json.Unmarshal(jsonBytes, &data)
	if err != nil {
		return nil, errors.Errorf("error unmarshaling JSON: %w", err)
	}

	hclBytes, err := MapToTFText(data, fileName)
	if err != nil {
		return nil, errors.Errorf("error converting map to HCL text: %w", err)
	}

	return hclBytes, nil
}
