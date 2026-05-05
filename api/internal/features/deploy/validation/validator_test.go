package validation

import (
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

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

func mustUUID() uuid.UUID {
	return uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
}

func checkErr(t *testing.T, err error, wantErr bool, errMsg string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Errorf("expected error containing %q, got nil", errMsg)
			return
		}
		if errMsg != "" && !contains(err.Error(), errMsg) {
			t.Errorf("expected error containing %q, got %q", errMsg, err.Error())
		}
	} else {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// NewValidatorWithLogger
// ---------------------------------------------------------------------------

func TestNewValidatorWithLogger(t *testing.T) {
	l := logger.NewLogger()
	v := NewValidatorWithLogger(&l)
	if v == nil {
		t.Fatal("expected non-nil validator")
	}
	if v.Logger != &l {
		t.Fatal("expected logger to be set")
	}
}

// ---------------------------------------------------------------------------
// ParseRequestBody
// ---------------------------------------------------------------------------

func TestParseRequestBody(t *testing.T) {
	v := NewValidator()
	body := io.NopCloser(strings.NewReader(`{"name":"test"}`))
	var decoded map[string]string
	if err := v.ParseRequestBody(nil, body, &decoded); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded["name"] != "test" {
		t.Errorf("expected name=test, got %q", decoded["name"])
	}
}

func TestParseRequestBody_InvalidJSON(t *testing.T) {
	v := NewValidator()
	body := io.NopCloser(strings.NewReader(`not-json`))
	var decoded map[string]string
	if err := v.ParseRequestBody(nil, body, &decoded); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// ---------------------------------------------------------------------------
// ValidateRequest — default branch (unknown type + logger)
// ---------------------------------------------------------------------------

func TestValidateRequest_UnknownType(t *testing.T) {
	v := NewValidator()
	err := v.ValidateRequest("not a request")
	if err != types.ErrInvalidRequestType {
		t.Errorf("expected ErrInvalidRequestType, got %v", err)
	}
}

func TestValidateRequest_UnknownType_WithLogger(t *testing.T) {
	l := logger.NewLogger()
	v := NewValidatorWithLogger(&l)
	err := v.ValidateRequest(42)
	if err != types.ErrInvalidRequestType {
		t.Errorf("expected ErrInvalidRequestType, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// ValidateRequest — all 13 type-switch branches (happy path)
// ---------------------------------------------------------------------------

func TestValidateRequest_AllBranches_HappyPath(t *testing.T) {
	id := mustUUID()
	v := NewValidator()

	// CreateDeploymentRequest
	if err := v.ValidateRequest(&types.CreateDeploymentRequest{
		Name: "app", Environment: "production", BuildPack: "dockerfile",
		Repository: "https://github.com/org/repo", Branch: "main", Port: 3000,
	}); err != nil {
		t.Errorf("CreateDeploymentRequest: %v", err)
	}

	// CreateProjectRequest
	if err := v.ValidateRequest(&types.CreateProjectRequest{
		Name: "app", Repository: "12345", Source: shared_types.SourceGithub,
	}); err != nil {
		t.Errorf("CreateProjectRequest: %v", err)
	}

	// DeployProjectRequest
	if err := v.ValidateRequest(&types.DeployProjectRequest{ID: id}); err != nil {
		t.Errorf("DeployProjectRequest: %v", err)
	}

	// UpdateDeploymentRequest
	if err := v.ValidateRequest(&types.UpdateDeploymentRequest{ID: id}); err != nil {
		t.Errorf("UpdateDeploymentRequest: %v", err)
	}

	// DeleteDeploymentRequest
	if err := v.ValidateRequest(&types.DeleteDeploymentRequest{ID: id}); err != nil {
		t.Errorf("DeleteDeploymentRequest: %v", err)
	}

	// ReDeployApplicationRequest
	if err := v.ValidateRequest(&types.ReDeployApplicationRequest{ID: id}); err != nil {
		t.Errorf("ReDeployApplicationRequest: %v", err)
	}

	// RollbackDeploymentRequest
	if err := v.ValidateRequest(&types.RollbackDeploymentRequest{ID: id}); err != nil {
		t.Errorf("RollbackDeploymentRequest: %v", err)
	}

	// RestartDeploymentRequest
	if err := v.ValidateRequest(&types.RestartDeploymentRequest{ID: id}); err != nil {
		t.Errorf("RestartDeploymentRequest: %v", err)
	}

	// DuplicateProjectRequest
	if err := v.ValidateRequest(&types.DuplicateProjectRequest{
		SourceProjectID: id, Environment: "staging",
	}); err != nil {
		t.Errorf("DuplicateProjectRequest: %v", err)
	}

	// GetProjectFamilyRequest
	if err := v.ValidateRequest(&types.GetProjectFamilyRequest{FamilyID: id}); err != nil {
		t.Errorf("GetProjectFamilyRequest: %v", err)
	}

	// AddApplicationToFamilyRequest
	if err := v.ValidateRequest(&types.AddApplicationToFamilyRequest{
		Name: "app", Repository: "12345",
	}); err != nil {
		t.Errorf("AddApplicationToFamilyRequest: %v", err)
	}

	// CreateTemplateDeploymentRequest
	if err := v.ValidateRequest(&types.CreateTemplateDeploymentRequest{
		TemplateID: "wordpress", Name: "my-wp",
	}); err != nil {
		t.Errorf("CreateTemplateDeploymentRequest: %v", err)
	}

	// CancelDeploymentRequest
	if err := v.ValidateRequest(&types.CancelDeploymentRequest{DeploymentID: id}); err != nil {
		t.Errorf("CancelDeploymentRequest: %v", err)
	}
}

// ---------------------------------------------------------------------------
// validateDeploymentRequest
// ---------------------------------------------------------------------------

func TestValidateDeploymentRequest(t *testing.T) {
	id := mustUUID()
	_ = id

	tests := []struct {
		name    string
		req     types.CreateDeploymentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: types.CreateDeploymentRequest{
				Name: "app", Environment: "production", BuildPack: "dockerfile",
				Repository: "https://github.com/org/repo", Branch: "main", Port: 3000,
			},
		},
		{
			name:    "empty name",
			req:     types.CreateDeploymentRequest{},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "empty environment",
			req:     types.CreateDeploymentRequest{Name: "app"},
			wantErr: true,
			errMsg:  "environment is required",
		},
		{
			name:    "invalid environment",
			req:     types.CreateDeploymentRequest{Name: "app", Environment: "badenv"},
			wantErr: true,
			errMsg:  "invalid environment",
		},
		{
			name:    "empty buildpack",
			req:     types.CreateDeploymentRequest{Name: "app", Environment: "production"},
			wantErr: true,
			errMsg:  "build_pack is required",
		},
		{
			name:    "invalid buildpack",
			req:     types.CreateDeploymentRequest{Name: "app", Environment: "production", BuildPack: "heroku"},
			wantErr: true,
			errMsg:  "invalid build pack",
		},
		{
			name:    "empty repository",
			req:     types.CreateDeploymentRequest{Name: "app", Environment: "production", BuildPack: "dockerfile"},
			wantErr: true,
			errMsg:  "repository is required",
		},
		{
			name: "public_git invalid URL",
			req: types.CreateDeploymentRequest{
				Name: "app", Environment: "production", BuildPack: "dockerfile",
				Repository: "git@github.com:org/repo.git", Source: shared_types.SourcePublicGit,
			},
			wantErr: true,
			errMsg:  "public_git source requires a valid HTTPS git URL",
		},
		{
			name: "empty branch",
			req: types.CreateDeploymentRequest{
				Name: "app", Environment: "production", BuildPack: "dockerfile",
				Repository: "repo123",
			},
			wantErr: true,
			errMsg:  "branch is required",
		},
		{
			name: "zero port",
			req: types.CreateDeploymentRequest{
				Name: "app", Environment: "production", BuildPack: "dockerfile",
				Repository: "repo123", Branch: "main",
			},
			wantErr: true,
			errMsg:  "port is required",
		},
		{
			name: "invalid domain",
			req: types.CreateDeploymentRequest{
				Name: "app", Environment: "production", BuildPack: "dockerfile",
				Repository: "repo123", Branch: "main", Port: 3000,
				Domains: []string{"not a domain!"},
			},
			wantErr: true,
			errMsg:  "invalid domain",
		},
		{
			name: "base path empty normalized to slash",
			req: types.CreateDeploymentRequest{
				Name: "app", Environment: "production", BuildPack: "dockerfile",
				Repository: "repo123", Branch: "main", Port: 3000,
				BasePath: "",
			},
		},
		{
			name: "base path without leading slash gets normalized",
			req: types.CreateDeploymentRequest{
				Name: "app", Environment: "production", BuildPack: "dockerfile",
				Repository: "repo123", Branch: "main", Port: 3000,
				BasePath: "api",
			},
		},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequest(&tt.req)
			checkErr(t, err, tt.wantErr, tt.errMsg)
		})
	}
}

// ---------------------------------------------------------------------------
// validateUpdateDeploymentRequest
// ---------------------------------------------------------------------------

func TestValidateUpdateDeploymentRequest(t *testing.T) {
	id := mustUUID()

	tests := []struct {
		name    string
		req     types.UpdateDeploymentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing ID",
			req:     types.UpdateDeploymentRequest{},
			wantErr: true,
			errMsg:  "id is required",
		},
		{
			name:    "name too short",
			req:     types.UpdateDeploymentRequest{ID: id, Name: "ab"},
			wantErr: true,
			errMsg:  "name must be at least 3 characters",
		},
		{
			name:    "invalid environment",
			req:     types.UpdateDeploymentRequest{ID: id, Environment: "badenv"},
			wantErr: true,
			errMsg:  "invalid environment",
		},
		{
			name:    "invalid buildpack",
			req:     types.UpdateDeploymentRequest{ID: id, BuildPack: "heroku"},
			wantErr: true,
			errMsg:  "invalid build pack",
		},
		{
			name:    "port out of range (negative)",
			req:     types.UpdateDeploymentRequest{ID: id, Port: -1},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name:    "port out of range (too high)",
			req:     types.UpdateDeploymentRequest{ID: id, Port: 70000},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "base path without leading slash gets normalized",
			req:  types.UpdateDeploymentRequest{ID: id, BasePath: "api"},
		},
		{
			name:    "too many domains",
			req:     types.UpdateDeploymentRequest{ID: id, Domains: []string{"a.com", "b.com", "c.com", "d.com", "e.com", "f.com"}},
			wantErr: true,
			errMsg:  "maximum 5 domains allowed",
		},
		{
			name: "valid minimal update",
			req:  types.UpdateDeploymentRequest{ID: id},
		},
		{
			name: "valid with all fields",
			req: types.UpdateDeploymentRequest{
				ID: id, Name: "my-app", Environment: "staging",
				BuildPack: "dockerfile", Port: 8080, BasePath: "/api",
				Domains: []string{"a.com", "b.com"},
			},
		},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequest(&tt.req)
			checkErr(t, err, tt.wantErr, tt.errMsg)
		})
	}
}

