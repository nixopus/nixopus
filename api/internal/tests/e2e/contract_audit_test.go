// Package e2e contains end-to-end and contract-validation tests.
//
// This file implements contract_audit_test.go — a zero-dependency, CI-safe test that
// reads the generated OpenAPI spec (api/doc/openapi.json) and cross-checks it against:
//  1. Struct field examples that describe the wrong format (e.g. "github.com/org/repo"
//     for a field that clone.go treats as a numeric ID with strconv.ParseInt).
//  2. Any example value for a field whose runtime usage is provably ParseInt/ParseUint.
//  3. Paths referenced in the skill file that are not present in the spec (stale docs).
//
// Run:  go test ./internal/tests/e2e/ -run TestContractAudit -v
package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// repoRoot returns the absolute path to the api/ directory (two levels above this file).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file = .../api/internal/tests/e2e/contract_audit_test.go
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}

// ---- OpenAPI helpers -------------------------------------------------------

type openAPISpec struct {
	Paths      map[string]map[string]openAPIOperation `json:"paths"`
	Components struct {
		Schemas map[string]openAPISchema `json:"schemas"`
	} `json:"components"`
}

type openAPIOperation struct {
	Summary     string                     `json:"summary"`
	OperationID string                     `json:"operationId"`
	RequestBody *openAPIRequestBody        `json:"requestBody"`
	Parameters  []openAPIParameter         `json:"parameters"`
	Responses   map[string]openAPIResponse `json:"responses"`
}

type openAPIRequestBody struct {
	Content map[string]openAPIMediaType `json:"content"`
}

type openAPIMediaType struct {
	Schema  openAPISchemaRef       `json:"schema"`
	Example map[string]interface{} `json:"example"`
}

type openAPISchemaRef struct {
	Ref        string                     `json:"$ref"`
	Type       string                     `json:"type"`
	Properties map[string]openAPIProperty `json:"properties"`
}

type openAPISchema struct {
	Type       string                     `json:"type"`
	Properties map[string]openAPIProperty `json:"properties"`
	Required   []string                   `json:"required"`
}

type openAPIProperty struct {
	Type        string      `json:"type"`
	Format      string      `json:"format"`
	Description string      `json:"description"`
	Example     interface{} `json:"example"` // can be string, number, bool, etc.
}

type openAPIParameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
}

type openAPIResponse struct {
	Description string `json:"description"`
}

// ---- Test: spec example values must not lie --------------------------------

// knownBadExamples maps a fully-qualified "SchemaName.fieldName" to the reason the
// current spec example is wrong and what the correct format looks like.
// Add entries here whenever a runtime failure reveals a documentation lie.
var knownBadExamples = map[string]string{
	// clone.go calls strconv.ParseInt on this field — "owner/repo" slug will fail.
	"CreateDeploymentRequest.repository":       "must be numeric GitHub repo ID (e.g. 912345678), not a slug",
	"CreateProjectRequest.repository":          "must be numeric GitHub repo ID (e.g. 912345678), not a slug",
	"PreviewComposeRequest.repository":         "must be numeric GitHub repo ID (e.g. 912345678), not a slug",
	"AddApplicationToFamilyRequest.repository": "must be numeric GitHub repo ID (e.g. 912345678), not a slug",
}

// badExamplePatterns are regexps that, when matched against an example value,
// indicate the example describes the wrong format for any field named "repository".
var badExamplePatterns = map[string]*regexp.Regexp{
	"repository": regexp.MustCompile(`[a-zA-Z]/`), // matches "owner/repo" or "github.com/org/repo"
}

func TestContractAudit_RepositoryFieldExamples(t *testing.T) {
	root := repoRoot(t)
	specPath := filepath.Join(root, "doc", "openapi.json")
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("cannot read %s: %v\n\nRun `swag init` or `make docs` to generate the spec first.", specPath, err)
	}

	var spec openAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("cannot parse openapi.json: %v", err)
	}

	failures := 0
	for schemaName, schema := range spec.Components.Schemas {
		for fieldName, prop := range schema.Properties {
			key := schemaName + "." + fieldName
			// Coerce example to string for pattern matching
			exampleStr := ""
			if prop.Example != nil {
				exampleStr = fmt.Sprintf("%v", prop.Example)
			}

			// 1. Check field is in our known-bad list and the example still says slug
			if reason, bad := knownBadExamples[key]; bad {
				pat, hasPat := badExamplePatterns[fieldName]
				if hasPat && pat.MatchString(exampleStr) {
					t.Errorf("STALE EXAMPLE: %s\n  example:  %q\n  reason:   %s\n  fix:      update the struct tag `example` in deploy/types/init.go", key, exampleStr, reason)
					failures++
				}
			}

			// 2. Generic check: any field named "repository" whose example contains "/"
			//    is almost certainly wrong unless the description explicitly says "URL".
			if fieldName == "repository" {
				pat := badExamplePatterns["repository"]
				descLower := strings.ToLower(prop.Description)
				exampleLooksLikeSlug := pat.MatchString(exampleStr)
				descSaysNumeric := strings.Contains(descLower, "numeric") || strings.Contains(descLower, "integer") || strings.Contains(descLower, "parsedint") || strings.Contains(descLower, "id")
				if exampleLooksLikeSlug && !descSaysNumeric {
					t.Errorf("MISLEADING REPOSITORY EXAMPLE: %s\n  example:  %q\n  The example looks like an owner/repo slug. If clone.go calls ParseInt on this\n  field the LLM (and any developer) reading the spec will pass the wrong format.\n  Fix: update the struct description and example to say 'numeric GitHub repo ID'.", key, exampleStr)
					failures++
				}
			}
		}
	}

	// Also check inline examples embedded in path operation requestBody
	for path, methods := range spec.Paths {
		for method, op := range methods {
			if op.RequestBody == nil {
				continue
			}
			for _, media := range op.RequestBody.Content {
				for fieldName, exVal := range media.Example {
					if fieldName != "repository" {
						continue
					}
					exStr, _ := exVal.(string)
					if exStr == "" {
						continue
					}
					pat := badExamplePatterns["repository"]
					if pat.MatchString(exStr) {
						t.Errorf("MISLEADING INLINE EXAMPLE: %s %s requestBody.example.repository = %q\n  This looks like an owner/repo slug. The LLM will copy this and break deployments.", strings.ToUpper(method), path, exStr)
						failures++
					}
				}
			}
		}
	}

	if failures == 0 {
		t.Log("All repository field examples look correct (numeric IDs, not slugs).")
	}
}

