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
	"fmt"
	"github.com/go-errors/errors"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	"net/url"
	"slices"
	"sort"
	"strings"
)

// AuthorizationServer is an interface representing a generic authorization server.
type AuthorizationServer interface {
	Type() string
}

// AuthServerInfo holds the consolidated information for a single OAuth2 authorization server.
type AuthServerInfo struct {
	AuthType         string   `yaml:"type"`
	IssuerURL        string   `yaml:"issuer_url"`
	AuthorizationURL string   `yaml:"authorization_url,omitempty"`
	TokenURL         string   `yaml:"token_url,omitempty"`
	Scopes           []string `yaml:"scopes"`
	Flows            []string `yaml:"flows"`
}

func (a *AuthServerInfo) Type() string { return a.AuthType }

// OIDCAuthServerInfo holds the consolidated information for a single OpenID Connect server.
type OIDCAuthServerInfo struct {
	AuthType         string `yaml:"type"`
	IssuerURL        string `yaml:"issuer_url"`
	OpenIdConnectUrl string `yaml:"openIdConnectUrl"`
}

func (o *OIDCAuthServerInfo) Type() string { return o.AuthType }

// SecurityRequirement holds a security scheme name and the scopes required.
type SecurityRequirement struct {
	Name   string
	Scopes []string
}

// SelectAuthorizationServer finds the authorization server corresponding to the highest priority security requirement.
// The selection is prioritized:
// 1. A requirement defined at the top-level of the document.
// 2. The requirement that is used most frequently across all operations.
func SelectAuthorizationServer(root *yaml.Node) (AuthorizationServer, error) {
	var err error
	var selectedSchemeName string
	var selectedSchemeType string

	var topLevelSecurityNodes []*yaml.Node
	// --- Priority 1: Check for top-level security requirements ---
	if topLevelSecurityNodes, err = findNodesByPath(root, "$.security"); err != nil {
		return nil, errors.Errorf("error finding top-level security node: %w", err)
	}

	if len(topLevelSecurityNodes) > 0 {
		var topLevelReqs []SecurityRequirement
		if topLevelReqs, err = parseSecurityNode(topLevelSecurityNodes[0]); err != nil {
			return nil, errors.Errorf("error parsing top-level security node: %w", err)
		}
		// Find the first requirement that is of a supported type.
		for _, req := range topLevelReqs {
			var schemeType string
			if schemeType, err = getSchemeType(req.Name, root); err != nil {
				return nil, err // Propagate error
			}
			if schemeType == "" {
				return nil, errors.Errorf("security requirement '%s' references an undefined or invalid security scheme", req.Name)
			}

			if schemeType == "oauth2" || schemeType == "openIdConnect" {
				// This is the highest priority, so select it and stop searching.
				selectedSchemeName = req.Name
				selectedSchemeType = schemeType
				break
			}
		}
	}

	// --- Priority 2: Find the most frequently used requirement in operations ---
	if selectedSchemeName == "" {
		var operationSecurityNodes []*yaml.Node
		if operationSecurityNodes, err = findNodesByPath(root, "$.paths..security"); err != nil {
			return nil, errors.Errorf("error finding operation-level security nodes: %w", err)
		}

		schemeCounts := make(map[string]int)
		var mostFrequentScheme, mostFrequentType string
		var mostFrequentCount int

		for _, securityNode := range operationSecurityNodes {
			var opReqs []SecurityRequirement
			if opReqs, err = parseSecurityNode(securityNode); err != nil {
				return nil, errors.Errorf("error parsing operation security node: %w", err)
			}

			for _, req := range opReqs {
				var schemeType string
				if schemeType, err = getSchemeType(req.Name, root); err != nil {
					return nil, err
				}
				if schemeType == "" {
					return nil, errors.Errorf("security requirement '%s' references an undefined or invalid security scheme", req.Name)
				}

				if schemeType == "oauth2" || schemeType == "openIdConnect" {
					schemeCounts[req.Name]++
					currentCount := schemeCounts[req.Name]

					if currentCount > mostFrequentCount {
						mostFrequentCount = currentCount
						mostFrequentScheme = req.Name
						mostFrequentType = schemeType
					}
				}
			}
		}

		if mostFrequentScheme != "" {
			selectedSchemeName = mostFrequentScheme
			selectedSchemeType = mostFrequentType
		}
	}

	if selectedSchemeName == "" {
		// No suitable OAuth2 or OpenID Connect requirement found anywhere.
		return nil, nil
	}

	// Now that we have the scheme name and type, build the appropriate auth server info.
	switch selectedSchemeType {
	case "oauth2":
		return buildOAuth2AuthServerForScheme(selectedSchemeName, root)
	case "openIdConnect":
		return buildOIDCAuthServerForScheme(selectedSchemeName, root)
	default:
		// This should be unreachable given the checks above.
		return nil, errors.Errorf("unsupported security scheme type selected: '%s'", selectedSchemeType)
	}
}

