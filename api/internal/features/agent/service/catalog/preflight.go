package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Validator validates a nixopus_api call against the OpenAPI spec
// BEFORE the HTTP request is sent. This gives the LLM corrective feedback
// immediately rather than letting a 400 error create stale deployment records.
//
// The spec is loaded lazily on the first Validate call so the validator works
// even when NewAgentService runs before PostProcessSpecWithRetry writes the file.
type Validator struct {
	specPath       string
	once           sync.Once
	requiredFields map[string][]string
	fieldSchemas   map[string]map[string]fieldSchema
	fieldRules     map[string]fieldRule
}

type fieldSchema struct {
	Type     string
	Format   string
	Enum     []string
	Nullable bool
}

type fieldRule struct {
	hint     string
	validate func(v interface{}) string
}

// rejectedFields are field names the API rejects outright.
var rejectedFields = map[string]string{
	"deploy_on_create": "unknown field — API rejects the entire request with HTTP 400",
}

// NewValidator returns a validator that will lazily load the spec on
// first use. specPath is typically "doc/openapi.json".
func NewValidator(specPath string) *Validator {
	return &Validator{
		specPath: specPath,
		fieldRules: map[string]fieldRule{
			"repository": {
				hint: "repository MUST be the numeric GitHub repo ID (e.g. 912345678). " +
					"Call GET /api/v1/github-connector/repositories first and use the numeric `id` field. " +
					"Never pass an owner/repo slug like \"nixopus/sample-app\" or a URL like \"https://github.com/...\".",
				validate: func(val interface{}) string {
					s := fmt.Sprintf("%v", val)
					if _, err := strconv.ParseInt(s, 10, 64); err != nil {
						return fmt.Sprintf("repository value %q is not a numeric ID (ParseInt failed). "+
							"Use the integer `id` from GET /api/v1/github-connector/repositories.", s)
					}
					return ""
				},
			},
		},
	}
}

func (v *Validator) loadOnce() {
	v.once.Do(func() {
		v.requiredFields = make(map[string][]string)
		v.fieldSchemas = make(map[string]map[string]fieldSchema)

		data, err := os.ReadFile(v.specPath)
		if err != nil {
			cwd, _ := os.Getwd()
			data, err = os.ReadFile(filepath.Join(cwd, v.specPath))
		}
		if err != nil {
			return
		}

		var spec struct {
			Paths map[string]map[string]struct {
				RequestBody *struct {
					Content map[string]struct {
						Schema struct {
							Ref        string                 `json:"$ref"`
							Required   []string               `json:"required"`
							Properties map[string]interface{} `json:"properties"`
						} `json:"schema"`
					} `json:"content"`
				} `json:"requestBody"`
			} `json:"paths"`
			Components struct {
				Schemas map[string]struct {
					Required   []string                   `json:"required"`
					Properties map[string]json.RawMessage `json:"properties"`
				} `json:"schemas"`
			} `json:"components"`
		}
		if err := json.Unmarshal(data, &spec); err != nil {
			return
		}

		for path, methods := range spec.Paths {
			for method, op := range methods {
				if op.RequestBody == nil {
					continue
				}
				var required []string
				schemas := make(map[string]fieldSchema)

				for _, media := range op.RequestBody.Content {
					s := media.Schema
					if s.Ref != "" {
						name := s.Ref[strings.LastIndex(s.Ref, "/")+1:]
						if schema, ok := spec.Components.Schemas[name]; ok {
							required = append(required, schema.Required...)
							for fieldName, raw := range schema.Properties {
								schemas[fieldName] = parseFieldSchema(raw)
							}
						}
					} else {
						required = append(required, s.Required...)
					}
				}

				key := strings.ToUpper(method) + " " + path
				if len(required) > 0 {
					v.requiredFields[key] = unique(required)
				}
				if len(schemas) > 0 {
					v.fieldSchemas[key] = schemas
				}
			}
		}
	})
}

func parseFieldSchema(raw json.RawMessage) fieldSchema {
	var parsed struct {
		Type     string        `json:"type"`
		Format   string        `json:"format"`
		Enum     []interface{} `json:"enum"`
		Nullable bool          `json:"nullable"`
	}
	_ = json.Unmarshal(raw, &parsed)

	fs := fieldSchema{
		Type:     parsed.Type,
		Format:   parsed.Format,
		Nullable: parsed.Nullable,
	}
	for _, e := range parsed.Enum {
		fs.Enum = append(fs.Enum, fmt.Sprintf("%v", e))
	}
	return fs
}

