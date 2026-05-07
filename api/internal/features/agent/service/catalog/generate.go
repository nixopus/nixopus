package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// specSchema mirrors the subset of OpenAPI 3.x we need for catalog generation.
type specSchema struct {
	Paths      map[string]map[string]specOperation `json:"paths"`
	Components struct {
		Schemas map[string]specComponentSchema `json:"schemas"`
	} `json:"components"`
}

type specOperation struct {
	Summary     string          `json:"summary"`
	OperationID string          `json:"operationId"`
	Description string          `json:"description"`
	Tags        []string        `json:"tags"`
	Parameters  []specParameter `json:"parameters"`
	RequestBody *struct {
		Content map[string]struct {
			Schema specSchemaRef `json:"schema"`
		} `json:"content"`
	} `json:"requestBody"`
}

type specParameter struct {
	Name     string        `json:"name"`
	In       string        `json:"in"`
	Required bool          `json:"required"`
	Schema   specSchemaRef `json:"schema"`
}

type specSchemaRef struct {
	Ref        string                  `json:"$ref"`
	Type       string                  `json:"type"`
	Format     string                  `json:"format"`
	Enum       []interface{}           `json:"enum"`
	Required   []string                `json:"required"`
	Properties map[string]specFieldDef `json:"properties"`
	Nullable   bool                    `json:"nullable"`
	Items      *specSchemaRef          `json:"items"`
}

type specFieldDef struct {
	Type        string         `json:"type"`
	Format      string         `json:"format"`
	Enum        []interface{}  `json:"enum"`
	Description string         `json:"description"`
	Example     interface{}    `json:"example"`
	Nullable    bool           `json:"nullable"`
	Items       *specSchemaRef `json:"items"`
}

type specComponentSchema struct {
	Type       string                  `json:"type"`
	Required   []string                `json:"required"`
	Properties map[string]specFieldDef `json:"properties"`
}

// knownFieldWarnings adds critical LLM guidance that cannot be inferred from the
// schema alone (e.g. "repository" looks like a string but must be a numeric ID).
var knownFieldWarnings = map[string]string{
	"repository": "numeric GitHub repo ID from GET /api/v1/github-connector/repositories (NOT owner/repo slug or URL). For source:public_git use full HTTPS git URL instead.",
}

// knownBadFields are fields the API rejects outright.
var knownBadFields = []string{
	"deploy_on_create",
}

// sectionOrder controls the display order of tag groups in the catalog.
var sectionOrder = []string{
	"Deploy", "Domains", "GitHub Connector", "Containers",
	"Machines", "MCP", "Extensions", "Notifications",
	"Health Checks", "Webhooks", "Health", "System",
	"Update", "Audit", "Feature Flags", "Auth", "User",
	"Trial", "File Manager", "Telemetry",
}

func sectionPriority(tag string) int {
	for i, t := range sectionOrder {
		if strings.EqualFold(t, tag) {
			return i
		}
	}
	return len(sectionOrder)
}

type endpointEntry struct {
	Method   string
	Path     string
	Summary  string
	Tag      string
	Query    []paramEntry
	Required []fieldEntry
	Optional []fieldEntry
	Warnings []string
}

type paramEntry struct {
	Name     string
	Type     string
	Required bool
	Enum     []string
}

type fieldEntry struct {
	Name    string
	Type    string
	Enum    []string
	Format  string
	Warning string
}

// GenerateCatalog reads the OpenAPI spec and produces a compact, LLM-optimized
// catalog string with required fields, types, enums, and warnings.
func GenerateCatalog(specPath string) (string, error) {
	data, err := readSpecFile(specPath)
	if err != nil {
		return "", err
	}

	var spec specSchema
	if err := json.Unmarshal(data, &spec); err != nil {
		return "", fmt.Errorf("catalog: parse spec: %w", err)
	}

	endpoints := buildEndpoints(&spec)
	return renderCatalog(endpoints), nil
}

func readSpecFile(specPath string) ([]byte, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		cwd, _ := os.Getwd()
		data, err = os.ReadFile(filepath.Join(cwd, specPath))
	}
	return data, err
}