// buildOAuth2AuthServerForScheme creates an AuthServerInfo object directly from a security scheme definition.
func buildOAuth2AuthServerForScheme(schemeName string, root *yaml.Node) (*AuthServerInfo, error) {
	path := fmt.Sprintf("$.components.securitySchemes['%s']", schemeName)
	var schemeNodes []*yaml.Node
	var err error
	if schemeNodes, err = findNodesByPath(root, path); err != nil {
		return nil, errors.Errorf("error finding security scheme '%s': %w", schemeName, err)
	}
	if len(schemeNodes) == 0 {
		return nil, errors.Errorf("security scheme '%s' not found in components", schemeName)
	}

	schemeNode := schemeNodes[0]
	var flowsNode *yaml.Node
	if flowsNode, err = findNodeByKey(schemeNode, "flows"); err != nil {
		return nil, errors.Errorf("error finding 'flows' for security scheme '%s': %w", schemeName, err)
	}
	if flowsNode == nil || len(flowsNode.Content) == 0 {
		return nil, errors.Errorf("no flows found for security scheme '%s'", schemeName)
	}

	var allScopes, allFlows []string
	var issuerURL, authorizationURL, tokenURL string

	// Iterate through each flow to collect scopes, flow names, and URLs.
	for j := 0; j < len(flowsNode.Content); j += 2 {
		flowKeyNode := flowsNode.Content[j]
		flowNode := flowsNode.Content[j+1]

		allFlows = appendIfMissing(allFlows, flowKeyNode.Value)

		// Collect all unique scopes from all flows.
		var scopesNode *yaml.Node
		if scopesNode, err = findNodeByKey(flowNode, "scopes"); err != nil {
			return nil, errors.Errorf("error finding 'scopes' in flow for security scheme '%s': %w", schemeName, err)
		}
		if scopesNode != nil {
			for k := 0; k < len(scopesNode.Content); k += 2 {
				scopeKeyNode := scopesNode.Content[k]
				allScopes = appendIfMissing(allScopes, scopeKeyNode.Value)
			}
		}

		// The authorization and token URLs should be the same for all flows within a scheme.
		// We capture them once.
		if authorizationURL == "" {
			var authURLNode *yaml.Node
			if authURLNode, err = findNodeByKey(flowNode, "authorizationUrl"); err != nil {
				return nil, errors.Errorf("error finding 'authorizationUrl' in flow for security scheme '%s': %w", schemeName, err)
			}
			if authURLNode != nil {
				authorizationURL = authURLNode.Value
			}
		}
		if tokenURL == "" {
			var tokenURLNode *yaml.Node
			if tokenURLNode, err = findNodeByKey(flowNode, "tokenUrl"); err != nil {
				return nil, errors.Errorf("error finding 'tokenUrl' in flow for security scheme '%s': %w", schemeName, err)
			}
			if tokenURLNode != nil {
				tokenURL = tokenURLNode.Value
			}
		}
	}

	// Infer the issuer URL from the authorization URL only.
	if authorizationURL != "" {
		var inferredURL string
		if inferredURL, err = inferIssuerURL(authorizationURL); err != nil {
			return nil, errors.Errorf("could not infer issuer URL for security scheme '%s': %w", schemeName, err)
		}
		issuerURL = inferredURL
	}

	sort.Strings(allScopes)
	sort.Strings(allFlows)

	return &AuthServerInfo{
		AuthType:         "oauth2",
		IssuerURL:        issuerURL,
		AuthorizationURL: authorizationURL,
		TokenURL:         tokenURL,
		Scopes:           allScopes,
		Flows:            allFlows,
	}, nil
}

