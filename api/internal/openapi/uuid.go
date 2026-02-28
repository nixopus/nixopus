package openapi

import (
	"encoding/json"
	"reflect"

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

	return nil
}
