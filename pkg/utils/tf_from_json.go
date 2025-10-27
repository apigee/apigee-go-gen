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
	"bytes"
	"fmt"
	"github.com/go-errors/errors"
	"sort"
	"strconv"
	"strings"
	"unicode" // Needed for sanitization & identifier check
)

const indentUnit = "  "

// MapToTFText converts a map[string]any into a formatted Terraform HCL text byte slice.
func MapToTFText(data map[string]any, fileName string) ([]byte, error) {
	var buf bytes.Buffer
	err := mapToTFStringRecursive(data, "", "", &buf) // Start with empty blockType and indent
	if err != nil {
		return nil, err
	}
	// Trim final newline if present
	finalBytes := buf.Bytes()
	if len(finalBytes) > 0 && finalBytes[len(finalBytes)-1] == '\n' {
		finalBytes = finalBytes[:len(finalBytes)-1]
	}
	return finalBytes, nil
}

// mapToTFStringRecursive recursively builds the HCL string from a map.
func mapToTFStringRecursive(data map[string]any, blockType string, indent string, buf *bytes.Buffer) error {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Consistent order

	firstItemWritten := false
	lastItemType := ""     // "attribute" or "block"
	hasAttributes := false // Track if any attributes exist in this scope

	// First pass to check if there are any attributes
	for _, key := range keys {
		val := data[key]
		if !isBlockValue(val) {
			hasAttributes = true
			break
		}
	}

	// Process attributes first
	for _, key := range keys {
		val := data[key]
		if isBlockValue(val) {
			continue
		} // Skip blocks for now

		if firstItemWritten {
			buf.WriteString("\n") // Add newline *before* this attribute if not the first item
		}
		attrStr, err := attributeValueToTFString(key, val, blockType, indent)
		if err != nil {
			return errors.Errorf("error formatting attribute %q: %w", key, err)
		}
		if attrStr != "" {
			buf.WriteString(indent + attrStr) // Write attribute
			firstItemWritten = true
			lastItemType = "attribute"
		}
	}

	// Process blocks next
	for _, key := range keys {
		val := data[key]
		if !isBlockValue(val) {
			continue
		} // Skip attributes

		switch v := val.(type) {
		case map[string]any: // Single block
			if firstItemWritten {
				// Add TWO newlines if the last item was a block OR if it was an attribute AND we have attributes
				if lastItemType == "block" || hasAttributes {
					buf.WriteString("\n\n")
				} else { // No attributes before this block
					// Don't add newline if this is the very first item
				}
			}
			err := blockMapToTFString(key, v, indent, buf)
			if err != nil {
				return err
			}
			firstItemWritten = true
			lastItemType = "block"
		case []any: // List of blocks
			for i, itemMap := range v {
				item, ok := itemMap.(map[string]any)
				if !ok {
					return errors.Errorf("non-map item at index %d in block list %q", i, key)
				}
				if firstItemWritten {
					// Add TWO newlines if last was block, or if last was attribute and this is the first block, or between blocks in a list
					if lastItemType == "block" || (hasAttributes && i == 0) || i > 0 {
						buf.WriteString("\n\n")
					} else { // First item, no preceding attributes
						// Don't add newline before the very first item
					}
				}
				err := blockMapToTFString(key, item, indent, buf)
				if err != nil {
					return err
				}
				firstItemWritten = true
				lastItemType = "block" // Mark after processing each block
			}
		}
	}
	// No final newline added here
	return nil
}

// isBlockValue checks if a value represents a block (map) or list of blocks ([]map).
func isBlockValue(val any) bool {
	switch v := val.(type) {
	case map[string]any:
		// Consider empty maps as potential empty blocks, not attributes
		return true
	case []any:
		if len(v) > 0 {
			// Check if ALL items are maps, not just the first
			for _, item := range v {
				if _, ok := item.(map[string]any); !ok {
					return false // Found a non-map, treat as attribute list
				}
			}
			return true // All items are maps
		}
		// Consider empty lists as attribute lists "[]", not blocks
		return false
	default:
		return false
	}
}