// buildOIDCAuthServerForScheme creates an OIDCAuthServerInfo object from a security scheme definition.
func buildOIDCAuthServerForScheme(schemeName string, root *yaml.Node) (*OIDCAuthServerInfo, error) {
	path := fmt.Sprintf("$.components.securitySchemes['%s']", schemeName)
	var schemeNodes []*yaml.Node
	var err error
	if schemeNodes, err = findNodesByPath(root, path); err != nil {
		return nil, errors.Errorf("error finding security scheme '%s': %w", schemeName, err)
	}
	if len(schemeNodes) == 0 {
		return nil, errors.Errorf("security scheme '%s' not found in components", schemeName)
	}

	schemeNode := schemeNodes[0]
	var oidcURLNode *yaml.Node
	if oidcURLNode, err = findNodeByKey(schemeNode, "openIdConnectUrl"); err != nil {
		return nil, errors.Errorf("error finding 'openIdConnectUrl' for security scheme '%s': %w", schemeName, err)
	}
	if oidcURLNode == nil || oidcURLNode.Value == "" {
		return nil, errors.Errorf("no openIdConnectUrl found for security scheme '%s'", schemeName)
	}

	oidcURL := oidcURLNode.Value
	var issuerURL string
	if issuerURL, err = inferIssuerURLFromOIDC(oidcURL); err != nil {
		return nil, errors.Errorf("could not infer issuer URL for security scheme '%s': %w", schemeName, err)
	}

	return &OIDCAuthServerInfo{
		AuthType:         "openIdConnect",
		IssuerURL:        issuerURL,
		OpenIdConnectUrl: oidcURL,
	}, nil
}

// getSchemeType gets the type of security scheme.
func getSchemeType(schemeName string, root *yaml.Node) (string, error) {
	jsonPath := fmt.Sprintf("$.components.securitySchemes['%s']", schemeName)
	var schemeNodes []*yaml.Node
	var err error
	if schemeNodes, err = findNodesByPath(root, jsonPath); err != nil {
		return "", errors.Errorf("error finding security scheme '%s': %w", schemeName, err)
	}

	if len(schemeNodes) == 0 {
		// If the scheme is not defined, return an empty string. The caller must handle it.
		return "", nil
	}

	schemeNode := schemeNodes[0]
	var typeNode *yaml.Node
	if typeNode, err = findNodeByKey(schemeNode, "type"); err != nil {
		return "", errors.Errorf("error finding 'type' for security scheme '%s': %w", schemeName, err)
	}
	if typeNode == nil {
		// A scheme without a type is invalid. Return an empty string. The caller must handle it.
		return "", nil
	}
	return typeNode.Value, nil
}

// parseSecurityNode processes a 'security' YAML node and extracts the requirements.
func parseSecurityNode(node *yaml.Node) ([]SecurityRequirement, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, errors.Errorf("expected security node to be a sequence (array), but got %v", node.Kind)
	}

	requirements := make([]SecurityRequirement, 0)
	// Each item in the sequence is a map, e.g., { "oAuth2": ["scope1"] }
	for _, reqObjectNode := range node.Content {
		if reqObjectNode.Kind != yaml.MappingNode {
			// Skip any non-map items in the sequence.
			continue
		}

		// A security requirement object is a map where the key is the scheme name
		// and the value is a list of required scopes.
		for i := 0; i < len(reqObjectNode.Content); i += 2 {
			schemeNameNode := reqObjectNode.Content[i]
			scopesNode := reqObjectNode.Content[i+1]

			if scopesNode.Kind != yaml.SequenceNode {
				// For openIdConnect, the value is an empty array.
				if scopesNode.Kind == yaml.ScalarNode && scopesNode.Value == "" {
					requirements = append(requirements, SecurityRequirement{
						Name:   schemeNameNode.Value,
						Scopes: []string{},
					})
					continue
				}
				return nil, errors.Errorf("expected scopes for scheme '%s' to be a sequence, but got %v", schemeNameNode.Value, scopesNode.Kind)
			}

			scopes := make([]string, 0)
			for _, scopeNode := range scopesNode.Content {
				scopes = append(scopes, scopeNode.Value)
			}

			requirements = append(requirements, SecurityRequirement{
				Name:   schemeNameNode.Value,
				Scopes: scopes,
			})
		}
	}
	return requirements, nil
}

