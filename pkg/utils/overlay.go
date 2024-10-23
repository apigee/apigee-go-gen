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

package utils

import (
	"github.com/go-errors/errors"
	libopenapijson "github.com/pb33f/libopenapi/json"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	"path"
	"path/filepath"
	"slices"
)

func OASOverlay(overlayFile string, specFile string, outputFile string) error {
	overlayText, err := ReadInputTextFile(overlayFile)
	if err != nil {
		return err
	}

	var overlayNode *yaml.Node
	overlayNode = &yaml.Node{}
	err = yaml.Unmarshal(overlayText, overlayNode)
	if err != nil {
		return errors.New(err)
	}

	extendsField, err := getExtends(overlayNode, overlayFile)
	if err != nil {
		return err
	}

	//if extends field is set, and there is no spec file specified, use the extends field
	if extendsField != nil && specFile == "" {
		if path.IsAbs(*extendsField) {
			specFile = *extendsField
		} else {
			//when using extends, if path is relative, then it's relative to the overlay itself
			dir := path.Dir(overlayFile)
			specFile = path.Join(dir, *extendsField)
		}
	}

	if specFile == "" {
		return errors.Errorf("neither --spec parameter nor Overlay 'extends' field  was specified")
	}

	specText, err := ReadInputTextFile(specFile)
	if err != nil {
		return err
	}

	var specNode *yaml.Node
	specNode = &yaml.Node{}
	err = yaml.Unmarshal(specText, specNode)
	if err != nil {
		return errors.New(err)
	}

	//verify we are actually working with OAS3
	if slices.IndexFunc(specNode.Content[0].Content, func(n *yaml.Node) bool {
		return n.Value == "openapi"
	}) < 0 {
		return errors.Errorf("%s is not an OpenAPI 3.X spec file", specFile)
	}

	resultNode, err := ApplyOASOverlay(overlayNode, specNode, overlayFile, specFile)
	if err != nil {
		return err
	}

	ext := filepath.Ext(outputFile)
	if ext == "" {
		ext = filepath.Ext(specFile)
	}

	//depending on the file extension write the outputFile as either JSON or YAML
	var outputText []byte
	if ext == ".json" {
		outputText, err = libopenapijson.YAMLNodeToJSON(resultNode, "  ")
		if err != nil {
			return errors.New(err)
		}
	} else {
		outputText, err = YAML2Text(UnFlowYAMLNode(resultNode), 2)
		if err != nil {
			return err
		}
	}

	return WriteOutputText(outputFile, outputText)
}

func ApplyOASOverlay(overlayNode *yaml.Node, specNode *yaml.Node, overlayFile string, specFile string) (*yaml.Node, error) {
	actions, err := getActions(overlayNode, overlayFile)
	if err != nil {
		return nil, err
	}

	var newSpec *yaml.Node = specNode
	for _, action := range actions {
		newSpec, err = ApplyOASOverlayAction(action, newSpec, overlayFile, specFile)
		if err != nil {
			return nil, err
		}
	}

	return specNode, nil
}

func ApplyOASOverlayAction(action *yaml.Node, specNode *yaml.Node, overlayFile string, specFile string) (*yaml.Node, error) {

	targetNode, err := getActionTarget(action, overlayFile)
	if err != nil {
		return nil, err
	}

	updateNode, err := getActionUpdate(action, overlayFile)
	if err != nil {
		return nil, err
	}

	remove, err := getActionRemove(action, overlayFile)
	if err != nil {
		return nil, err
	}

	if targetNode == nil {
		return nil, errors.Errorf("'target' field is required for action")
	}

	if remove == nil && updateNode == nil {
		return nil, errors.Errorf("action does not contain neither 'remove' nor 'update' field at %s:%d", overlayFile, action.Line)
	} else if remove != nil && *remove == true {
		//handle remove action
		specNode, err = removeYAMLNode(specNode, targetNode, overlayFile, specFile)
		if err != nil {
			return nil, err
		}
	} else if updateNode != nil {
		//handle updateNode action
		specNode, err = updateYAMLNode(specNode, targetNode, updateNode, overlayFile, specFile)
		if err != nil {
			return nil, err
		}
	} else {
		//no-op action
		return specNode, nil
	}

	return specNode, nil
}