// blockMapToTFString formats a single block from its type and map data.
func blockMapToTFString(blockType string, data map[string]any, indent string, buf *bytes.Buffer) error {
	labels := []string{}
	currentMap := data
	originalBlockType := blockType

	// Extract labels by unwrapping nested single-key maps
	for {
		if len(currentMap) != 1 {
			break
		}
		var key string
		var val any
		for k, v := range currentMap {
			key, val = k, v
		}
		if nextMap, ok := val.(map[string]any); ok {
			labels = append(labels, key)
			currentMap = nextMap
		} else {
			break
		}
	}

	// Sanitize the block type name itself
	sanitizedBlockType := sanitizeIdentifier(blockType)

	// Write block header with sanitized type
	buf.WriteString(indent + sanitizedBlockType)
	for _, label := range labels {
		buf.WriteString(" " + strconv.Quote(label))
	} // Always quote labels
	buf.WriteString(" {") // No newline before opening brace

	// Recursively write the block body
	if len(currentMap) > 0 {
		buf.WriteString("\n") // Add newline only if body has content
		// Pass the ORIGINAL block type for context in recursive calls
		err := mapToTFStringRecursive(currentMap, originalBlockType, indent+indentUnit, buf)
		if err != nil {
			return err
		}
		// Ensure the last item in the recursive call adds a newline before the closing brace indent
		if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != '\n' {
			buf.WriteString("\n") // Add newline if recursive call didn't end with one
		}
		buf.WriteString(indent) // Add indent for closing brace
	}

	buf.WriteString("}") // Closing brace (might be on same line if body empty)
	// No trailing newline here
	return nil
}

// attributeValueToTFString formats a key-value pair for HCL, applying exception rules.
func attributeValueToTFString(key string, val any, blockType string, indent string) (string, error) {
	parseAsExpr := false
	isException := false
	formattedVal := ""
	var err error

	// Apply reverse exception logic based on blockType and key
	switch blockType {
	case "resource", "data":
		if key == "provider" {
			parseAsExpr = true
			isException = true
		}
		if key == "depends_on" {
			// List of refs needs parseAsExpr=true for items
			formattedVal, err = formatTFValue(val, true, indent)
			isException = true
		}
	case "lifecycle":
		if key == "ignore_changes" {
			if s, ok := val.(string); ok && s == "all" {
				formattedVal = "all"
				parseAsExpr = true
			} else {
				// List of refs needs parseAsExpr=true for items
				formattedVal, err = formatTFValue(val, true, indent)
			}
			isException = true
		}
	case "connection":
		if key == "type" {
			parseAsExpr = false
			isException = true
		} // Literal string
	case "variable":
		if key == "type" { // Special: Raw string value
			if s, ok := val.(string); ok {
				formattedVal = s
			} else {
				err = errors.Errorf("variable type value is not a string: %T", val)
			}
			isException = true
		}
		// default, description are literals (default parseAsExpr=false is correct)
	case "module":
		if key == "source" || key == "version" {
			parseAsExpr = false
			isException = true
		} // Literals
		if key == "providers" {
			formattedVal, err = formatProvidersMap(val, indent)
			isException = true
		}
	case "provider":
		if key == "alias" || key == "version" {
			parseAsExpr = false
			isException = true
		} // Literals
	case "required_providers": // Attributes *within* the provider type block (e.g., inside 'aws = { ... }')
		formattedVal, err = formatRequiredProviderConfig(val, indent)
		isException = true
	}
	if err != nil {
		return "", err
	}

	// Default formatting if no exception handled it or if exception needs default formatting
	if !isException || formattedVal == "" {
		if blockType == "output" && key == "value" {
			parseAsExpr = true
		} // Output value is an expression
		if blockType == "" && key == "locals" { // Top-level locals block
			formattedVal, err = formatLocalsMap(val, indent)
			if err != nil {
				return "", err
			}
		} else {
			// Use the determined parseAsExpr hint for default formatting
			formattedVal, err = formatTFValue(val, parseAsExpr, indent)
			if err != nil {
				return "", err
			}
		}
	}

	// Use sanitized key for HCL output attribute keys
	keyStr := sanitizeIdentifier(key)
	// No quoting needed after sanitization for attribute keys

	return fmt.Sprintf("%s = %s", keyStr, formattedVal), nil
}

