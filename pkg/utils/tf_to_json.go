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
	"fmt"
	"github.com/go-errors/errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"math/big"
	"strings"
)

// TFFileToMap converts a parsed Terraform hcl.File into a generic map[string]interface{}.
func TFFileToMap(f *hcl.File, fileName string) (map[string]any, error) {
	ctx := hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: map[string]function.Function{},
	}

	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.Errorf("invalid file body type: %T", f.Body)
	}

	result, err := bodyToMap(f, body, &ctx, "") // Root block has no type
	if err != nil {
		return nil, err
	}

	return result, nil
}

// bodyToMap recursively processes the contents of an HCL body into a map.
// blockType informs the function about its context for applying special rules.
func bodyToMap(f *hcl.File, body *hclsyntax.Body, ctx *hcl.EvalContext, blockType string) (map[string]any, error) {
	result := make(map[string]any)

	// Process all attributes (key-value pairs) first.
	for name, attr := range body.Attributes {
		var val any
		var err error
		isException := false // Flag to mark if a special rule was applied

		// This switch handles the block-type-specific exceptions.
		switch blockType {
		case "resource", "data":
			if name == "provider" {
				val = expressionToLiteralString(attr.Expr, f)
				isException = true
			}
			if name == "depends_on" {
				val, err = expressionToRefStringList(attr.Expr, f, ctx)
				if err != nil {
					return nil, errors.Errorf("error processing depends_on: %w", err)
				}
				isException = true
			}
		case "lifecycle":
			if name == "ignore_changes" {
				if trav, ok := attr.Expr.(*hclsyntax.ScopeTraversalExpr); ok && len(trav.Traversal) == 1 {
					if root, ok := trav.Traversal[0].(hcl.TraverseRoot); ok && root.Name == "all" {
						val = "all"
						isException = true
					}
				}
				if !isException {
					val, err = expressionToRefStringList(attr.Expr, f, ctx)
					if err != nil {
						return nil, errors.Errorf("error processing ignore_changes: %w", err)
					}
					isException = true
				}
			}
		case "connection":
			if name == "type" {
				val = expressionToLiteralString(attr.Expr, f)
				isException = true
			}
		case "variable":
			if name == "type" {
				val = fileBytesToString(f, attr.Expr.Range())
				isException = true
			}
		case "module":
			if name == "source" || name == "version" {
				val = expressionToLiteralString(attr.Expr, f)
				isException = true
			}
			if name == "providers" {
				rawMap, mapErr := expressionToMap(attr.Expr, f, ctx)
				if mapErr != nil {
					return nil, errors.Errorf("error processing providers map expr: %w", mapErr)
				}
				if providerMap, ok := rawMap.(map[string]any); ok {
					processedMap := make(map[string]any)
					for k, v := range providerMap {
						if s, ok := v.(string); ok {
							processedMap[k] = unwrapString(s)
						} else {
							processedMap[k] = v
						}
					}
					val = processedMap
				} else {
					val = rawMap
				}
				isException = true
			}
		case "provider":
			if name == "alias" || name == "version" {
				val = expressionToLiteralString(attr.Expr, f)
				isException = true
			}
		case "required_providers":
			val, err = processRequiredProviderAttribute(attr.Expr, f, ctx)
			if err != nil {
				return nil, errors.Errorf("error processing required_provider %q: %w", name, err)
			}
			isException = true
		}

		// If no special rule applied, use the general-purpose expression parser.
		if !isException {
			val, err = expressionToMap(attr.Expr, f, ctx)
			if err != nil {
				return nil, errors.Errorf("error processing attribute %q in block %q: %w", name, blockType, err)
			}
		}
		result[name] = val
	}

	// Process all nested blocks.
	for _, block := range body.Blocks {
		blockBody, err := bodyToMap(f, block.Body, ctx, block.Type)
		if err != nil {
			return nil, errors.Errorf("error processing body of block %q type %q: %w", block.Labels, block.Type, err)
		}

		var contentToAppend any
		if len(block.Labels) > 0 {
			nestedContent := blockBody
			for i := len(block.Labels) - 1; i >= 1; i-- {
				nestedContent = map[string]any{block.Labels[i]: nestedContent}
			}
			contentToAppend = map[string]any{block.Labels[0]: nestedContent}
		} else {
			contentToAppend = blockBody
		}

		key := block.Type
		if existing, exists := result[key]; !exists {
			result[key] = contentToAppend
		} else {
			if list, ok := existing.([]any); ok {
				result[key] = append(list, contentToAppend)
			} else {
				result[key] = []any{existing, contentToAppend}
			}
		}
	}

	return result, nil
}