// ---------------------------------------------------------------------------
// validateDeleteDeploymentRequest
// ---------------------------------------------------------------------------

func TestValidateDeleteDeploymentRequest(t *testing.T) {
	v := NewValidator()

	t.Run("missing ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.DeleteDeploymentRequest{})
		checkErr(t, err, true, "id is required")
	})
	t.Run("valid ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.DeleteDeploymentRequest{ID: mustUUID()})
		checkErr(t, err, false, "")
	})
}

// ---------------------------------------------------------------------------
// validateRedeployApplicationRequest
// ---------------------------------------------------------------------------

func TestValidateRedeployApplicationRequest(t *testing.T) {
	v := NewValidator()

	t.Run("missing ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.ReDeployApplicationRequest{})
		checkErr(t, err, true, "id is required")
	})
	t.Run("valid ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.ReDeployApplicationRequest{ID: mustUUID()})
		checkErr(t, err, false, "")
	})
}

// ---------------------------------------------------------------------------
// validateRollbackDeploymentRequest
// ---------------------------------------------------------------------------

func TestValidateRollbackDeploymentRequest(t *testing.T) {
	v := NewValidator()

	t.Run("missing ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.RollbackDeploymentRequest{})
		checkErr(t, err, true, "id is required")
	})
	t.Run("valid ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.RollbackDeploymentRequest{ID: mustUUID()})
		checkErr(t, err, false, "")
	})
}

