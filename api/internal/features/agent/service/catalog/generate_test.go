package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCatalog_FromRealSpec(t *testing.T) {
	specPath := findSpecPath(t)

	catalog, err := GenerateCatalog(specPath)
	if err != nil {
		t.Fatalf("GenerateCatalog failed: %v", err)
	}

	if catalog == "" {
		t.Fatal("generated catalog is empty")
	}

	if !strings.HasPrefix(catalog, "[api-catalog]") {
		t.Error("catalog must start with [api-catalog]")
	}
	if !strings.HasSuffix(catalog, "[/api-catalog]") {
		t.Error("catalog must end with [/api-catalog]")
	}

	t.Logf("Generated catalog: %d bytes, %d lines", len(catalog), strings.Count(catalog, "\n"))
}

func TestGenerateCatalog_RequiredFieldsPresent(t *testing.T) {
	specPath := findSpecPath(t)

	catalog, err := GenerateCatalog(specPath)
	if err != nil {
		t.Fatalf("GenerateCatalog failed: %v", err)
	}

	mustContain := []struct {
		text   string
		reason string
	}{
		{"REQUIRED:", "catalog must show required fields"},
		{"environment", "environment is a required field for create deployment"},
		{"build_pack", "build_pack is a required field for create deployment"},
		{"repository", "repository is a required field"},
		{"deploy_on_create", "must warn about deploy_on_create bad field"},
	}

	for _, tc := range mustContain {
		if !strings.Contains(catalog, tc.text) {
			t.Errorf("catalog missing %q: %s", tc.text, tc.reason)
		}
	}
}

func TestGenerateCatalog_NoRequiredFieldDrift(t *testing.T) {
	specPath := findSpecPath(t)
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Skipf("cannot read spec: %v", err)
	}

	var spec struct {
		Components struct {
			Schemas map[string]struct {
				Required   []string               `json:"required"`
				Properties map[string]interface{} `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("parse spec: %v", err)
	}

	catalog, err := GenerateCatalog(specPath)
	if err != nil {
		t.Fatalf("GenerateCatalog failed: %v", err)
	}

	for schemaName, schema := range spec.Components.Schemas {
		if !strings.HasSuffix(schemaName, "Request") {
			continue
		}
		for _, req := range schema.Required {
			if !strings.Contains(catalog, req) {
				t.Errorf("DRIFT: required field %q from schema %s is missing in generated catalog", req, schemaName)
			}
		}
	}
}

func TestGenerateCatalog_EnumsIncluded(t *testing.T) {
	specPath := findSpecPath(t)

	catalog, err := GenerateCatalog(specPath)
	if err != nil {
		t.Fatalf("GenerateCatalog failed: %v", err)
	}

	if !strings.Contains(catalog, "enum:") {
		t.Error("catalog should contain enum values for fields that have them")
	}
}

func TestGenerateCatalog_MissingSpecReturnsError(t *testing.T) {
	_, err := GenerateCatalog("/nonexistent/path/openapi.json")
	if err == nil {
		t.Error("expected error for missing spec file")
	}
}

func findSpecPath(t *testing.T) string {
	t.Helper()
	candidates := []string{
		"../../../../doc/openapi.json",
		"../../../../../doc/openapi.json",
		"doc/openapi.json",
	}

	cwd, _ := os.Getwd()
	for _, c := range candidates {
		abs := filepath.Join(cwd, c)
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	abs := filepath.Join(cwd, "..", "..", "..", "..", "doc", "openapi.json")
	if _, err := os.Stat(abs); err != nil {
		t.Skipf("openapi.json not found (cwd=%s)", cwd)
	}
	return abs
}
