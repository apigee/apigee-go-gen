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

package git_test

import (
	"github.com/apigee/apigee-go-gen/pkg/git"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseURI(t *testing.T) {
	tests := []struct {
		name         string
		uri          string
		wantRepo     string
		wantRef      string
		wantResource string
		wantErr      bool
	}{
		// --- Success Cases (Web-View Formats) ---
		{
			name:         "GitHub_Blob_SimpleRef",
			uri:          "https://github.com/org/repo/blob/main/path/to/file.yaml",
			wantRepo:     "https://github.com/org/repo",
			wantRef:      "main",
			wantResource: "path/to/file.yaml",
			wantErr:      false,
		},
		{
			name:         "BitBucket_Src_TagRef",
			uri:          "https://bitbucket.org/team/project/src/v1.0.0/config.json",
			wantRepo:     "https://bitbucket.org/team/project",
			wantRef:      "v1.0.0",
			wantResource: "config.json",
			wantErr:      false,
		},
		{
			name:         "GitLab_DashBlob_CommitHash",
			uri:          "https://gitlab.com/group/proj/-/blob/a1b2c3d4e5f6/src/index.go",
			wantRepo:     "https://gitlab.com/group/proj",
			wantRef:      "a1b2c3d4e5f6",
			wantResource: "src/index.go",
			wantErr:      false,
		},
		{
			name:         "GitLab_DashBlob_TopLevelFile",
			uri:          "https://gitlab.com/group/proj/-/blob/main/README.md",
			wantRepo:     "https://gitlab.com/group/proj",
			wantRef:      "main",
			wantResource: "README.md",
			wantErr:      false,
		},

		// --- Success Cases (Generic Formats) ---
		{
			name:         "Generic_SimpleRef_HTTPS",
			uri:          "https://repo.corp.com/api/templates/-/dev/schemas/user.yaml",
			wantRepo:     "https://repo.corp.com/api/templates",
			wantRef:      "dev",
			wantResource: "schemas/user.yaml",
			wantErr:      false,
		},
		{
			name:         "Generic_ComplexRef_SSH_HashSeparator",
			uri:          "ssh://git@host:2222/api/templates/-/feature/add-auth#config.yaml",
			wantRepo:     "ssh://git@host:2222/api/templates",
			wantRef:      "feature/add-auth",
			wantResource: "config.yaml",
			wantErr:      false,
		},
		{
			name:         "Generic_ComplexRef_HTTPS_HashSeparator_WithFilePrefix",
			uri:          "https://private.repo/project/app/-/user/ui-v2#web/app.html",
			wantRepo:     "https://private.repo/project/app",
			wantRef:      "user/ui-v2",
			wantResource: "web/app.html",
			wantErr:      false,
		},
		{
			name:         "Generic_RefWithHyphen",
			uri:          "https://repo.corp.com/api/templates/-/release-1.0/file.txt",
			wantRepo:     "https://repo.corp.com/api/templates",
			wantRef:      "release-1.0",
			wantResource: "file.txt",
			wantErr:      false,
		},
		{
			name:         "Generic_GitAt_SimpleRef",
			uri:          "git@github.com:org/repo/-/main/template.go.tmpl",
			wantRepo:     "git@github.com:org/repo",
			wantRef:      "main",
			wantResource: "template.go.tmpl",
			wantErr:      false,
		},
		{
			name:         "Generic_GitAt_ComplexRef_HashSeparator",
			uri:          "git@gitlab.com:org/repo/-/feature/test-1#assets/data.json",
			wantRepo:     "git@gitlab.com:org/repo",
			wantRef:      "feature/test-1",
			wantResource: "assets/data.json",
			wantErr:      false,
		},

		// --- Error Cases ---
		{
			name:         "Error_UnknownFormat_MissingSeparator",
			uri:          "https://github.com/org/repo/main/path/to/file.yaml", // Missing /blob/, /src/, or /-/
			wantRepo:     "",
			wantRef:      "",
			wantResource: "",
			wantErr:      true,
		},
		{
			name:         "Error_MissingRefAndResource",
			uri:          "https://github.com/org/repo/blob/", // Separator present, but nothing follows
			wantRepo:     "",
			wantRef:      "",
			wantResource: "",
			wantErr:      true,
		},
		{
			name:         "Error_MissingResource",
			uri:          "https://github.com/org/repo/blob/main", // Ref present, but missing final / or # separator
			wantRepo:     "",
			wantRef:      "",
			wantResource: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ref, resource, err := git.ParseURI(tt.uri)
			if tt.wantErr == false {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			require.Equal(t, tt.wantRepo, repo)
			require.Equal(t, tt.wantRef, ref)
			require.Equal(t, tt.wantResource, resource)
		})
	}
}

func TestIsGitURI(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want bool
	}{
		// --- Valid Git URI Cases (Should be true) ---
		{
			name: "Valid_GitHub_Blob",
			uri:  "https://github.com/org/repo/blob/main/file.txt",
			want: true,
		},
		{
			name: "Valid_BitBucket_Src",
			uri:  "https://bitbucket.org/team/project/src/v1.0.0/config.json",
			want: true,
		},
		{
			name: "Valid_Generic_ComplexRef_SSH_GitAt",
			uri:  "git@host:org/repo/-/feature/add-auth#config.yaml",
			want: true,
		},
		{
			name: "Valid_Generic_SimpleRef_HTTPS",
			uri:  "https://repo.corp.com/api/templates/-/dev/schemas/user.yaml",
			want: true,
		},

		// --- Ambiguous Path Case (Should now be false) ---
		{
			name: "Invalid_LocalPath_AmbiguousSrcSeparator",
			uri:  "/home/user/my_project/src/main/config.yaml", // ParseURI succeeds, but lacks scheme
			want: false,
		},

		// --- Invalid/Malformed Git URI Cases (Should be false) ---
		{
			name: "Invalid_MissingSeparator",
			uri:  "https://repo.com/org/repo/main/file.txt", // Missing /blob/, /src/, or /-/
			want: false,
		},
		{
			name: "Invalid_MissingRef",
			uri:  "https://github.com/org/repo/blob/",
			want: false,
		},
		{
			name: "Invalid_MissingResource",
			uri:  "https://gitlab.com/group/proj/-/blob/main",
			want: false,
		},
		{
			name: "Invalid_StandardURL",
			uri:  "https://api.example.com/v1/users", // Just a regular web URL
			want: false,
		},
		{
			name: "Invalid_WebURLWithRepoName",
			uri:  "https://my-repo.git.com/project/api", // Looks like a repo URL but lacks separator
			want: false,
		},

		// --- Local File Path Cases (Should be false) ---
		{
			name: "Invalid_RelativeFilePath",
			uri:  "./templates/config.yaml",
			want: false,
		},
		{
			name: "Invalid_AbsolutePath_Linux",
			uri:  "/etc/templates/schema.json",
			want: false,
		},
		{
			name: "Invalid_AbsolutePath_Windows",
			uri:  "C:\\Users\\Template\\file.txt",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := git.IsGitURI(tt.uri)
			require.Equal(t, tt.want, got)
		})
	}
}