// ---------------------------------------------------------------------------
// validateRestartDeploymentRequest
// ---------------------------------------------------------------------------

func TestValidateRestartDeploymentRequest(t *testing.T) {
	v := NewValidator()

	t.Run("missing ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.RestartDeploymentRequest{})
		checkErr(t, err, true, "id is required")
	})
	t.Run("valid ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.RestartDeploymentRequest{ID: mustUUID()})
		checkErr(t, err, false, "")
	})
}

// ---------------------------------------------------------------------------
// validateCreateProjectRequest (existing + missing branches)
// ---------------------------------------------------------------------------

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
		},
		{
			name: "valid public_git without .git suffix",
			req: types.CreateProjectRequest{
				Name:       "sample-app",
				Repository: "https://github.com/nixopus/sample-app",
				Source:     shared_types.SourcePublicGit,
			},
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
		},
		{
			name:    "empty name",
			req:     types.CreateProjectRequest{},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "empty repository",
			req:     types.CreateProjectRequest{Name: "app"},
			wantErr: true,
			errMsg:  "repository is required",
		},
		{
			name: "invalid environment",
			req: types.CreateProjectRequest{
				Name: "app", Repository: "12345", Environment: "badenv",
			},
			wantErr: true,
			errMsg:  "invalid environment",
		},
		{
			name: "base path without leading slash gets normalized",
			req: types.CreateProjectRequest{
				Name: "app", Repository: "12345", BasePath: "api",
			},
		},
		{
			name: "invalid domain",
			req: types.CreateProjectRequest{
				Name: "app", Repository: "12345",
				Domains: []string{"bad domain!"},
			},
			wantErr: true,
			errMsg:  "invalid domain",
		},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequest(&tt.req)
			checkErr(t, err, tt.wantErr, tt.errMsg)
		})
	}
}