// ---- Test: skill paths exist in spec ---------------------------------------

func TestContractAudit_SkillPathsInSpec(t *testing.T) {
	root := repoRoot(t)

	specPath := filepath.Join(root, "doc", "openapi.json")
	specData, err := os.ReadFile(specPath)
	if err != nil {
		t.Skipf("openapi.json not found: %v", err)
	}
	var spec openAPISpec
	if err := json.Unmarshal(specData, &spec); err != nil {
		t.Fatalf("cannot parse openapi.json: %v", err)
	}

	skillPath := filepath.Join(root, "skills", "api-catalog", "SKILL.md")
	skillData, err := os.ReadFile(skillPath)
	if err != nil {
		t.Skipf("SKILL.md not found: %v", err)
	}
	skill := string(skillData)

	// Collect all /api/v1/... paths mentioned in the skill, normalising path params.
	rePath := regexp.MustCompile(`/api/v1[^\s\?{}\[\]'"]*`)
	normalise := func(p string) string {
		re := regexp.MustCompile(`/[a-f0-9-]{36}|/[0-9]+`)
		return re.ReplaceAllString(p, "/{id}")
	}

	specPaths := make(map[string]bool)
	for p := range spec.Paths {
		norm := regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(p, "{id}")
		specPaths[norm] = true
	}

	skillPaths := rePath.FindAllString(skill, -1)
	seen := make(map[string]bool)
	stale := 0
	for _, raw := range skillPaths {
		norm := normalise(raw)
		norm = regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(norm, "{id}")
		if seen[norm] {
			continue
		}
		seen[norm] = true
		if !specPaths[norm] {
			t.Logf("SKILL PATH NOT IN SPEC: %s (normalised: %s)", raw, norm)
			stale++
		}
	}
	if stale > 0 {
		t.Logf("Found %d paths in SKILL.md that are not in openapi.json.", stale)
		t.Logf("Either the spec is stale (re-run `make docs`) or the skill has incorrect paths.")
		// Not a hard failure — spec may need regeneration. Log only.
	} else {
		t.Log("All skill paths found in spec.")
	}
}

// ---- Test: unknown request fields not in skill should be tested ------------

// TestContractAudit_UnknownFieldRejection is a documentation test: it asserts
// that our skill explicitly warns about the fields we KNOW the API rejects.
// Add to `mustWarnAbout` any field the API returns 400 for.
func TestContractAudit_UnknownFieldRejection(t *testing.T) {
	root := repoRoot(t)
	skillPath := filepath.Join(root, "skills", "api-catalog", "SKILL.md")
	skillData, err := os.ReadFile(skillPath)
	if err != nil {
		t.Skipf("SKILL.md not found: %v", err)
	}
	skill := string(skillData)

	// Fields confirmed via E2E testing to cause HTTP 400 when included.
	mustWarnAbout := []struct {
		field  string
		reason string
	}{
		{"deploy_on_create", "API returns 400 'unknown field' — confirmed via E2E"},
		// Required fields that the LLM has been observed omitting in real deployments
		{"environment", "CreateDeploymentRequest.Environment is validate:required — confirmed via server logs (2026-05-07)"},
		{"build_pack", "CreateDeploymentRequest.BuildPack is validate:required — omitting causes HTTP 400"},
	}

	for _, w := range mustWarnAbout {
		if !strings.Contains(skill, w.field) {
			t.Errorf("SKILL MISSING WARNING: field %q is known to cause HTTP 400 (%s) but SKILL.md never mentions it.\n  Add: '❌ DO NOT include `%s` in any request body'", w.field, w.reason, w.field)
		} else {
			t.Logf("OK: %q is documented in SKILL.md", w.field)
		}
	}
}