// formatTFValue formats a Go value into its Terraform HCL string representation.
func formatTFValue(v any, parseAsExpr bool, indent string) (string, error) {
	switch val := v.(type) {
	case nil:
		return "null", nil
	case string:
		isExpressionString := len(val) > 3 && strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}")
		if isExpressionString {
			return unwrapString(val), nil
		} // Rule 1: Unwrap ${...}
		// Rule 2: Handle explicitly requested expression/reference formatting
		// Allow '.' for traversals, '[' for index, '(' for function calls
		if parseAsExpr && (isValidIdentifier(val) || strings.ContainsAny(val, ".[]()")) {
			return val, nil
		}
		return strconv.Quote(val), nil // Rule 3: Default quote
	case int, int64, float64:
		if f64, ok := val.(float64); ok {
			return strconv.FormatFloat(f64, 'g', -1, 64), nil
		}
		return fmt.Sprintf("%v", val), nil
	case bool:
		return strconv.FormatBool(val), nil
	case []any: // Format as Tuple [...]
		items := make([]string, len(val))
		currentItemIndent := indent + indentUnit
		for i, item := range val {
			itemParseAsExpr := false
			if s, ok := item.(string); ok {
				itemIsExpressionString := len(s) > 3 && strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}")
				// Propagate parseAsExpr hint for list items (e.g., in depends_on)
				itemIsLikelyRef := parseAsExpr && (isValidIdentifier(s) || strings.ContainsAny(s, ".[]()"))
				itemParseAsExpr = itemIsExpressionString || itemIsLikelyRef
			}
			itemStr, err := formatTFValue(item, itemParseAsExpr, currentItemIndent)
			if err != nil {
				return "", err
			}
			items[i] = itemStr
		}
		if len(items) == 0 {
			return "[]", nil
		}
		listContent := strings.Join(items, ",\n"+currentItemIndent)
		// Use parent indent for closing bracket
		return "[\n" + currentItemIndent + listContent + "\n" + indent + "]", nil
	case map[string]any: // Format as Object { ... }
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		pairs := make([]string, len(keys))
		currentItemIndent := indent + indentUnit
		for i, k := range keys {
			itemVal := val[k]
			// Object keys use quoting, not sanitization
			keyStr := k
			if !isValidIdentifier(k) {
				keyStr = strconv.Quote(k)
			}
			valParseAsExpr := false
			if s, ok := itemVal.(string); ok {
				valIsExpressionString := len(s) > 3 && strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}")
				// Propagate parseAsExpr hint for map values (e.g., in locals)
				valIsLikelyRef := parseAsExpr && (isValidIdentifier(s) || strings.ContainsAny(s, ".[]()"))
				valParseAsExpr = valIsExpressionString || valIsLikelyRef
			}
			valStr, err := formatTFValue(itemVal, valParseAsExpr, currentItemIndent)
			if err != nil {
				return "", err
			}
			pairs[i] = fmt.Sprintf("%s = %s", keyStr, valStr)
		}
		if len(pairs) == 0 {
			return "{}", nil
		}
		mapContent := strings.Join(pairs, "\n"+currentItemIndent)
		// Use parent indent for closing bracket
		return "{\n" + currentItemIndent + mapContent + "\n" + indent + "}", nil
	default:
		return "", errors.Errorf("unsupported type in formatTFValue: %T", v)
	}
}

// formatProvidersMap formats the 'providers' map in a module block.
func formatProvidersMap(v any, indent string) (string, error) {
	providerMap, ok := v.(map[string]any)
	if !ok {
		return "", errors.Errorf("expected a map for providers, got %T", v)
	}
	keys := make([]string, 0, len(providerMap))
	for k := range providerMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, len(keys))
	currentItemIndent := indent + indentUnit
	for i, k := range keys {
		valStr, ok := providerMap[k].(string)
		if !ok {
			return "", errors.Errorf("provider value for %q is not a string reference", k)
		}
		// Object keys use quoting, not sanitization
		keyStr := k
		if !isValidIdentifier(k) {
			keyStr = strconv.Quote(k)
		}
		// Value must be a reference, output directly (allow .)
		// Use isValidIdentifierOrTraversal
		if isValidIdentifierOrTraversal(valStr) {
			pairs[i] = fmt.Sprintf("%s = %s", keyStr, valStr)
		} else {
			return "", errors.Errorf("provider reference %q for key %q is not valid", valStr, k)
		}
	}
	if len(pairs) == 0 {
		return "{}", nil
	}
	mapContent := strings.Join(pairs, "\n"+currentItemIndent)
	return "{\n" + currentItemIndent + mapContent + "\n" + indent + "}", nil
}