// findNodesByPath uses yamlpath to find all matching nodes.
func findNodesByPath(node *yaml.Node, pathStr string) ([]*yaml.Node, error) {
	var path *yamlpath.Path
	var err error
	if path, err = yamlpath.NewPath(pathStr); err != nil {
		return nil, errors.Errorf("failed to create yamlpath for '%s': %w", pathStr, err)
	}
	var nodes []*yaml.Node
	if nodes, err = path.Find(node); err != nil {
		return nil, errors.Errorf("error finding nodes with path '%s': %w", pathStr, err)
	}
	return nodes, nil
}

// findNodeByKey is a helper to find a map value by its key using yamlpath.
func findNodeByKey(node *yaml.Node, key string) (*yaml.Node, error) {
	// Construct a yamlpath expression to select the child by key.
	pathStr := fmt.Sprintf("$.%s", key)

	var nodes []*yaml.Node
	var err error
	if nodes, err = findNodesByPath(node, pathStr); err != nil {
		// Propagate the error from the underlying find operation.
		return nil, err
	}
	if len(nodes) == 0 {
		// No node found, but not an error.
		return nil, nil
	}

	// Return the first matching node.
	return nodes[0], nil
}

// appendIfMissing adds a string to a slice only if it's not already present.
func appendIfMissing(slice []string, i string) []string {
	if slices.Index(slice, i) == -1 {
		return append(slice, i)
	}
	return slice
}

// inferIssuerURL infers the issuer URL based on the authorization URL.
// If the authURL follows a Keycloak-style pattern (e.g., ".../realms/{realm}/...") or
// an Okta-style pattern (e.g., ".../oauth2/{serverId}/..."), it includes the dynamic
// segment in the issuer URL, provided the pattern appears at the root of the path
// and the dynamic segment is not 'auth' or 'authorize'.
// Otherwise, it uses just the scheme and host.
func inferIssuerURL(authURL string) (string, error) {
	var pAuth *url.URL
	var err error
	if pAuth, err = url.Parse(authURL); err != nil {
		return "", errors.Errorf("failed to parse authorization URL '%s': %w", authURL, err)
	}

	pathSegments := strings.Split(strings.Trim(pAuth.Path, "/"), "/")

	// Check if there are enough path segments for the patterns.
	if len(pathSegments) >= 2 {
		// Check for Keycloak-style "/realms/" path at the root.
		if pathSegments[0] == "realms" {
			identifier := pathSegments[1]
			// Reconstruct the issuer URL up to and including the realm.
			return fmt.Sprintf("%s://%s/realms/%s", pAuth.Scheme, pAuth.Host, identifier), nil
		}

		// Check for Okta-style "/oauth2/" path at the root.
		if pathSegments[0] == "oauth2" {
			identifier := pathSegments[1]
			if identifier != "authorize" && identifier != "auth" {
				// Reconstruct the issuer URL up to and including the server ID.
				return fmt.Sprintf("%s://%s/oauth2/%s", pAuth.Scheme, pAuth.Host, identifier), nil
			}
		}
	}

	// Default case: use the scheme and host of the authorization URL.
	return fmt.Sprintf("%s://%s", pAuth.Scheme, pAuth.Host), nil
}

// inferIssuerURLFromOIDC infers the issuer URL from an OpenID Connect discovery URL.
func inferIssuerURLFromOIDC(oidcURL string) (string, error) {
	var parsedURL *url.URL
	var err error
	if parsedURL, err = url.Parse(oidcURL); err != nil {
		return "", errors.Errorf("failed to parse OpenID Connect URL '%s': %w", oidcURL, err)
	}

	// The issuer URL is the base URL of the discovery document, which is located at the /.well-known/openid-configuration path.
	wellKnownPath := "/.well-known/openid-configuration"
	if idx := strings.Index(parsedURL.Path, wellKnownPath); idx != -1 {
		parsedURL.Path = parsedURL.Path[:idx]
	}

	// The resulting issuer URL should not have a trailing slash.
	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/")

	return parsedURL.String(), nil
}
