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

package mcp

import (
	"github.com/go-errors/errors"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	"strings"
)

// FindYAMLReferences uses a recursive approach combined with JSONPath
// to find all references starting from docNode, resolving nested $refs
// against the provided openAPIRoot (which contains the components/schemas section).
// It returns a single yaml.MappingNode containing the collected and rewritten schemas.
func FindYAMLReferences(docNode *yaml.Node, openAPIRoot *yaml.Node) (*yaml.Node, error) {
	if openAPIRoot == nil {
		return nil, errors.New("openAPIRoot node must not be nil")
	}

	// Initialize the map to hold the resolved references and track cycles.
	resolvedRefs := make(map[string]*yaml.Node)

	// 1. Prepare the path expression to find all $ref scalar values within the document.
	pathExpr := "$..$ref"

	refPathFinder, err := yamlpath.NewPath(pathExpr)
	if err != nil {
		return nil, errors.Errorf("failed to compile initial JSONPath '%s': %s", pathExpr, err.Error())
	}

	// 2. Find all initial references from the starting document node
	initialRefs, err := refPathFinder.Find(docNode)
	if err != nil {
		return nil, errors.Errorf("initial JSONPath search failed: %s", err.Error())
	}

	// 3. Start the recursive resolution for each top-level reference found.
	for _, refNode := range initialRefs {
		if refNode.Kind == yaml.ScalarNode {
			refPath := refNode.Value

			if err := resolveRecursive(refPath, openAPIRoot, resolvedRefs); err != nil {
				return nil, err
			}
		}
	}

	// Create the final YAML MappingNode
	finalSchemaMap := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{}}

	// Iterate over the map and add key/value pairs to the MappingNode's Content
	for name, node := range resolvedRefs {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: name}
		finalSchemaMap.Content = append(finalSchemaMap.Content, keyNode, node)
	}

	// Apply the final transformation: rewrite all $ref paths
	// from "#/components/schemas/" to "#/$defs/".
	rewriteRefs(finalSchemaMap)

	return finalSchemaMap, nil
}

// resolveRecursive resolves the current reference, stores the result, and then
// recursively resolves any nested references within the newly found schema.
// Cycle detection is handled by the resolvedRefs map passed in. The openAPIRoot
// is the yaml node representing the entire OpenAPI document root.
func resolveRecursive(refPath string, openAPIRoot *yaml.Node, resolvedRefs map[string]*yaml.Node) error {
	// We use extractSchemaName here mainly to validate the format and get the key for the map.
	schemaName, err := extractSchemaName(refPath)
	if err != nil {
		return errors.Errorf("malformed $ref path encountered '%s': %s", refPath, err.Error())
	}

	// 1. Cycle/Duplicate check: If schema is already known, stop recursion.
	if _, exists := resolvedRefs[schemaName]; exists {
		// Already resolved or currently in-process (cycle detected).
		return nil
	}

	// 2. Mark as 'in-process' to prevent cycles
	resolvedRefs[schemaName] = nil

	// 3. Resolve the reference against the provided openAPIRoot
	sourceNode, err := ResolveYAMLReference(refPath, openAPIRoot)
	if err != nil {
		// Resolution failed, remove the marker and return error.
		delete(resolvedRefs, schemaName)
		return errors.Errorf("failed to resolve ref '%s': %s", refPath, err.Error())
	}

	// Deep clone the schema node to ensure we do not modify the original openAPIRoot structure.
	clonedNode := DeepCloneYAML(sourceNode)

	// 4. Store the resolved and cloned node
	resolvedRefs[schemaName] = clonedNode

	// 5. Look for nested references within the newly resolved schema definition (clonedNode)
	pathExpr := "$..$ref"

	refPathFinder, err := yamlpath.NewPath(pathExpr)
	if err != nil {
		return errors.Errorf("failed to compile nested JSONPath '%s': %s", pathExpr, err.Error())
	}

	// Search for nested references inside the cloned structure
	nestedRefs, err := refPathFinder.Find(clonedNode)
	if err != nil {
		return errors.Errorf("failed to search for nested refs in schema '%s': %s", schemaName, err.Error())
	}

	// 6. Recursively resolve nested references
	for _, nestedRefNode := range nestedRefs {
		if nestedRefNode.Kind == yaml.ScalarNode {
			nestedRefPath := nestedRefNode.Value

			// Recursive call, passing the map and the openAPIRoot
			if err := resolveRecursive(nestedRefPath, openAPIRoot, resolvedRefs); err != nil {
				return err // Bubble up error
			}
		}
	}

	return nil
}

