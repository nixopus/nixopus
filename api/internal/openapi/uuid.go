package openapi

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

// SchemaCustomizer fixes types that kin-openapi generates empty or broken schemas for.
// Use with fuego.WithOpenAPIGeneratorSchemaCustomizer when creating the server.
func SchemaCustomizer(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// uuid.UUID (type [16]byte) -> string format uuid
	if t == reflect.TypeOf(uuid.UUID{}) {
		schema.Type = &openapi3.Types{"string"}
		schema.Format = "uuid"
		schema.Nullable = false
		return nil
	}

	// json.RawMessage ([]byte) -> string for JSON content
	if t == reflect.TypeOf(json.RawMessage(nil)) {
		schema.Type = &openapi3.Types{"string"}
		schema.Description = "JSON-encoded value"
		schema.Nullable = false
		return nil
	}

	// map[string]interface{} -> object with string values (variable values, etc.)
	if t.Kind() == reflect.Map && t.Key().Kind() == reflect.String && t.Elem().Kind() == reflect.Interface {
		// Fix empty additionalProperties: {} by setting a permissive value schema
		if schema.AdditionalProperties.Schema == nil || (schema.AdditionalProperties.Schema.Value != nil && schema.AdditionalProperties.Schema.Value.Type == nil) {
			schema.AdditionalProperties = openapi3.AdditionalProperties{
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:        &openapi3.Types{"string"},
						Description: "Variable value (string, number, boolean, or JSON as string)",
					},
				},
			}
		}
		return nil
	}

	// Common query/filter constraints improve LLM parameter selection quality.
	switch strings.ToLower(name) {
	case "page":
		schema.Type = &openapi3.Types{"integer"}
		schema.Min = ptrFloat(1)
		schema.Default = float64(1)
	case "page_size", "pagesize", "limit":
		schema.Type = &openapi3.Types{"integer"}
		schema.Min = ptrFloat(1)
		schema.Max = ptrFloat(100)
		if schema.Default == nil {
			schema.Default = float64(20)
		}
	case "sort_direction", "sortorder", "sort_order":
		schema.Type = &openapi3.Types{"string"}
		schema.Enum = []any{"asc", "desc"}
	case "period":
		schema.Type = &openapi3.Types{"string"}
		schema.Enum = []any{"1h", "24h", "7d", "30d"}
		if schema.Default == nil {
			schema.Default = "24h"
		}
	}

	if strings.EqualFold(name, "id") || strings.HasSuffix(strings.ToLower(name), "_id") {
		if schema.Type == nil || schema.Type.Is("string") {
			schema.Type = &openapi3.Types{"string"}
			if schema.Format == "" {
				schema.Format = "uuid"
			}
		}
	}

	return nil
}

func ptrFloat(v float64) *float64 {
	// Ensure finite numbers are always emitted.
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}