// ---------------------------------------------------------------------------
// validateDeployProjectRequest
// ---------------------------------------------------------------------------

func TestValidateDeployProjectRequest(t *testing.T) {
	v := NewValidator()

	t.Run("missing ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.DeployProjectRequest{})
		checkErr(t, err, true, "id is required")
	})
	t.Run("valid ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.DeployProjectRequest{ID: mustUUID()})
		checkErr(t, err, false, "")
	})
}

// ---------------------------------------------------------------------------
// validateDuplicateProjectRequest
// ---------------------------------------------------------------------------

func TestValidateDuplicateProjectRequest(t *testing.T) {
	id := mustUUID()

	tests := []struct {
		name    string
		req     types.DuplicateProjectRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing source project ID",
			req:     types.DuplicateProjectRequest{},
			wantErr: true,
			errMsg:  "source project id is required",
		},
		{
			name: "invalid domain",
			req: types.DuplicateProjectRequest{
				SourceProjectID: id,
				Domains:         []string{"bad domain"},
			},
			wantErr: true,
			errMsg:  "invalid domain",
		},
		{
			name: "empty environment",
			req: types.DuplicateProjectRequest{
				SourceProjectID: id,
			},
			wantErr: true,
			errMsg:  "invalid environment",
		},
		{
			name: "invalid environment",
			req: types.DuplicateProjectRequest{
				SourceProjectID: id, Environment: "badenv",
			},
			wantErr: true,
			errMsg:  "invalid environment",
		},
		{
			name: "valid",
			req: types.DuplicateProjectRequest{
				SourceProjectID: id, Environment: "staging",
			},
		},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequest(&tt.req)
			checkErr(t, err, tt.wantErr, tt.errMsg)
		})
	}
}

// ---------------------------------------------------------------------------
// validateGetProjectFamilyRequest
// ---------------------------------------------------------------------------

func TestValidateGetProjectFamilyRequest(t *testing.T) {
	v := NewValidator()

	t.Run("missing ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.GetProjectFamilyRequest{})
		checkErr(t, err, true, "id is required")
	})
	t.Run("valid ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.GetProjectFamilyRequest{FamilyID: mustUUID()})
		checkErr(t, err, false, "")
	})
}

// ---------------------------------------------------------------------------
// validateAddApplicationToFamilyRequest
// ---------------------------------------------------------------------------

func TestValidateAddApplicationToFamilyRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     types.AddApplicationToFamilyRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing name",
			req:     types.AddApplicationToFamilyRequest{},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "missing repository",
			req:     types.AddApplicationToFamilyRequest{Name: "app"},
			wantErr: true,
			errMsg:  "repository is required",
		},
		{
			name: "invalid domain",
			req: types.AddApplicationToFamilyRequest{
				Name: "app", Repository: "12345",
				Domains: []string{"not valid!"},
			},
			wantErr: true,
			errMsg:  "invalid domain",
		},
		{
			name: "invalid environment",
			req: types.AddApplicationToFamilyRequest{
				Name: "app", Repository: "12345", Environment: "badenv",
			},
			wantErr: true,
			errMsg:  "invalid environment",
		},
		{
			name: "valid with defaults",
			req:  types.AddApplicationToFamilyRequest{Name: "app", Repository: "12345"},
		},
		{
			name: "valid with explicit fields",
			req: types.AddApplicationToFamilyRequest{
				Name: "app", Repository: "12345", Environment: "staging",
				BuildPack: "dockerfile", Branch: "main", Port: 8080,
			},
		},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequest(&tt.req)
			checkErr(t, err, tt.wantErr, tt.errMsg)
		})
	}
}

