package validation

import (
	"testing"

	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func TestValidateCreateProjectRequest_PublicGit(t *testing.T) {
	tests := []struct {
		name    string
		req     types.CreateProjectRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid public_git with HTTPS URL",
			req: types.CreateProjectRequest{
				Name:       "sample-app",
				Repository: "https://github.com/nixopus/sample-app.git",
				Source:     shared_types.SourcePublicGit,
			},
			wantErr: false,
		},
		{
			name: "valid public_git without .git suffix",
			req: types.CreateProjectRequest{
				Name:       "sample-app",
				Repository: "https://github.com/nixopus/sample-app",
				Source:     shared_types.SourcePublicGit,
			},
			wantErr: false,
		},
		{
			name: "public_git rejects non-HTTPS URL",
			req: types.CreateProjectRequest{
				Name:       "sample-app",
				Repository: "git@github.com:nixopus/sample-app.git",
				Source:     shared_types.SourcePublicGit,
			},
			wantErr: true,
			errMsg:  "public_git source requires a valid HTTPS git URL",
		},
		{
			name: "public_git rejects empty repository",
			req: types.CreateProjectRequest{
				Name:       "sample-app",
				Repository: "",
				Source:     shared_types.SourcePublicGit,
			},
			wantErr: true,
			errMsg:  "repository is required",
		},
		{
			name: "github source accepts numeric repository",
			req: types.CreateProjectRequest{
				Name:       "sample-app",
				Repository: "12345",
				Source:     shared_types.SourceGithub,
			},
			wantErr: false,
		},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequest(&tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateDeploymentRequest_PublicGit(t *testing.T) {
	req := &types.CreateDeploymentRequest{
		Name:        "sample-app",
		Repository:  "https://github.com/nixopus/sample-app.git",
		Source:      shared_types.SourcePublicGit,
		Branch:      "main",
		Port:        3000,
		BuildPack:   "dockerfile",
		Environment: "production",
	}

	v := NewValidator()
	err := v.ValidateRequest(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIsValidPublicGitURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://github.com/nixopus/sample-app.git", true},
		{"https://github.com/nixopus/sample-app", true},
		{"https://gitlab.com/user/repo", true},
		{"http://github.com/nixopus/sample-app", false},
		{"git@github.com:nixopus/sample-app.git", false},
		{"", false},
		{"https://", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := isValidPublicGitURL(tt.url)
			if got != tt.want {
				t.Errorf("isValidPublicGitURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