// Validate checks the (method, rawPath, body) triple before the HTTP call is made.
// Returns a human-readable error string for the LLM. Returns "" when valid.
func (v *Validator) Validate(method, rawPath string, body json.RawMessage) string {
	v.loadOnce()

	specPath := NormalisePath(rawPath)
	key := strings.ToUpper(method) + " " + specPath

	var bodyMap map[string]interface{}
	if body != nil {
		_ = json.Unmarshal(body, &bodyMap)
	}

	sourceVal, _ := bodyMap["source"].(string)
	isPublicGit := sourceVal == "public_git"

	var errors []string

	// Check for rejected fields in the body
	for field, reason := range rejectedFields {
		if _, present := bodyMap[field]; present {
			errors = append(errors, fmt.Sprintf(
				"PREFLIGHT ERROR: field `%s` is rejected — %s. Remove it and retry.", field, reason))
		}
	}

	required := v.requiredFields[key]
	schemas := v.fieldSchemas[key]

	// Check missing required fields
	var missing []string
	for _, field := range required {
		val, present := bodyMap[field]
		if !present || val == nil || val == "" {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		msg := fmt.Sprintf(
			"PREFLIGHT ERROR: missing required fields for %s %s: [%s]. You MUST include all required fields and retry.",
			strings.ToUpper(method), specPath, strings.Join(missing, ", "))
		errors = append(errors, msg)
		for _, f := range missing {
			if rule, ok := v.fieldRules[f]; ok {
				errors = append(errors, "  Hint for "+f+": "+rule.hint)
			}
			if fs, ok := schemas[f]; ok && len(fs.Enum) > 0 {
				errors = append(errors, fmt.Sprintf("  Valid values for %s: %s", f, strings.Join(fs.Enum, ", ")))
			}
		}
	}

	// Check custom field rules (e.g. repository must be numeric)
	reported := map[string]bool{}
	for _, field := range required {
		val, present := bodyMap[field]
		if !present || val == nil || val == "" {
			continue
		}
		if field == "repository" && isPublicGit {
			continue
		}
		if rule, ok := v.fieldRules[field]; ok {
			if msg := rule.validate(val); msg != "" {
				errors = append(errors, msg)
				reported[field] = true
			}
		}
	}
	for field, val := range bodyMap {
		if reported[field] {
			continue
		}
		if field == "repository" && isPublicGit {
			continue
		}
		if rule, ok := v.fieldRules[field]; ok {
			if msg := rule.validate(val); msg != "" {
				errors = append(errors, msg)
			}
		}
	}

	// Check enum violations
	for field, val := range bodyMap {
		fs, ok := schemas[field]
		if !ok || len(fs.Enum) == 0 {
			continue
		}
		strVal := fmt.Sprintf("%v", val)
		found := false
		for _, allowed := range fs.Enum {
			if strVal == allowed {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, fmt.Sprintf(
				"PREFLIGHT ERROR: field `%s` value %q is not valid. Allowed values: [%s].",
				field, strVal, strings.Join(fs.Enum, ", ")))
		}
	}

	// Check type mismatches for known problematic types
	for field, val := range bodyMap {
		fs, ok := schemas[field]
		if !ok {
			continue
		}
		if msg := validateFieldType(field, val, fs); msg != "" {
			errors = append(errors, msg)
		}
	}

	if len(errors) == 0 {
		return ""
	}
	return strings.Join(errors, "\n")
}

func validateFieldType(field string, val interface{}, fs fieldSchema) string {
	switch fs.Type {
	case "integer":
		switch v := val.(type) {
		case float64:
			// JSON numbers are float64; this is fine
		case string:
			if _, err := strconv.Atoi(v); err != nil {
				return fmt.Sprintf("PREFLIGHT ERROR: field `%s` must be an integer, got string %q. Pass a number, not a string.", field, v)
			}
		default:
			return fmt.Sprintf("PREFLIGHT ERROR: field `%s` must be an integer, got %T.", field, val)
		}
	case "string":
		if fs.Format == "uuid" {
			s, ok := val.(string)
			if !ok {
				return fmt.Sprintf("PREFLIGHT ERROR: field `%s` must be a UUID string, got %T.", field, val)
			}
			if !reUUID.MatchString(s) {
				return fmt.Sprintf("PREFLIGHT ERROR: field `%s` value %q is not a valid UUID.", field, s)
			}
		}
	}
	return ""
}

var (
	reUUID    = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	reNumeric = regexp.MustCompile(`/[0-9]+(/|$)`)
)

// NormalisePath strips query string and replaces UUID-like and numeric path
// segments with the OpenAPI placeholder {id} for spec key matching.
func NormalisePath(raw string) string {
	if i := strings.IndexByte(raw, '?'); i >= 0 {
		raw = raw[:i]
	}
	raw = reUUID.ReplaceAllString(raw, "{id}")
	raw = reNumeric.ReplaceAllStringFunc(raw, func(m string) string {
		if strings.HasSuffix(m, "/") {
			return "/{id}/"
		}
		return "/{id}"
	})
	return raw
}

func unique(s []string) []string {
	seen := make(map[string]bool, len(s))
	out := make([]string, 0, len(s))
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