func buildEndpoints(spec *specSchema) []endpointEntry {
	var endpoints []endpointEntry

	for path, methods := range spec.Paths {
		for method, op := range methods {
			e := endpointEntry{
				Method:  strings.ToUpper(method),
				Path:    path,
				Summary: op.Summary,
			}

			if len(op.Tags) > 0 {
				e.Tag = op.Tags[0]
			} else {
				e.Tag = inferTag(path)
			}

			for _, p := range op.Parameters {
				if p.In == "query" {
					pe := paramEntry{
						Name:     p.Name,
						Required: p.Required,
						Type:     schemaRefType(&p.Schema),
					}
					pe.Enum = enumStrings(p.Schema.Enum)
					e.Query = append(e.Query, pe)
				}
			}

			if op.RequestBody != nil {
				for _, media := range op.RequestBody.Content {
					required, optional, warnings := resolveBodyFields(&media.Schema, spec)
					e.Required = required
					e.Optional = optional
					e.Warnings = warnings
					break
				}
			}

			for _, bad := range knownBadFields {
				e.Warnings = append(e.Warnings, fmt.Sprintf("DO NOT include `%s` in body", bad))
			}

			endpoints = append(endpoints, e)
		}
	}

	sort.Slice(endpoints, func(i, j int) bool {
		pi, pj := sectionPriority(endpoints[i].Tag), sectionPriority(endpoints[j].Tag)
		if pi != pj {
			return pi < pj
		}
		if endpoints[i].Path != endpoints[j].Path {
			return endpoints[i].Path < endpoints[j].Path
		}
		return methodOrder(endpoints[i].Method) < methodOrder(endpoints[j].Method)
	})

	return endpoints
}

func resolveBodyFields(ref *specSchemaRef, spec *specSchema) (required []fieldEntry, optional []fieldEntry, warnings []string) {
	var props map[string]specFieldDef
	var requiredNames []string

	if ref.Ref != "" {
		name := ref.Ref[strings.LastIndex(ref.Ref, "/")+1:]
		if schema, ok := spec.Components.Schemas[name]; ok {
			props = schema.Properties
			requiredNames = schema.Required
		}
	} else {
		props = ref.Properties
		requiredNames = ref.Required
	}

	reqSet := make(map[string]bool, len(requiredNames))
	for _, r := range requiredNames {
		reqSet[r] = true
	}

	fieldNames := make([]string, 0, len(props))
	for name := range props {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	for _, name := range fieldNames {
		prop := props[name]
		fe := fieldEntry{
			Name:   name,
			Type:   fieldType(prop),
			Enum:   enumStrings(prop.Enum),
			Format: prop.Format,
		}
		if w, ok := knownFieldWarnings[name]; ok {
			fe.Warning = w
		}

		if reqSet[name] {
			required = append(required, fe)
		} else {
			optional = append(optional, fe)
		}
	}

	return required, optional, warnings
}

func fieldType(f specFieldDef) string {
	t := f.Type
	if t == "" {
		t = "object"
	}
	if f.Format != "" && f.Format != "uuid" {
		t = f.Format
	}
	if f.Format == "uuid" {
		t = "uuid"
	}
	if t == "array" && f.Items != nil {
		inner := f.Items.Type
		if inner == "" {
			inner = "object"
		}
		t = "array[" + inner + "]"
	}
	return t
}

func schemaRefType(ref *specSchemaRef) string {
	if ref == nil {
		return "string"
	}
	t := ref.Type
	if t == "" {
		t = "string"
	}
	if ref.Format == "uuid" {
		return "uuid"
	}
	if ref.Format != "" {
		return ref.Format
	}
	return t
}

func enumStrings(vals []interface{}) []string {
	if len(vals) == 0 {
		return nil
	}
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		out = append(out, fmt.Sprintf("%v", v))
	}
	return out
}

func inferTag(path string) string {
	p := strings.TrimPrefix(path, "/api/v1/")
	if idx := strings.IndexByte(p, '/'); idx > 0 {
		p = p[:idx]
	}
	if idx := strings.IndexByte(p, '?'); idx > 0 {
		p = p[:idx]
	}
	switch p {
	case "deploy":
		return "Deploy"
	case "domain":
		return "Domains"
	case "github-connector":
		return "GitHub Connector"
	case "container":
		return "Containers"
	case "machines":
		return "Machines"
	case "mcp":
		return "MCP"
	case "notification":
		return "Notifications"
	case "healthcheck":
		return "Health Checks"
	case "extensions":
		return "Extensions"
	case "health":
		return "Health"
	case "update":
		return "Update"
	case "audit":
		return "Audit"
	case "feature-flags":
		return "Feature Flags"
	default:
		return "System"
	}
}

