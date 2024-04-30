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

package sharedflow

import (
	"bytes"
	"fmt"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/apigee/apigee-go-gen/pkg/zip"
	"github.com/go-errors/errors"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

func Bundle2YAMLFile(bundle string, outputFile string, dryRun bool) error {
	extension := filepath.Ext(bundle)
	if extension == ".zip" {
		err := BundleZip2YAMLFile(bundle, outputFile, dryRun)
		if err != nil {
			return err
		}
	} else if extension != "" {
		return errors.Errorf("input extension %s is not supported", extension)
	} else {
		err := BundleDir2YAMLFile(bundle, outputFile, bool(dryRun))
		if err != nil {
			return err
		}
	}

	return nil
}

func BundleZip2YAMLFile(inputZip string, outputFile string, dryRun bool) error {
	tmpDir, err := os.MkdirTemp("", "unzipped-bundle-*")
	if err != nil {
		return errors.New(err)
	}

	err = zip.Unzip(tmpDir, inputZip)
	if err != nil {
		return errors.New(err)
	}

	return BundleDir2YAMLFile(tmpDir, outputFile, dryRun)

}

func BundleDir2YAMLFile(inputDir string, outputFile string, dryRun bool) error {
	policyFiles := []string{}
	sharedFlowsFiles := []string{}
	resourcesFiles := []string{}
	manifestFiles := []string{}

	sharedFlowBundleDir := filepath.Join(inputDir, "sharedflowbundle")
	stat, err := os.Stat(sharedFlowBundleDir)
	if err != nil {
		return errors.Errorf("%s not found. %s", sharedFlowBundleDir, err.Error())
	} else if !stat.IsDir() {
		return errors.Errorf("%s is not a directory", sharedFlowBundleDir)
	}

	fSys := os.DirFS(sharedFlowBundleDir)

	manifestFiles, _ = fs.Glob(fSys, "*.xml")
	policyFiles, _ = fs.Glob(fSys, "policies/*.xml")
	sharedFlowsFiles, _ = fs.Glob(fSys, "sharedflows/*.xml")
	resourcesFiles, _ = fs.Glob(fSys, "resources/*/*")

	allFiles := []string{}
	if len(manifestFiles) == 0 {
		return errors.Errorf("no shared flow XML file found in %s", sharedFlowBundleDir)
	}

	allFiles = append(allFiles, manifestFiles[0])
	allFiles = append(allFiles, policyFiles...)
	allFiles = append(allFiles, sharedFlowsFiles...)

	createMapEntry := func(parent *yaml.Node, key string, value *yaml.Node) *yaml.Node {
		parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, value)
		return value
	}

	fileToYAML := func(filePath string) (*yaml.Node, error) {
		fullPath := filepath.Join(sharedFlowBundleDir, filePath)
		fileContents, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, errors.New(err)
		}
		yamlNode, err := utils.XMLText2YAML(bytes.NewReader(fileContents))
		if err != nil {
			return nil, err
		}

		return yamlNode, nil
	}

	addSequence := func(parentNode *yaml.Node, key string, files []string) error {
		sequence := createMapEntry(parentNode, key, &yaml.Node{Kind: yaml.SequenceNode})
		for _, filePath := range files {
			yamlNode, err := fileToYAML(filePath)
			if err != nil {
				return err
			}
			if len(yamlNode.Content) > 0 {
				sequence.Content = append(sequence.Content, yamlNode)
			}
		}
		return nil
	}

	docNode := &yaml.Node{Kind: yaml.DocumentNode}
	mainNode := &yaml.Node{Kind: yaml.MappingNode}
	docNode.Content = append(docNode.Content, mainNode)

	manifestNode, err := fileToYAML(manifestFiles[0])
	if err != nil {
		return err
	}

	mainNode.Content = append(mainNode.Content, manifestNode.Content...)

	err = addSequence(mainNode, "Policies", policyFiles)
	if err != nil {
		return err
	}
	err = addSequence(mainNode, "SharedFlows", sharedFlowsFiles)
	if err != nil {
		return err
	}

	//copy resource files
	resourcesNode := createMapEntry(mainNode, "Resources", &yaml.Node{Kind: yaml.SequenceNode})
	for _, resourceFile := range resourcesFiles {
		dirName, fileName := filepath.Split(resourceFile)
		fileType := filepath.Base(dirName)

		location := path.Join(".", fileName)
		resourceNode := &yaml.Node{Kind: yaml.MappingNode}
		resourceDataNode := createMapEntry(resourceNode, "Resource", &yaml.Node{Kind: yaml.MappingNode})
		createMapEntry(resourceDataNode, "Type", &yaml.Node{Kind: yaml.ScalarNode, Value: fileType})
		createMapEntry(resourceDataNode, "Path", &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("./%s", location)})
		resourcesNode.Content = append(resourcesNode.Content, resourceNode)
		outputDir := filepath.Dir(outputFile)
		err := utils.CopyFile(filepath.Join(outputDir, fileName), filepath.Join(sharedFlowBundleDir, resourceFile))
		if err != nil {
			return err
		}
	}

	var docBytes []byte
	if docBytes, err = utils.YAML2Text(docNode, 2); err != nil {
		return err
	}

	if dryRun {
		fmt.Print(string(docBytes))
		return nil
	}

	err = utils.YAMLDoc2File(docNode, outputFile)
	if err != nil {
		return err
	}

	return nil
}
