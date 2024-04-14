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

package bundle

import (
	"bytes"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/micovery/apigee-yaml-toolkit/pkg/apigee/v1"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"github.com/micovery/apigee-yaml-toolkit/pkg/zip"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func ProxyBundle2YAMLFile(proxyBundle string, outputFile string, dryRun utils.FlagBool) error {
	extension := filepath.Ext(proxyBundle)
	if extension == ".zip" {
		err := ProxyBundleZip2YAMLFile(proxyBundle, outputFile, dryRun)
		if err != nil {
			return err
		}
	} else if extension != "" {
		return errors.Errorf("input extension %s is not supported", extension)
	} else {
		err := ProxyBundleDir2YAMLFile(proxyBundle, outputFile, dryRun)
		if err != nil {
			return err
		}
	}

	return nil
}

func ProxyBundleZip2YAMLFile(inputZip string, outputFile string, dryRun utils.FlagBool) error {
	tmpDir, err := os.MkdirTemp("", "unzipped-bundle-*")
	if err != nil {
		return errors.New(err)
	}

	err = zip.Unzip(tmpDir, inputZip)
	if err != nil {
		return errors.New(err)
	}

	return ProxyBundleDir2YAMLFile(tmpDir, outputFile, dryRun)

}

func ProxyBundleDir2YAMLFile(inputDir string, outputFile string, dryRun utils.FlagBool) error {
	policyFiles := []string{}
	proxyEndpointsFiles := []string{}
	targetEndpointsFiles := []string{}
	resourcesFiles := []string{}
	manifestFiles := []string{}

	apiProxyDir := filepath.Join(inputDir, "apiproxy")
	stat, err := os.Stat(apiProxyDir)
	if err != nil {
		return errors.Errorf("%s not found. %s", apiProxyDir, err.Error())
	} else if !stat.IsDir() {
		return errors.Errorf("%s is not a directory", apiProxyDir)
	}

	fSys := os.DirFS(apiProxyDir)

	manifestFiles, _ = fs.Glob(fSys, "*.xml")
	policyFiles, _ = fs.Glob(fSys, "policies/*.xml")
	proxyEndpointsFiles, _ = fs.Glob(fSys, "proxies/*.xml")
	targetEndpointsFiles, _ = fs.Glob(fSys, "targets/*.xml")
	resourcesFiles, _ = fs.Glob(fSys, "resources/*/*")

	allFiles := []string{}
	if len(manifestFiles) == 0 {
		return errors.Errorf("no proxy XML file found in %s", apiProxyDir)
	}

	allFiles = append(allFiles, manifestFiles[0])
	allFiles = append(allFiles, policyFiles...)
	allFiles = append(allFiles, proxyEndpointsFiles...)
	allFiles = append(allFiles, targetEndpointsFiles...)

	createMapEntry := func(parent *yaml.Node, key string, value *yaml.Node) *yaml.Node {
		parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, value)
		return value
	}

	fileToYAML := func(filePath string) (*yaml.Node, error) {
		fullPath := filepath.Join(apiProxyDir, filePath)
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
	err = addSequence(mainNode, "ProxyEndpoints", proxyEndpointsFiles)
	if err != nil {
		return err
	}
	err = addSequence(mainNode, "TargetEndpoints", targetEndpointsFiles)
	if err != nil {
		return err
	}

	//copy resource files
	resourcesNode := createMapEntry(mainNode, "Resources", &yaml.Node{Kind: yaml.MappingNode})
	for _, resourceFile := range resourcesFiles {
		dirName, fileName := filepath.Split(resourceFile)
		fileType := filepath.Base(dirName)

		location := path.Join(".", fileName)
		resourceDataNode := createMapEntry(resourcesNode, "Resource", &yaml.Node{Kind: yaml.MappingNode})
		createMapEntry(resourceDataNode, "Type", &yaml.Node{Kind: yaml.ScalarNode, Value: fileType})
		createMapEntry(resourceDataNode, "Path", &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("./%s", location)})

		outputDir := filepath.Dir(outputFile)
		err := utils.CopyFile(filepath.Join(outputDir, fileName), filepath.Join(apiProxyDir, resourceFile))
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

	err = YAMLDoc2File(docNode, outputFile, manifestFiles[0])
	if err != nil {
		return err
	}

	return nil
}

func YAMLDoc2File(docNode *yaml.Node, outputFile string, fileName string) error {
	var err error
	var docBytes []byte
	if docBytes, err = utils.YAML2Text(docNode, 2); err != nil {
		return err
	}

	//generate output directory
	outputDir := filepath.Dir(outputFile)
	if err = os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return errors.New(err)
	}

	//generate the main YAML file
	fileName = filepath.Base(fileName)
	fileName = fmt.Sprintf("%s.yaml", strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	if err = os.WriteFile(outputFile, docBytes, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func ProxyBundle2Disk(proxyModel *v1.APIProxyModel, output string) error {
	extension := filepath.Ext(output)
	if extension == ".zip" {
		err := ProxyBundle2Zip(proxyModel, output)
		if err != nil {
			return err
		}
	} else if extension != "" {
		return errors.Errorf("output extension %s is not supported", extension)
	} else {
		err := ProxyBundle2Dir(proxyModel, output)
		if err != nil {
			return err
		}
		//directory
	}

	return nil
}

func ProxyBundle2Dir(proxyModel *v1.APIProxyModel, output string) error {

	err := os.MkdirAll(output, os.ModePerm)
	if err != nil {
		return errors.New(err)
	}

	bundleFiles := proxyModel.GetBundleFiles()
	for _, bundleFile := range bundleFiles {
		filePath := bundleFile.FilePath()
		fileDir := filepath.Dir(filePath)

		dirDiskPath := filepath.Join(output, fileDir)
		err := os.MkdirAll(dirDiskPath, os.ModePerm)
		if err != nil {
			return errors.New(err)
		}

		fileContent, err := bundleFile.FileContents()
		if err != nil {
			return err
		}

		fileDiskPath := filepath.Join(output, filePath)
		err = os.WriteFile(fileDiskPath, fileContent, os.ModePerm)
		if err != nil {
			return errors.New(err)
		}
	}

	return nil
}

func ProxyBundle2Zip(proxyModel *v1.APIProxyModel, outputZip string) error {
	tmpDir, err := os.MkdirTemp("", "unzipped-bundle-*")
	if err != nil {
		return errors.New(err)
	}

	err = ProxyBundle2Disk(proxyModel, tmpDir)
	if err != nil {
		return err
	}

	err = zip.Zip(outputZip, tmpDir)
	if err != nil {
		return err
	}

	return nil
}