// InlineYAMLReferences deep clones the document, then performs a single-pass recursive inlining
// of all references from rootDoc, ensuring no $refs remain and correctly handling cycles.
func InlineYAMLReferences(docNode *yaml.Node, rootDoc *yaml.Node) (*yaml.Node, error) {
	if docNode == nil {
		return nil, nil
	}

	if rootDoc == nil {
		return nil, errors.New("rootDoc node must not be nil")
	}

	// 1. Deep clone the input document to ensure the original is not modified.
	clonedDoc := DeepCloneYAML(docNode)

	// 2. Perform single-pass recursive inlining with cycle detection.
	// We use a map to track reference paths currently being processed on the recursion stack.
	if err := inlineRecursive(clonedDoc, rootDoc, make(map[string]struct{})); err != nil {
		return nil, err
	}

	return clonedDoc, nil
}

// inlineRecursive traverses the document and replaces $ref nodes with their resolved content.
// currentPath is used for cycle detection, tracking ref paths currently on the recursion stack.
func inlineRecursive(node *yaml.Node, rootDoc *yaml.Node, currentPath map[string]struct{}) error {
	if node == nil {
		return nil
	}

	// 1. Check if the current node itself is a $ref mapping.
	if isRefValueMap(node) {
		refPath := node.Content[1].Value

		// Cycle Detection Check
		if _, found := currentPath[refPath]; found {
			// Cycle detected: replace the node structure with an empty schema object {}
			// This effectively inlines an empty object in place of the cycle.
			node.Kind = yaml.MappingNode
			node.Content = []*yaml.Node{}
			node.Value = ""
			// Continue execution to step 2 to allow for traversal if the empty schema
			// was somehow complex (though it won't be here).
		} else {
			// Resolve the reference
			sourceNode, err := ResolveYAMLReference(refPath, rootDoc)
			if err != nil {
				return errors.Errorf("failed to inline ref '%s': %s", refPath, err.Error())
			}

			// Deep clone the source node
			clonedNode := DeepCloneYAML(sourceNode)

			// Mark the current path as 'in process' (before descending into the clone)
			currentPath[refPath] = struct{}{}

			// Recursively inline any nested references within the cloned content.
			if err := inlineRecursive(clonedNode, rootDoc, currentPath); err != nil {
				delete(currentPath, refPath) // Clean up path on error
				return err
			}

			// Unmark the path (done processing this path and all its descendants)
			delete(currentPath, refPath)

			// Replace the current node's fields with the fully inlined cloned schema's fields.
			// This replaces the $ref structure with the resolved schema contents in place.
			node.Kind = clonedNode.Kind
			node.Style = clonedNode.Style
			node.Tag = clonedNode.Tag
			node.Value = clonedNode.Value
			node.Content = clonedNode.Content

			// Continue execution below to allow further traversal if the resolved content
			// was a complex structure like a map or sequence.
		}
	}

	// 2. Traverse children. All nested $ref handling is now implicitly managed by the
	// recursive call hitting the check in step 1.
	if node.Kind == yaml.MappingNode {
		// Iterate over map content (key, value, key, value, ...)
		for i := 1; i < len(node.Content); i += 2 {
			valueNode := node.Content[i]

			// Recursively call on the value node. If valueNode is a $ref mapping,
			// the call will resolve and replace its content in-place (in step 1).
			if err := inlineRecursive(valueNode, rootDoc, currentPath); err != nil {
				return err
			}
		}
	} else if node.Kind == yaml.SequenceNode || node.Kind == yaml.DocumentNode {
		// If it's a sequence or document node, recurse into all content
		for _, contentNode := range node.Content {
			if err := inlineRecursive(contentNode, rootDoc, currentPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// isRefValueMap checks if a node is a mapping node whose only content is a $ref key-value pair.
func isRefValueMap(n *yaml.Node) bool {
	return n.Kind == yaml.MappingNode && len(n.Content) == 2 &&
		n.Content[0].Kind == yaml.ScalarNode && n.Content[0].Value == "$ref" &&
		n.Content[1].Kind == yaml.ScalarNode
}

// rewriteRefs recursively traverses a node and changes all $ref paths
// from "#/components/schemas/" to "#/$defs/".
func rewriteRefs(node *yaml.Node) {
	if node == nil {
		return
	}

	// Check if the current node is a scalar node representing a $ref value
	// We use strings.HasPrefix check for performance/specificity.
	const oldPrefix = "#/components/schemas/"
	const newPrefix = "#/$defs/"

	if node.Kind == yaml.ScalarNode && strings.HasPrefix(node.Value, oldPrefix) {
		node.Value = strings.Replace(node.Value, oldPrefix, newPrefix, 1)
		return // Scalar node processed, stop recursion on this path
	}

	// Recurse into children (content) for non-scalar nodes (Mapping, Sequence, Document)
	for _, contentNode := range node.Content {
		rewriteRefs(contentNode)
	}
}

func extractSchemaName(refPath string) (string, error) {
	allowedPrefixes := []string{"#/components/schemas/", "#/components/parameters/", "#/$defs/"}
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(refPath, prefix) {
			schemaName := strings.TrimPrefix(refPath, prefix)
			if schemaName == "" {
				return "", errors.Errorf("$ref reference '%s' must contain a schema name", refPath)
			}
			return schemaName, nil
		}
	}

	return "", errors.Errorf("$ref path '%s' must point to one of %v", refPath, allowedPrefixes)
}

func ResolveYAMLReference(refPath string, openAPIRoot *yaml.Node) (*yaml.Node, error) {
	// 1. Ensure the refPath is a local schema reference we can handle
	schemaName, err := extractSchemaName(refPath)
	if err != nil {
		return nil, errors.Errorf("invalid local reference format: %s", err.Error())
	}

	jsonPath, err := JSONPointerToJSONPath(refPath)
	if err != nil {
		return nil, errors.Errorf("failed to convert JSON Pointer '%s' to JSONPath: %s", refPath, err.Error())
	}

	path, err := yamlpath.NewPath(jsonPath)
	if err != nil {
		return nil, errors.Errorf("failed to compile JSONPath '%s': %s", jsonPath, err.Error())
	}

	// 3. Execute find against the openAPIRoot document.
	foundNodes, err := path.Find(openAPIRoot)
	if err != nil {
		return nil, errors.Errorf("JSONPath search failed for '%s' against openAPIRoot: %s", jsonPath, err.Error())
	}

	if len(foundNodes) == 0 {
		return nil, errors.Errorf("JSONPath found no matching schema node for '%s' in openAPIRoot", schemaName)
	}
	if len(foundNodes) > 1 {
		return nil, errors.Errorf("JSONPath returned multiple nodes (%d) for unique reference '%s'", len(foundNodes), jsonPath)
	}

	return foundNodes[0], nil
}

// DeepCloneYAML creates a full, recursive copy of a yaml.Node.
func DeepCloneYAML(n *yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}

	clone := &yaml.Node{
		Kind:        n.Kind,
		Style:       n.Style,
		Tag:         n.Tag,
		Value:       n.Value,
		Line:        n.Line,
		Column:      n.Column,
		LineComment: n.LineComment,
		HeadComment: n.HeadComment,
		FootComment: n.FootComment,
	}

	if n.Content != nil {
		clone.Content = make([]*yaml.Node, len(n.Content))
		for i, contentNode := range n.Content {
			clone.Content[i] = DeepCloneYAML(contentNode)
		}
	}
	return clone
}

// JSONPointerToJSONPath converts a JSON Pointer (like "#/a/b") to a JSONPath expression (like "$['a']['b']").
// This correctly handles segments containing special characters like dots (e.g., "$['a.b']").
// It also correctly unescapes JSON Pointer's ~0 and ~1 sequences.
func JSONPointerToJSONPath(refPath string) (string, error) {
	if !strings.HasPrefix(refPath, "#") {
		return "", errors.New("reference path must be a local JSON Pointer starting with '#'")
	}

	// 1. Remove '#' prefix (e.g., "/components/schemas/User.v1")
	jsonPointer := strings.TrimPrefix(refPath, "#")

	// 2. Split into segments. JSON Pointer segments are separated by '/'
	// The first segment will be empty due to the leading '/'.
	segments := strings.Split(jsonPointer, "/")

	// Start with the root symbol $
	jsonPathParts := []string{"$"}

	// Iterate over segments starting from index 1 (index 0 is empty string)
	for i := 1; i < len(segments); i++ {
		segment := segments[i]

		// Unescape JSON Pointer special characters: ~1 -> / and ~0 -> ~
		// Order matters: must replace ~1 first.
		unescapedSegment := strings.ReplaceAll(segment, "~1", "/")
		unescapedSegment = strings.ReplaceAll(unescapedSegment, "~0", "~")

		// Use single quotes for the JSONPath bracket notation. This is safe for keys containing dots.
		// NOTE: This assumes schema names do not contain single quotes (').
		jsonPathParts = append(jsonPathParts, "['"+unescapedSegment+"']")
	}

	return strings.Join(jsonPathParts, ""), nil
}