// ---------------------------------------------------------------------------
// validateTemplateDeploymentRequest
// ---------------------------------------------------------------------------

func TestValidateTemplateDeploymentRequest(t *testing.T) {
	v := NewValidator()

	t.Run("missing template_id", func(t *testing.T) {
		err := v.ValidateRequest(&types.CreateTemplateDeploymentRequest{Name: "my-app"})
		checkErr(t, err, true, "template_id is required")
	})
	t.Run("missing name", func(t *testing.T) {
		err := v.ValidateRequest(&types.CreateTemplateDeploymentRequest{TemplateID: "wordpress"})
		checkErr(t, err, true, "name is required")
	})
	t.Run("valid", func(t *testing.T) {
		err := v.ValidateRequest(&types.CreateTemplateDeploymentRequest{TemplateID: "wordpress", Name: "my-wp"})
		checkErr(t, err, false, "")
	})
}

// ---------------------------------------------------------------------------
// validateCancelDeploymentRequest
// ---------------------------------------------------------------------------

func TestValidateCancelDeploymentRequest(t *testing.T) {
	v := NewValidator()

	t.Run("missing ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.CancelDeploymentRequest{})
		checkErr(t, err, true, "id is required")
	})
	t.Run("valid ID", func(t *testing.T) {
		err := v.ValidateRequest(&types.CancelDeploymentRequest{DeploymentID: mustUUID()})
		checkErr(t, err, false, "")
	})
}

// ---------------------------------------------------------------------------
// isDomainValid — exhaustive pure-logic coverage
// ---------------------------------------------------------------------------

func TestIsDomainValid(t *testing.T) {
	tests := []struct {
		domain string
		want   bool
	}{
		// valid
		{"example.com", true},
		{"sub.example.com", true},
		{"my-app.example.com", true},
		{"a1.b2.c3.d4", true},
		// empty
		{"", false},
		// too long (254 chars)
		{strings.Repeat("a", 250) + ".com", false},
		// contains slash
		{"foo/bar.com", false},
		// contains backslash
		{"foo\\bar.com", false},
		// contains space
		{"foo bar.com", false},
		// contains tab
		{"foo\tbar.com", false},
		// contains newline
		{"foo\nbar.com", false},
		// contains carriage return
		{"foo\rbar.com", false},
		// single label (no dot)
		{"localhost", false},
		// label empty (double dot)
		{"foo..bar", false},
		// label too long (64 chars)
		{strings.Repeat("a", 64) + ".com", false},
		// hyphen at start of label
		{"-foo.com", false},
		// hyphen at end of label
		{"foo-.com", false},
		// non-alnum non-hyphen in label
		{"foo!.com", false},
		{"foo_.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := isDomainValid(tt.domain)
			if got != tt.want {
				t.Errorf("isDomainValid(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// validateDomains
// ---------------------------------------------------------------------------

func TestValidateDomains(t *testing.T) {
	t.Run("nil domains", func(t *testing.T) {
		if err := validateDomains(nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("empty slice", func(t *testing.T) {
		if err := validateDomains([]string{}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("valid domains", func(t *testing.T) {
		if err := validateDomains([]string{"example.com", "app.example.com"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("invalid domain in list", func(t *testing.T) {
		if err := validateDomains([]string{"example.com", "bad domain"}); err == nil {
			t.Error("expected error for invalid domain")
		}
	})
}

// ---------------------------------------------------------------------------
// isValidPublicGitURL — complete branch coverage
// ---------------------------------------------------------------------------

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
		// no https prefix
		{"http://github.com/nixopus/sample-app", false},
		{"git@github.com:nixopus/sample-app.git", false},
		{"", false},
		// https:// with no host
		{"https://", false},
		// https://host with no path separator (no slash after host)
		{"https://github.com", false},
		// path part is only whitespace
		{"https://github.com/   ", false},
		// empty host (https:///path → host part is empty)
		{"https:///path", false},
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
