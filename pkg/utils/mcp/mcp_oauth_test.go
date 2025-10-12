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
	"gopkg.in/yaml.v3"
	"reflect"
	"testing"
)

// TestSelectAuthorizationServer tests the top-level logic for selecting an auth server.
func TestSelectAuthorizationServer(t *testing.T) {
	testCases := []struct {
		name          string
		yamlContent   string
		expectedType  string // "oauth2", "openIdConnect", or "" for nil
		expectedURL   string // IssuerURL or OpenIdConnectUrl depending on type
		expectError   bool
		expectedFlows []string // Only for oauth2
	}{
		{
			name: "Priority 1: Top-level oauth2 requirement",
			yamlContent: `
security:
  - topLevelOauth:
      - scope1
components:
  securitySchemes:
    topLevelOauth:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://example.com/auth
          tokenUrl: https://example.com/token
          scopes:
            scope1: "desc1"
`,
			expectedType:  "oauth2",
			expectedURL:   "https://example.com",
			expectError:   false,
			expectedFlows: []string{"authorizationCode"},
		},
		{
			name: "Priority 1: Top-level openIdConnect requirement",
			yamlContent: `
security:
  - topLevelOidc: []
components:
  securitySchemes:
    topLevelOidc:
      type: openIdConnect
      openIdConnectUrl: https://example.com/oidc/.well-known/openid-configuration
`,
			expectedType: "openIdConnect",
			expectedURL:  "https://example.com/oidc",
			expectError:  false,
		},
		{
			name: "Priority 2: Most frequent oauth2 in operations",
			yamlContent: `
paths:
  /path1:
    get:
      security:
        - opOauth: [scope1]
  /path2:
    post:
      security:
        - opOauth: [scope1]
  /path3:
    get:
      security:
        - anotherScheme: []
components:
  securitySchemes:
    opOauth:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://frequent.com/token
          scopes:
            scope1: "desc1"
    anotherScheme:
      type: apiKey
      in: header
      name: X-API-KEY
`,
			expectedType:  "oauth2",
			expectedURL:   "", // No authorizationUrl to infer from
			expectError:   false,
			expectedFlows: []string{"clientCredentials"},
		},
		{
			name: "No suitable security requirement found",
			yamlContent: `
components:
  securitySchemes:
    apiKeyScheme:
      type: apiKey
      in: header
      name: X-API-KEY
`,
			expectedType: "",
			expectError:  false,
		},
		{
			name:        "Malformed YAML content",
			yamlContent: `security: [{]`,
			expectError: true,
		},
		{
			name: "Error: Security scheme not found in components",
			yamlContent: `
security:
  - missingScheme: []
components:
  securitySchemes:
    someOtherScheme:
      type: oauth2
`,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var root yaml.Node
			err := yaml.Unmarshal([]byte(tc.yamlContent), &root)
			if err != nil {
				if tc.expectError {
					return // Expected parsing error
				}
				t.Fatalf("Failed to unmarshal YAML: %v", err)
			}

			authServer, err := SelectAuthorizationServer(root.Content[0])

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Did not expect an error, but got: %v", err)
			}

			if tc.expectedType == "" {
				if authServer != nil {
					t.Errorf("Expected nil auth server, but got type %s", authServer.Type())
				}
				return
			}

			if authServer == nil {
				t.Fatalf("Expected auth server of type %s, but got nil", tc.expectedType)
			}

			if authServer.Type() != tc.expectedType {
				t.Errorf("Expected type %s, got %s", tc.expectedType, authServer.Type())
			}

			switch srv := authServer.(type) {
			case *AuthServerInfo:
				if srv.IssuerURL != tc.expectedURL {
					t.Errorf("Expected issuer URL '%s', got '%s'", tc.expectedURL, srv.IssuerURL)
				}
				if !reflect.DeepEqual(srv.Flows, tc.expectedFlows) {
					t.Errorf("Expected flows %v, got %v", tc.expectedFlows, srv.Flows)
				}
			case *OIDCAuthServerInfo:
				// For OIDC, we check the original connect URL, which infers the issuer URL
				if srv.IssuerURL != tc.expectedURL {
					t.Errorf("Expected inferred issuer URL '%s', got '%s'", tc.expectedURL, srv.IssuerURL)
				}
			}
		})
	}
}

