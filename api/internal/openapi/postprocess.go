package openapi

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// PostProcessSpecWithRetry waits for the generated OpenAPI file to exist and
// applies LLM-focused contract enhancements.
func PostProcessSpecWithRetry(specPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		if _, err := os.Stat(specPath); err != nil {
			lastErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}

		if err := PostProcessSpec(specPath); err != nil {
			lastErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}
		return nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("timed out waiting for spec generation")
	}
	return lastErr
}

// PostProcessSpec updates generated OpenAPI with examples, stronger query
// constraints, and a standard error envelope schema.
func PostProcessSpec(specPath string) error {
	raw, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return err
	}

	paths, _ := doc["paths"].(map[string]any)
	components, _ := doc["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	seenOperationIDs := map[string]int{}

	for routePath, pathValue := range paths {
		ops, ok := pathValue.(map[string]any)
		if !ok {
			continue
		}
		for method, opValue := range ops {
			op, ok := opValue.(map[string]any)
			if !ok {
				continue
			}
			summary, _ := op["summary"].(string)
			if strings.TrimSpace(summary) == "" {
				summary = fallbackSummary(method, routePath)
				op["summary"] = summary
			}
			op["description"] = buildDescription(summary, method, routePath)
			normalizeOperationID(op, method, routePath, seenOperationIDs)
			addParameterExamplesAndConstraints(op)
			addRequestExamples(op, schemas)
			standardizeErrorResponses(op)
			addSuccessExamples(op, method, schemas)
		}
	}

	ensureErrorEnvelopeSchema(schemas)

	encoded, err := json.MarshalIndent(doc, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(specPath, encoded, 0o644)
}

func normalizeOperationID(op map[string]any, method, routePath string, seen map[string]int) {
	summary, _ := op["summary"].(string)
	if strings.TrimSpace(summary) == "" {
		summary = fallbackSummary(method, routePath)
	}
	base := toLowerCamel(summary)
	if base == "" {
		base = toLowerCamel(fallbackSummary(method, routePath))
	}
	if base == "" {
		base = strings.ToLower(method) + "Operation"
	}

	seen[base]++
	if seen[base] > 1 {
		op["operationId"] = fmt.Sprintf("%s%d", base, seen[base])
		return
	}
	op["operationId"] = base
}

func addParameterExamplesAndConstraints(op map[string]any) {
	params, _ := op["parameters"].([]any)
	for _, p := range params {
		param, ok := p.(map[string]any)
		if !ok {
			continue
		}
		in, _ := param["in"].(string)
		name, _ := param["name"].(string)
		if in != "query" || name == "" {
			continue
		}

		schema, _ := param["schema"].(map[string]any)
		if schema == nil {
			continue
		}

		if _, hasExample := param["example"]; !hasExample {
			if example, ok := queryParamExample(name); ok {
				param["example"] = example
			}
		}

		switch name {
		case "page":
			schema["type"] = "integer"
			schema["minimum"] = float64(1)
			schema["default"] = float64(1)
		case "page_size", "limit":
			schema["type"] = "integer"
			schema["minimum"] = float64(1)
			schema["maximum"] = float64(100)
			if _, ok := schema["default"]; !ok {
				schema["default"] = float64(20)
			}
		case "sort_direction", "sort_order":
			schema["type"] = "string"
			schema["enum"] = []any{"asc", "desc"}
		case "period":
			schema["type"] = "string"
			schema["enum"] = []any{"1h", "24h", "7d", "30d"}
			schema["default"] = "24h"
		}

		if schemaType, _ := schema["type"].(string); schemaType == "string" && looksLikeUUID(name) {
			schema["format"] = "uuid"
		}
	}
}

func addRequestExamples(op map[string]any, schemas map[string]any) {
	requestBody, _ := op["requestBody"].(map[string]any)
	if requestBody == nil {
		return
	}
	content, _ := requestBody["content"].(map[string]any)
	for _, mediaValue := range content {
		media, ok := mediaValue.(map[string]any)
		if !ok {
			continue
		}
		if _, hasExample := media["example"]; hasExample {
			continue
		}
		schema, _ := media["schema"].(map[string]any)
		refName := refSchemaName(schema)
		if refName == "" {
			continue
		}
		if example := buildSchemaExample(refName, schemas, 0, map[string]bool{}); example != nil {
			media["example"] = example
		}
	}
}

func addSuccessExamples(op map[string]any, method string, schemas map[string]any) {
	responses, _ := op["responses"].(map[string]any)
	if responses == nil {
		return
	}

	successCode := "200"
	if strings.EqualFold(method, "post") {
		if _, ok := responses["201"]; ok {
			successCode = "201"
		}
	}

	resp, _ := responses[successCode].(map[string]any)
	content, _ := resp["content"].(map[string]any)
	for _, mediaValue := range content {
		media, ok := mediaValue.(map[string]any)
		if !ok {
			continue
		}
		if _, hasExample := media["example"]; hasExample {
			continue
		}
		schema, _ := media["schema"].(map[string]any)
		refName := refSchemaName(schema)
		if refName == "" {
			continue
		}
		if example := buildSchemaExample(refName, schemas, 0, map[string]bool{}); example != nil {
			media["example"] = example
		}
	}
}

func standardizeErrorResponses(op map[string]any) {
	responses, _ := op["responses"].(map[string]any)
	if responses == nil {
		return
	}

	errorCodes := []string{"400", "401", "403", "404", "409", "422", "429", "500"}
	for _, code := range errorCodes {
		resp, _ := responses[code].(map[string]any)
		if resp == nil {
			continue
		}
		content, _ := resp["content"].(map[string]any)
		for _, mediaValue := range content {
			media, ok := mediaValue.(map[string]any)
			if !ok {
				continue
			}
			media["schema"] = map[string]any{
				"$ref": "#/components/schemas/ErrorEnvelope",
			}
			if _, ok := media["example"]; !ok {
				media["example"] = map[string]any{
					"code":    errorCodeName(code),
					"message": "Request failed",
					"details": map[string]any{},
				}
			}
		}
	}
}

func ensureErrorEnvelopeSchema(schemas map[string]any) {
	if schemas == nil {
		return
	}
	schemas["ErrorEnvelope"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"code": map[string]any{
				"type":        "string",
				"description": "Machine-readable error code",
				"example":     "invalid_request",
			},
			"message": map[string]any{
				"type":        "string",
				"description": "Human-readable error message",
				"example":     "Request validation failed",
			},
			"details": map[string]any{
				"type":        "object",
				"description": "Additional context for programmatic handling",
				"additionalProperties": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []any{"code", "message"},
	}
}

func queryParamExample(name string) (any, bool) {
	switch name {
	case "page":
		return float64(1), true
	case "page_size", "limit":
		return float64(20), true
	case "search", "search_term":
		return "nginx", true
	case "sort_by":
		return "created_at", true
	case "sort_direction", "sort_order":
		return "desc", true
	case "is_active":
		return true, true
	case "period":
		return "24h", true
	case "level":
		return "info", true
	case "start_time":
		return "2026-01-01T00:00:00Z", true
	case "end_time":
		return "2026-01-02T00:00:00Z", true
	}
	return nil, false
}

func looksLikeUUID(name string) bool {
	return name == "id" || strings.HasSuffix(name, "_id")
}

func refSchemaName(schema map[string]any) string {
	if schema == nil {
		return ""
	}
	ref, _ := schema["$ref"].(string)
	if ref == "" {
		return ""
	}
	const prefix = "#/components/schemas/"
	return strings.TrimPrefix(ref, prefix)
}

func errorCodeName(statusCode string) string {
	switch statusCode {
	case "400":
		return "invalid_request"
	case "401":
		return "unauthorized"
	case "403":
		return "forbidden"
	case "404":
		return "not_found"
	case "409":
		return "conflict"
	case "422":
		return "unprocessable_entity"
	case "429":
		return "rate_limited"
	case "500":
		return "internal_error"
	default:
		return "error"
	}
}

func buildSchemaExample(name string, schemas map[string]any, depth int, seen map[string]bool) any {
	if depth > 3 || seen[name] {
		return nil
	}
	seen[name] = true
	defer delete(seen, name)

	rawSchema, ok := schemas[name]
	if !ok {
		return nil
	}
	schema, ok := rawSchema.(map[string]any)
	if !ok {
		return nil
	}

	schemaType, _ := schema["type"].(string)
	switch schemaType {
	case "object":
		props, _ := schema["properties"].(map[string]any)
		if len(props) == 0 {
			return map[string]any{}
		}
		out := map[string]any{}
		for propName, propValue := range props {
			propSchema, _ := propValue.(map[string]any)
			if propSchema == nil {
				continue
			}
			if ref := refSchemaName(propSchema); ref != "" {
				if v := buildSchemaExample(ref, schemas, depth+1, seen); v != nil {
					out[propName] = v
				}
				continue
			}
			if ex, ok := propSchema["example"]; ok {
				out[propName] = ex
				continue
			}
			out[propName] = fallbackPropertyExample(propName, propSchema)
		}
		return out
	case "array":
		itemSchema, _ := schema["items"].(map[string]any)
		if itemSchema == nil {
			return []any{}
		}
		if ref := refSchemaName(itemSchema); ref != "" {
			if v := buildSchemaExample(ref, schemas, depth+1, seen); v != nil {
				return []any{v}
			}
			return []any{}
		}
		return []any{fallbackPropertyExample("item", itemSchema)}
	default:
		return fallbackPropertyExample(name, schema)
	}
}

func fallbackPropertyExample(name string, schema map[string]any) any {
	if ex, ok := schema["example"]; ok {
		return ex
	}
	schemaType, _ := schema["type"].(string)
	switch schemaType {
	case "string":
		if format, _ := schema["format"].(string); format == "uuid" || looksLikeUUID(name) {
			return "00000000-0000-0000-0000-000000000000"
		}
		if strings.Contains(name, "email") {
			return "user@example.com"
		}
		if strings.Contains(name, "time") || strings.Contains(name, "date") {
			return "2026-01-01T00:00:00Z"
		}
		if enumVals, ok := schema["enum"].([]any); ok && len(enumVals) > 0 {
			return enumVals[0]
		}
		return "string"
	case "integer", "number":
		if min, ok := schema["minimum"]; ok {
			return min
		}
		return float64(1)
	case "boolean":
		return true
	case "array":
		return []any{}
	case "object":
		return map[string]any{}
	default:
		return nil
	}
}