func methodOrder(m string) int {
	switch m {
	case "GET":
		return 0
	case "POST":
		return 1
	case "PUT":
		return 2
	case "PATCH":
		return 3
	case "DELETE":
		return 4
	default:
		return 5
	}
}

func renderCatalog(endpoints []endpointEntry) string {
	var sb strings.Builder

	sb.WriteString("[api-catalog]\n")
	sb.WriteString("Use nixopus_api(method, path, body) for ALL Nixopus API calls below.\n")
	sb.WriteString("Pass the HTTP method and API path directly. For path params, embed them in the path string.\n")
	sb.WriteString("REQUIRED fields MUST be included or the call WILL fail with HTTP 400.\n\n")

	sb.WriteString("CALLING FORMAT:\n")
	sb.WriteString("  nixopus_api({ method: \"GET\",  path: \"/api/v1/deploy/applications?page=1\" })\n")
	sb.WriteString("  nixopus_api({ method: \"POST\", path: \"/api/v1/deploy/application\", body: { name: \"my-app\", repository: \"912345678\", environment: \"production\", build_pack: \"dockerfile\", branch: \"main\", port: 3000 } })\n\n")

	sb.WriteString("GLOBAL RULES:\n")
	for _, bad := range knownBadFields {
		sb.WriteString(fmt.Sprintf("- DO NOT include `%s` in any request body (unknown field, causes HTTP 400)\n", bad))
	}
	sb.WriteString("- `repository` format depends on `source`: source:\"github\" -> numeric repo ID; source:\"public_git\" -> full HTTPS git URL\n")
	sb.WriteString("- `source` valid values: github (default), public_git, s3, zip, template\n")
	sb.WriteString("- For deployment queries use param `id` NOT `application_id`\n")
	sb.WriteString("- Container ops use path params: /api/v1/container/{id} (NOT query ?id=)\n\n")

	currentTag := ""
	for _, ep := range endpoints {
		if ep.Tag != currentTag {
			currentTag = ep.Tag
			sb.WriteString("## " + currentTag + "\n")
		}
		renderEndpoint(&sb, ep)
	}

	sb.WriteString("[/api-catalog]")
	return sb.String()
}

func renderEndpoint(sb *strings.Builder, ep endpointEntry) {
	summary := ep.Summary
	if summary == "" {
		summary = ep.Path
	}

	sb.WriteString(fmt.Sprintf("%s %s", ep.Method, ep.Path))

	if len(ep.Query) > 0 {
		qParts := make([]string, 0, len(ep.Query))
		for _, q := range ep.Query {
			s := q.Name
			if len(q.Enum) > 0 {
				s += "(enum:" + strings.Join(q.Enum, "|") + ")"
			}
			if !q.Required {
				s += "?"
			}
			qParts = append(qParts, s)
		}
		sb.WriteString(fmt.Sprintf(" — %s. Query: %s", summary, strings.Join(qParts, ", ")))
	} else {
		sb.WriteString(fmt.Sprintf(" — %s", summary))
	}
	sb.WriteString("\n")

	if len(ep.Required) > 0 {
		parts := make([]string, 0, len(ep.Required))
		for _, f := range ep.Required {
			s := f.Name + "(" + f.Type
			if len(f.Enum) > 0 {
				s += " enum:" + strings.Join(f.Enum, "|")
			}
			s += ")"
			if f.Warning != "" {
				s += " ← " + f.Warning
			}
			parts = append(parts, s)
		}
		sb.WriteString("  REQUIRED: " + strings.Join(parts, ", ") + "\n")
	}

	if len(ep.Optional) > 0 {
		parts := make([]string, 0, len(ep.Optional))
		for _, f := range ep.Optional {
			s := f.Name + "(" + f.Type
			if len(f.Enum) > 0 {
				s += " enum:" + strings.Join(f.Enum, "|")
			}
			s += ")"
			parts = append(parts, s)
		}
		sb.WriteString("  OPTIONAL: " + strings.Join(parts, ", ") + "\n")
	}
}
