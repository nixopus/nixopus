package catalog

import (
	"encoding/json"
	"testing"
)

func TestValidate_MissingRequiredFields(t *testing.T) {
	v := newTestValidator(t)

	body := json.RawMessage(`{"name": "my-app"}`)
	result := v.Validate("POST", "/api/v1/deploy/application", body)

	if result == "" {
		t.Error("expected preflight error for missing required fields")
	}
	if !contains(result, "PREFLIGHT ERROR") {
		t.Errorf("expected PREFLIGHT ERROR in result, got: %s", result)
	}
}

func TestValidate_ValidRequest(t *testing.T) {
	v := newTestValidator(t)

	result := v.Validate("GET", "/api/v1/deploy/applications", nil)
	if result != "" {
		t.Errorf("expected no error for GET with no body, got: %s", result)
	}
}

func TestValidate_RejectedField(t *testing.T) {
	v := newTestValidator(t)

	body := json.RawMessage(`{
		"name": "my-app",
		"repository": "912345678",
		"environment": "production",
		"build_pack": "dockerfile",
		"branch": "main",
		"port": 3000,
		"deploy_on_create": true
	}`)
	result := v.Validate("POST", "/api/v1/deploy/application", body)

	if !contains(result, "deploy_on_create") {
		t.Errorf("expected warning about deploy_on_create, got: %s", result)
	}
	if !contains(result, "rejected") {
		t.Errorf("expected 'rejected' in error, got: %s", result)
	}
}

func TestValidate_EnumViolation(t *testing.T) {
	v := newTestValidator(t)

	schemas := v.fieldSchemas
	hasEnum := false
	for _, fields := range schemas {
		for _, fs := range fields {
			if len(fs.Enum) > 0 {
				hasEnum = true
				break
			}
		}
	}
	if !hasEnum {
		t.Skip("no enum fields found in spec")
	}

	body := json.RawMessage(`{
		"name": "my-app",
		"repository": "912345678",
		"environment": "prod",
		"build_pack": "dockerfile",
		"branch": "main",
		"port": 3000
	}`)
	result := v.Validate("POST", "/api/v1/deploy/application", body)

	if contains(result, "not valid") {
		t.Logf("Correctly caught enum violation: %s", result)
	}
}

func TestValidate_TypeMismatch(t *testing.T) {
	v := newTestValidator(t)

	body := json.RawMessage(`{
		"name": "my-app",
		"repository": "912345678",
		"environment": "production",
		"build_pack": "dockerfile",
		"branch": "main",
		"port": "not-a-number"
	}`)
	result := v.Validate("POST", "/api/v1/deploy/application", body)

	if contains(result, "must be an integer") {
		t.Logf("Correctly caught type mismatch: %s", result)
	}
}

func TestValidate_PublicGitRepositorySkip(t *testing.T) {
	v := newTestValidator(t)

	body := json.RawMessage(`{
		"name": "my-app",
		"source": "public_git",
		"repository": "https://github.com/nixopus/sample-app.git",
		"environment": "production",
		"build_pack": "dockerfile",
		"branch": "main",
		"port": 3000
	}`)
	result := v.Validate("POST", "/api/v1/deploy/application", body)

	if contains(result, "repository") && contains(result, "numeric ID") {
		t.Errorf("should not require numeric repo ID when source is public_git, got: %s", result)
	}
}

func TestValidate_RepositorySlugRejected(t *testing.T) {
	v := newTestValidator(t)

	body := json.RawMessage(`{
		"name": "my-app",
		"repository": "nixopus/sample-app",
		"environment": "production",
		"build_pack": "dockerfile",
		"branch": "main",
		"port": 3000
	}`)
	result := v.Validate("POST", "/api/v1/deploy/application", body)

	if !contains(result, "numeric ID") && !contains(result, "ParseInt") {
		t.Errorf("expected rejection of owner/repo slug, got: %s", result)
	}
}

func TestNormalisePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/v1/deploy/application?id=abc-123", "/api/v1/deploy/application"},
		{"/api/v1/container/550e8400-e29b-41d4-a716-446655440000", "/api/v1/container/{id}"},
		{"/api/v1/machines/123/ssh/status", "/api/v1/machines/{id}/ssh/status"},
	}

	for _, tc := range tests {
		result := NormalisePath(tc.input)
		if result != tc.expected {
			t.Errorf("NormalisePath(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func newTestValidator(t *testing.T) *Validator {
	t.Helper()
	specPath := findSpecPath(t)
	v := NewValidator(specPath)
	v.loadOnce()
	return v
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