// fileBytesToString returns the raw string content of an expression from the source file.
func fileBytesToString(f *hcl.File, r hcl.Range) string {
	return string(r.SliceBytes(f.Bytes))
}

// unwrapString removes the "${...}" wrapper from a string, if present.
func unwrapString(s string) string {
	if len(s) > 3 && strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		return s[2 : len(s)-1]
	}
	return s
}

// expressionToRefStringList processes an expression that should be a list of string references.
func expressionToRefStringList(expr hclsyntax.Expression, f *hcl.File, ctx *hcl.EvalContext) ([]any, error) {
	val, err := expressionToMap(expr, f, ctx)
	if err != nil {
		return nil, err
	}

	if list, ok := val.([]any); ok {
		result := make([]any, len(list))
		for i, item := range list {
			if s, ok := item.(string); ok {
				result[i] = unwrapString(s)
			} else {
				result[i] = item
			}
		}
		return result, nil
	} else if s, ok := val.(string); ok {
		return []any{unwrapString(s)}, nil
	} else if val == nil {
		return []any{}, nil
	} else {
		return []any{val}, nil
	}
}

// expressionToLiteralString handles expressions that should be a literal string value or a raw reference string.
func expressionToLiteralString(expr hclsyntax.Expression, f *hcl.File) string {
	if tmpl, ok := expr.(*hclsyntax.TemplateExpr); ok {
		if len(tmpl.Parts) == 1 {
			if lit, ok := tmpl.Parts[0].(*hclsyntax.LiteralValueExpr); ok && lit.Val.Type() == cty.String {
				return lit.Val.AsString()
			}
		}
	}
	return fileBytesToString(f, expr.Range())
}

// getObjectConsKey extracts the string key from an object item's key expression.
func getObjectConsKey(keyExpr hclsyntax.Expression, f *hcl.File, ctx *hcl.EvalContext) (string, error) {
	if ocke, ok := keyExpr.(*hclsyntax.ObjectConsKeyExpr); ok {
		if !ocke.ForceNonLiteral {
			if trav, ok := ocke.Wrapped.(*hclsyntax.ScopeTraversalExpr); ok && len(trav.Traversal) == 1 {
				if root, ok := trav.Traversal[0].(hcl.TraverseRoot); ok {
					return root.Name, nil
				}
			}
		}
		keyVal, err := expressionToMap(ocke.Wrapped, f, ctx)
		if err != nil {
			return "", err
		}
		if s, ok := keyVal.(string); ok {
			return s, nil
		}
		return "", errors.Errorf("object key expression did not evaluate to a string: %v", keyVal)
	} else {
		keyVal, err := expressionToMap(keyExpr, f, ctx)
		if err != nil {
			return "", err
		}
		if s, ok := keyVal.(string); ok {
			return s, nil
		}
		return "", errors.Errorf("dynamic object key is not a string: %v", keyVal)
	}
}

// processRequiredProviderAttribute handles the special object attributes inside a required_providers block.
func processRequiredProviderAttribute(expr hclsyntax.Expression, f *hcl.File, ctx *hcl.EvalContext) (any, error) {
	objExpr, ok := expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return expressionToMap(expr, f, ctx)
	} // Fallback if not an object

	objMap := make(map[string]any)
	for _, item := range objExpr.Items {
		key, err := getObjectConsKey(item.KeyExpr, f, ctx)
		if err != nil {
			return nil, err
		}

		var itemVal any
		var itemErr error
		if key == "configuration_aliases" {
			itemVal, itemErr = expressionToRefStringList(item.ValueExpr, f, ctx)
		} else if key == "source" || key == "version" {
			itemVal = expressionToLiteralString(item.ValueExpr, f)
		} else {
			itemVal, itemErr = expressionToMap(item.ValueExpr, f, ctx)
		}
		if itemErr != nil {
			return nil, errors.Errorf("error processing attribute %q in required_provider: %w", key, itemErr)
		}
		objMap[key] = itemVal
	}
	return objMap, nil
}