// formatLocalsMap formats the 'locals' block map.
func formatLocalsMap(v any, indent string) (string, error) {
	localsMap, ok := v.(map[string]any)
	if !ok {
		return "", errors.Errorf("expected a map for locals, got %T", v)
	}
	keys := make([]string, 0, len(localsMap))
	for k := range localsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, len(keys))
	currentItemIndent := indent + indentUnit
	for i, k := range keys {
		itemVal := localsMap[k]
		// Object keys use quoting, not sanitization
		keyStr := k
		if !isValidIdentifier(k) {
			keyStr = strconv.Quote(k)
		}
		// Local values ARE expressions
		valStr, err := formatTFValue(itemVal, true, currentItemIndent) // Hint value is expr
		if err != nil {
			return "", errors.Errorf("error formatting local value for %q: %w", k, err)
		}
		pairs[i] = fmt.Sprintf("%s = %s", keyStr, valStr)
	}
	if len(pairs) == 0 {
		return "{}", nil
	}
	mapContent := strings.Join(pairs, "\n"+currentItemIndent)
	return "{\n" + currentItemIndent + mapContent + "\n" + indent + "}", nil
}

// formatRequiredProviderConfig formats the value of a required_providers attribute (like 'aws').
func formatRequiredProviderConfig(v any, indent string) (string, error) {
	configMap, ok := v.(map[string]any)
	if !ok {
		return "", errors.Errorf("expected a map for required provider config, got %T", v)
	}
	keys := make([]string, 0, len(configMap))
	for k := range configMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, len(keys))
	currentItemIndent := indent + indentUnit
	for i, k := range keys {
		itemVal := configMap[k]
		// Object keys use quoting, not sanitization
		keyStr := k
		if !isValidIdentifier(k) {
			keyStr = strconv.Quote(k)
		}
		var valStr string
		var err error
		if k == "configuration_aliases" { // List of refs
			valStr, err = formatTFValue(itemVal, true, currentItemIndent) // Hint list items are refs
		} else if k == "source" || k == "version" { // Literals
			valStr, err = formatTFValue(itemVal, false, currentItemIndent)
		} else { // Default literal
			valStr, err = formatTFValue(itemVal, false, currentItemIndent)
		}
		if err != nil {
			return "", errors.Errorf("error formatting required_provider attribute %q: %w", k, err)
		}
		pairs[i] = fmt.Sprintf("%s = %s", keyStr, valStr)
	}
	if len(pairs) == 0 {
		return "{}", nil
	}
	mapContent := strings.Join(pairs, "\n"+currentItemIndent)
	return "{\n" + currentItemIndent + mapContent + "\n" + indent + "}", nil
}

// sanitizeIdentifier makes a string safe to use as an HCL identifier (attribute key or block type).
// Replaces invalid characters with underscores and ensures it doesn't start with a digit or hyphen.
func sanitizeIdentifier(s string) string {
	if s == "" {
		return "_" // Handle empty string case
	}

	var builder strings.Builder
	firstChar := rune(s[0])

	// Ensure the first character is not a digit or hyphen
	if unicode.IsDigit(firstChar) || firstChar == '-' {
		builder.WriteRune('_')
	} else if unicode.IsLetter(firstChar) || firstChar == '_' {
		builder.WriteRune(firstChar)
	} else {
		builder.WriteRune('_') // Replace other invalid start char
	}

	// Process remaining characters: letters, digits, underscore, hyphen
	for _, r := range s[1:] {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_') // Replace invalid char
		}
	}

	return builder.String()
}

// isValidIdentifier checks if a string is a valid HCL identifier
// (letters, digits, _, - starting with letter or _). Used for value formatting.
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	firstChar := rune(s[0])
	// Correct check for start character (letter or underscore)
	if !unicode.IsLetter(firstChar) && firstChar != '_' {
		return false // Must start with letter or underscore
	}
	for _, r := range s[1:] {
		// Check allowed characters (letter, digit, underscore, hyphen)
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return false // Invalid character
		}
	}
	return true
}

// isValidIdentifierOrTraversal checks if a string is a valid identifier or
// looks like a traversal (contains .). Used for reference values.
func isValidIdentifierOrTraversal(s string) bool {
	if s == "" {
		return false
	}
	// Allow starting with letter or _
	firstChar := rune(s[0])
	if !unicode.IsLetter(firstChar) && firstChar != '_' {
		return false
	}
	// Allow letters, digits, _, -, and . (for traversal)
	// Include '[' and ']' for index access
	for _, r := range s { // Check all characters now
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' && r != '.' && r != '[' && r != ']' {
			// Disallow other special chars like (), spaces etc. in basic traversals/refs
			return false
		}
	}
	// Basic check passes
	return true
}