// TestInferIssuerURL tests the issuer URL inference logic.
func TestInferIssuerURL(t *testing.T) {
	testCases := []struct {
		name        string
		authURL     string
		expectedURL string
		expectError bool
	}{
		{"Keycloak style", "https://keycloak.example.com/realms/my-realm/protocol/openid-connect/auth", "https://keycloak.example.com/realms/my-realm", false},
		{"Okta style", "https://okta.example.com/oauth2/default/v1/authorize", "https://okta.example.com/oauth2/default", false},
		{"Generic style", "https://generic.example.com/oauth/authorize", "https://generic.example.com", false},
		{"Root path", "https://root.example.com/authorize", "https://root.example.com", false},
		{"Malformed URL", "::not-a-url", "", true},
		{"Okta style with 'authorize'", "https://okta.example.com/oauth2/authorize/v1/authorize", "https://okta.example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inferredURL, err := inferIssuerURL(tc.authURL)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
				if inferredURL != tc.expectedURL {
					t.Errorf("Expected URL '%s', got '%s'", tc.expectedURL, inferredURL)
				}
			}
		})
	}
}

// TestInferIssuerURLFromOIDC tests the OIDC issuer URL inference logic.
func TestInferIssuerURLFromOIDC(t *testing.T) {
	testCases := []struct {
		name        string
		oidcURL     string
		expectedURL string
		expectError bool
	}{
		{"Standard OIDC URL", "https://issuer.com/auth/.well-known/openid-configuration", "https://issuer.com/auth", false},
		{"URL with trailing slash", "https://issuer.com/auth//.well-known/openid-configuration", "https://issuer.com/auth", false},
		{"URL at root", "https://issuer.com/.well-known/openid-configuration", "https://issuer.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inferredURL, err := inferIssuerURLFromOIDC(tc.oidcURL)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
				if inferredURL != tc.expectedURL {
					t.Errorf("Expected URL '%s', got '%s'", tc.expectedURL, inferredURL)
				}
			}
		})
	}
}

// TestParseSecurityNode tests the parsing of the security node.
func TestParseSecurityNode(t *testing.T) {
	yamlContent := `
- oAuthScheme:
    - read
    - write
- oidcScheme: []
- anotherScheme:
    - scope:a
`
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal test YAML: %v", err)
	}

	expected := []SecurityRequirement{
		{Name: "oAuthScheme", Scopes: []string{"read", "write"}},
		{Name: "oidcScheme", Scopes: []string{}},
		{Name: "anotherScheme", Scopes: []string{"scope:a"}},
	}

	reqs, err := parseSecurityNode(node.Content[0])
	if err != nil {
		t.Fatalf("parseSecurityNode failed: %v", err)
	}

	if !reflect.DeepEqual(reqs, expected) {
		t.Errorf("Parsed requirements do not match expected.\nGot: %v\nWant: %v", reqs, expected)
	}
}

// TestAppendIfMissing tests the utility function for appending to slices.
func TestAppendIfMissing(t *testing.T) {
	testCases := []struct {
		name     string
		slice    []string
		item     string
		expected []string
	}{
		{"Add to empty", []string{}, "a", []string{"a"}},
		{"Add new item", []string{"a", "b"}, "c", []string{"a", "b", "c"}},
		{"Add existing item", []string{"a", "b"}, "a", []string{"a", "b"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := appendIfMissing(tc.slice, tc.item)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected slice %v, got %v", tc.expected, result)
			}
		})
	}
}