// expressionToMap converts an hclsyntax.Expression into a native Go type suitable for JSON encoding.
func expressionToMap(expr hclsyntax.Expression, f *hcl.File, ctx *hcl.EvalContext) (any, error) {
	switch e := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		val := e.Val
		if !val.IsKnown() || val.IsNull() {
			return nil, nil
		}
		switch val.Type() {
		case cty.String:
			return val.AsString(), nil
		case cty.Number:
			bf := val.AsBigFloat()
			if bf.IsInt() {
				i, acc := bf.Int64()
				if acc == big.Exact {
					return i, nil
				}
			}
			f64, acc := bf.Float64()
			if acc == big.Exact {
				return f64, nil
			}
			return bf.Text('f', -1), nil // Fallback to string representation
		case cty.Bool:
			return val.True(), nil
		default:
			return ctyValueToGo(e.Val), nil // Fallback using helper
		}
	case *hclsyntax.TemplateExpr:
		if len(e.Parts) == 1 {
			if lit, ok := e.Parts[0].(*hclsyntax.LiteralValueExpr); ok && lit.Val.Type() == cty.String {
				return lit.Val.AsString(), nil
			}
		}
		var result strings.Builder
		for _, part := range e.Parts {
			if lit, ok := part.(*hclsyntax.LiteralValueExpr); ok && lit.Val.Type() == cty.String {
				result.WriteString(lit.Val.AsString())
			} else {
				partStr := fileBytesToString(f, part.Range())
				result.WriteString(fmt.Sprintf("${%s}", partStr))
			}
		}
		return result.String(), nil
	case *hclsyntax.TupleConsExpr:
		list := make([]any, len(e.Exprs))
		for i, itemExpr := range e.Exprs {
			itemVal, err := expressionToMap(itemExpr, f, ctx)
			if err != nil {
				return nil, err
			}
			list[i] = itemVal
		}
		return list, nil
	case *hclsyntax.ObjectConsExpr:
		objMap := make(map[string]any)
		for _, item := range e.Items {
			key, err := getObjectConsKey(item.KeyExpr, f, ctx)
			if err != nil {
				return nil, err
			}
			val, err := expressionToMap(item.ValueExpr, f, ctx)
			if err != nil {
				return nil, err
			}
			objMap[key] = val
		}
		return objMap, nil
	case *hclsyntax.ScopeTraversalExpr, *hclsyntax.FunctionCallExpr, *hclsyntax.ConditionalExpr,
		*hclsyntax.IndexExpr, *hclsyntax.RelativeTraversalExpr, *hclsyntax.SplatExpr,
		*hclsyntax.ForExpr, *hclsyntax.AnonSymbolExpr, *hclsyntax.TemplateWrapExpr,
		*hclsyntax.TemplateJoinExpr, *hclsyntax.ExprSyntaxError:
		exprString := fileBytesToString(f, e.Range())
		return fmt.Sprintf("${%s}", exprString), nil
	case *hclsyntax.ParenthesesExpr:
		return expressionToMap(e.Expression, f, ctx)
	case *hclsyntax.ObjectConsKeyExpr:
		return expressionToMap(e.Wrapped, f, ctx) // Unwrap if used outside object construction
	default:
		exprString := fileBytesToString(f, e.Range())
		return fmt.Sprintf("${%s}", exprString), nil // Fallback wrap
	}
}

// ctyValueToGo provides a basic conversion from cty.Value to Go types for literals.
func ctyValueToGo(val cty.Value) any {
	if !val.IsKnown() || val.IsNull() {
		return nil
	}
	switch {
	case val.Type() == cty.String:
		return val.AsString()
	case val.Type() == cty.Number:
		f := val.AsBigFloat()
		if f.IsInt() {
			i, _ := f.Int64()
			return int(i)
		} // Simplified
		f64, _ := f.Float64()
		return f64
	case val.Type() == cty.Bool:
		return val.True()
	}
	return nil // Fallback
}
