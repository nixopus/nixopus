package openapi

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPtrFloat(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		p := ptrFloat(3.5)
		require.NotNil(t, p)
		assert.InDelta(t, 3.5, *p, 0)
	})
	t.Run("nan is nil", func(t *testing.T) { assert.Nil(t, ptrFloat(math.NaN())) })
	t.Run("inf is nil", func(t *testing.T) { assert.Nil(t, ptrFloat(math.Inf(1))) })
}

func TestSchemaCustomizer(t *testing.T) {
	t.Run("unwraps pointer to uuid", func(t *testing.T) {
		var u *uuid.UUID
		id := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		u = &id
		schema := &openapi3.Schema{}
		err := SchemaCustomizer("id", reflect.TypeOf(u), ``, schema)
		require.NoError(t, err)
		require.NotNil(t, schema.Type)
		assert.True(t, schema.Type.Is("string"))
		assert.Equal(t, "uuid", schema.Format)
	})

	t.Run("uuid value", func(t *testing.T) {
		schema := &openapi3.Schema{}
		err := SchemaCustomizer("x", reflect.TypeOf(uuid.UUID{}), ``, schema)
		require.NoError(t, err)
		assert.True(t, schema.Type.Is("string"))
		assert.Equal(t, "uuid", schema.Format)
	})

	t.Run("json.RawMessage", func(t *testing.T) {
		schema := &openapi3.Schema{}
		err := SchemaCustomizer("data", reflect.TypeOf(json.RawMessage(`{}`)), ``, schema)
		require.NoError(t, err)
		assert.Contains(t, schema.Description, "JSON")
	})

	t.Run("map string to interface empty additional properties", func(t *testing.T) {
		m := map[string]interface{}{}
		schema := &openapi3.Schema{AdditionalProperties: openapi3.AdditionalProperties{}}
		err := SchemaCustomizer("meta", reflect.TypeOf(m), ``, schema)
		require.NoError(t, err)
		require.NotNil(t, schema.AdditionalProperties.Schema)
		v := schema.AdditionalProperties.Schema.Value
		require.NotNil(t, v)
		require.NotNil(t, v.Type)
		assert.True(t, v.Type.Is("string"))
	})

	t.Run("map string to interface with schema type nil", func(t *testing.T) {
		m := map[string]interface{}{}
		schema := &openapi3.Schema{
			AdditionalProperties: openapi3.AdditionalProperties{
				Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: nil}},
			},
		}
		err := SchemaCustomizer("meta", reflect.TypeOf(m), ``, schema)
		require.NoError(t, err)
	})
	t.Run("map with additional properties already typed leaves them", func(t *testing.T) {
		m := map[string]interface{}{}
		schema := &openapi3.Schema{
			AdditionalProperties: openapi3.AdditionalProperties{
				Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"number"}}},
			},
		}
		err := SchemaCustomizer("meta", reflect.TypeOf(m), ``, schema)
		require.NoError(t, err)
		require.NotNil(t, schema.AdditionalProperties.Schema.Value)
		assert.True(t, schema.AdditionalProperties.Schema.Value.Type.Is("number"))
	})

	t.Run("name switch page", func(t *testing.T) {
		schema := &openapi3.Schema{}
		require.NoError(t, SchemaCustomizer("Page", reflect.TypeOf(0), ``, schema))
		assert.True(t, schema.Type.Is("integer"))
		require.NotNil(t, schema.Min)
		assert.Equal(t, 1.0, *schema.Min)
	})
	t.Run("name switch pagesize alias", func(t *testing.T) {
		schema := &openapi3.Schema{}
		require.NoError(t, SchemaCustomizer("pagesize", reflect.TypeOf(0), ``, schema))
		assert.True(t, schema.Type.Is("integer"))
	})
	t.Run("name switch limit default when missing", func(t *testing.T) {
		schema := &openapi3.Schema{}
		require.NoError(t, SchemaCustomizer("limit", reflect.TypeOf(0), ``, schema))
		if assert.NotNil(t, schema.Default) {
			assert.InDelta(t, 20, schema.Default, 0)
		}
	})
	t.Run("name switch sort", func(t *testing.T) {
		schema := &openapi3.Schema{}
		require.NoError(t, SchemaCustomizer("sort_order", reflect.TypeOf(""), ``, schema))
		assert.Greater(t, len(schema.Enum), 0)
	})
	t.Run("name switch sortorder alias", func(t *testing.T) {
		schema := &openapi3.Schema{}
		require.NoError(t, SchemaCustomizer("sortorder", reflect.TypeOf(""), ``, schema))
		assert.Greater(t, len(schema.Enum), 0)
	})
	t.Run("name switch period", func(t *testing.T) {
		schema := &openapi3.Schema{}
		require.NoError(t, SchemaCustomizer("period", reflect.TypeOf(""), ``, schema))
		assert.Greater(t, len(schema.Enum), 0)
	})
	t.Run("name switch period keeps existing default", func(t *testing.T) {
		schema := &openapi3.Schema{Default: "7d"}
		require.NoError(t, SchemaCustomizer("period", reflect.TypeOf(""), ``, schema))
		assert.Equal(t, "7d", schema.Default)
	})

	t.Run("id field adds uuid format when string and empty format", func(t *testing.T) {
		schema := &openapi3.Schema{Type: &openapi3.Types{"string"}}
		require.NoError(t, SchemaCustomizer("user_id", reflect.TypeOf(""), ``, schema))
		assert.Equal(t, "uuid", schema.Format)
	})
	t.Run("id field does not override existing format", func(t *testing.T) {
		schema := &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date"}
		require.NoError(t, SchemaCustomizer("other_id", reflect.TypeOf(""), ``, schema))
		assert.Equal(t, "date", schema.Format)
	})
	t.Run("id suffix skip when not stringish type", func(t *testing.T) {
		schema := &openapi3.Schema{Type: &openapi3.Types{"integer"}}
		require.NoError(t, SchemaCustomizer("other_id", reflect.TypeOf(0), ``, schema))
		assert.True(t, schema.Type.Is("integer"))
	})
}