func getActionTarget(actionNode *yaml.Node, overlayFile string) (*yaml.Node, error) {
	targetPath, err := yamlpath.NewPath("$.target")
	if err != nil {
		return nil, errors.New(err)
	}

	results, err := targetPath.Find(actionNode)
	if err != nil {
		return nil, errors.New(err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	targetNode := results[0]

	if targetNode.Kind != yaml.ScalarNode {
		return nil, errors.Errorf("'target' field within overlay action is not a string at %s:%d", overlayFile, targetNode.Line)
	}

	return targetNode, nil
}

func getActionUpdate(actionNode *yaml.Node, overlayFile string) (*yaml.Node, error) {
	updatePath, err := yamlpath.NewPath("$.update")
	if err != nil {
		return nil, errors.New(err)
	}

	results, err := updatePath.Find(actionNode)
	if err != nil {
		return nil, errors.New(err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return results[0], nil
}

func getActionRemove(actionNode *yaml.Node, overlayFile string) (*bool, error) {
	var remove bool
	removePath, err := yamlpath.NewPath("$.remove")
	if err != nil {
		return &remove, errors.New(err)
	}

	results, err := removePath.Find(actionNode)
	if err != nil {
		return nil, errors.New(err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	removeNode := results[0]

	if removeNode.Kind != yaml.ScalarNode ||
		!(removeNode.Value == "true" || removeNode.Value == "false") {
		return nil, errors.Errorf("'remove' field within overlay action is not boolean at %s:%d", overlayFile, removeNode.Line)
	}

	remove = results[0].Value == "true"
	return &remove, nil
}

func getActions(overlayNode *yaml.Node, overlayFile string) ([]*yaml.Node, error) {
	actionsPath, err := yamlpath.NewPath("$.actions")
	if err != nil {
		return nil, errors.New(err)
	}

	actionsNodes, err := actionsPath.Find(overlayNode)
	if err != nil {
		return nil, errors.New(err)
	}

	if len(actionsNodes) == 0 {
		return nil, nil
	}

	actionsNode := actionsNodes[0]

	if actionsNode.Kind != yaml.SequenceNode {
		return nil, errors.Errorf("'actions' field must be an array at %s:%d", overlayFile, actionsNode.Line)
	}

	return actionsNode.Content, nil
}

func getExtends(overlayNode *yaml.Node, overlayFile string) (*string, error) {
	var extends *string
	extendsPath, err := yamlpath.NewPath("$.extends")
	if err != nil {
		return extends, errors.New(err)
	}

	results, err := extendsPath.Find(overlayNode)
	if err != nil {
		return extends, errors.New(err)
	}

	if len(results) == 0 {
		return extends, nil
	}

	extendsNode := *results[0]
	if extendsNode.Kind != yaml.ScalarNode {
		return nil, errors.Errorf("extends field is not a string at %s:%d", overlayFile, extendsNode.Line)
	}

	return &extendsNode.Value, nil
}

func updateYAMLNode(root *yaml.Node, targetNode *yaml.Node, updateNode *yaml.Node, overlayFile string, specFile string) (*yaml.Node, error) {
	pathToUpdate, err := yamlpath.NewPath(targetNode.Value)
	if err != nil {
		return nil, errors.Errorf("%s at %s:%d", err.Error(), overlayFile, targetNode.Line)
	}

	nodesToUpdate, err := pathToUpdate.Find(root)
	if err != nil {
		return nil, errors.New(err)
	}

	if len(nodesToUpdate) == 0 {
		return root, nil
	}

	for _, nodeToUpdate := range nodesToUpdate {
		updateYAMLNodeRecursiveInPlace(nodeToUpdate, updateNode)
	}

	return root, nil

}

func removeYAMLNode(root *yaml.Node, targetNode *yaml.Node, overlayFile string, specFile string) (*yaml.Node, error) {
	pathToRemove, err := yamlpath.NewPath(targetNode.Value)
	if err != nil {
		return nil, errors.Errorf("%s at %s:%d", err.Error(), overlayFile, targetNode.Line)
	}

	nodesToRemove, err := pathToRemove.Find(root)
	if err != nil {
		return nil, errors.New(err)
	}

	if len(nodesToRemove) == 0 {
		return root, nil
	}

	for _, nodeToRemove := range nodesToRemove {
		removeYAMLNodeRecursiveInPlace(nodeToRemove, root)
	}

	return root, nil
}

func removeYAMLNodeRecursiveInPlace(needle *yaml.Node, haystack *yaml.Node) {
	if needle == nil {
		return
	}

	if haystack.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(haystack.Content); i += 2 {
			value := haystack.Content[i+1]
			if value == needle {
				haystack.Content = slices.Delete(haystack.Content, i, i+2)
				return
			} else {
				removeYAMLNodeRecursiveInPlace(needle, value)
			}
		}
	} else if haystack.Kind == yaml.DocumentNode {
		removeYAMLNodeRecursiveInPlace(needle, haystack.Content[0])
	} else if haystack.Kind == yaml.SequenceNode {
		for i := 0; i < len(haystack.Content); i += 1 {
			value := haystack.Content[i]
			if value == needle {
				haystack.Content = slices.Delete(haystack.Content, i, i+1)
				return
			} else {
				removeYAMLNodeRecursiveInPlace(needle, value)
			}
		}
	}

	return
}

func updateYAMLNodeRecursiveInPlace(target *yaml.Node, source *yaml.Node) {
	if target.Kind == yaml.MappingNode && source.Kind == yaml.MappingNode {
		//create a lookup map for the target
		lookupMap := map[string]*yaml.Node{}
		for i := 0; i+1 < len(target.Content); i += 2 {
			key := target.Content[i].Value
			value := target.Content[i+1]
			lookupMap[key] = value
		}

		//merge keys that match
		for i := 0; i+1 < len(source.Content); i += 2 {
			key := source.Content[i].Value
			if subTarget, found := lookupMap[key]; found {
				updateYAMLNodeRecursiveInPlace(subTarget, source.Content[i+1])
			} else {
				target.Content = append(target.Content, source.Content[i], source.Content[i+1])
			}
		}
		return
	}

	if target.Kind == yaml.SequenceNode && (source.Kind == yaml.MappingNode ||
		source.Kind == yaml.ScalarNode) {
		target.Content = append(target.Content, source)
	}

	if target.Kind == yaml.ScalarNode && source.Kind == yaml.ScalarNode {
		target.Value = source.Value
		return
	}

	if target.Kind == yaml.SequenceNode && source.Kind == yaml.SequenceNode {
		target.Content = append(target.Content, source.Content...)
		return
	}

}
