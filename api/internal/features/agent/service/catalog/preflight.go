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
	specPath string
	once     sync.Once
	requiredFields map[string][]string
	fieldRules     map[string]fieldRule
}

type fieldRule struct {
	hint     string
	validate func(v interface{}) string
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
					Required   []string               `json:"required"`
					Properties map[string]interface{} `json:"properties"`
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
				for _, media := range op.RequestBody.Content {
					s := media.Schema
					if s.Ref != "" {
						name := s.Ref[strings.LastIndex(s.Ref, "/")+1:]
						if schema, ok := spec.Components.Schemas[name]; ok {
							required = append(required, schema.Required...)
						}
					} else {
						required = append(required, s.Required...)
					}
				}
				if len(required) > 0 {
					key := strings.ToUpper(method) + " " + path
					v.requiredFields[key] = unique(required)
				}
			}
		}
	})
}

// Validate checks the (method, rawPath, body) triple before the HTTP call is made.
// Returns a human-readable error string for the LLM. Returns "" when valid.
func (v *Validator) Validate(method, rawPath string, body json.RawMessage) string {
	v.loadOnce()

	if len(v.requiredFields) == 0 {
		return ""
	}

	specPath := NormalisePath(rawPath)
	key := strings.ToUpper(method) + " " + specPath

	required, ok := v.requiredFields[key]
	if !ok {
		return ""
	}

	if len(required) == 0 || (body == nil && len(required) == 0) {
		return ""
	}

	var bodyMap map[string]interface{}
	if body != nil {
		_ = json.Unmarshal(body, &bodyMap)
	}

	sourceVal, _ := bodyMap["source"].(string)
	isPublicGit := sourceVal == "public_git"

	var missing []string
	var ruleErrors []string

	for _, field := range required {
		val, present := bodyMap[field]
		if !present || val == nil || val == "" {
			missing = append(missing, field)
			continue
		}
		if field == "repository" && isPublicGit {
			continue
		}
		if rule, hasRule := v.fieldRules[field]; hasRule {
			if msg := rule.validate(val); msg != "" {
				ruleErrors = append(ruleErrors, msg)
			}
		}
	}

	for field, val := range bodyMap {
		if field == "repository" && isPublicGit {
			continue
		}
		if rule, hasRule := v.fieldRules[field]; hasRule {
			if msg := rule.validate(val); msg != "" {
				alreadyReported := false
				for _, e := range ruleErrors {
					if strings.Contains(e, field) {
						alreadyReported = true
					}
				}
				if !alreadyReported {
					ruleErrors = append(ruleErrors, msg)
				}
			}
		}
	}

	if len(missing) == 0 && len(ruleErrors) == 0 {
		return ""
	}

	var parts []string
	if len(missing) > 0 {
		parts = append(parts, fmt.Sprintf(
			"PREFLIGHT ERROR: missing required fields for %s %s: [%s]. "+
				"You MUST include all required fields and retry.",
			strings.ToUpper(method), specPath, strings.Join(missing, ", "),
		))
		for _, f := range missing {
			if rule, hasRule := v.fieldRules[f]; hasRule {
				parts = append(parts, "  Hint for "+f+": "+rule.hint)
			}
		}
	}
	parts = append(parts, ruleErrors...)
	return strings.Join(parts, "\n")
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
