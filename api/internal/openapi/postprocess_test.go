package openapi

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorCodeName(t *testing.T) {
	for code, name := range map[string]string{
		"400": "invalid_request", "401": "unauthorized", "403": "forbidden", "404": "not_found",
		"409": "conflict", "422": "unprocessable_entity", "429": "rate_limited", "500": "internal_error",
	} {
		assert.Equal(t, name, errorCodeName(code), code)
	}
	assert.Equal(t, "error", errorCodeName("418"))
}

func TestLooksLikeUUID(t *testing.T) {
	assert.True(t, looksLikeUUID("id"))
	assert.True(t, looksLikeUUID("user_id"))
	assert.False(t, looksLikeUUID("name"))
}

func TestRefSchemaName(t *testing.T) {
	assert.Equal(t, "Foo", refSchemaName(map[string]any{"$ref": "#/components/schemas/Foo"}))
	assert.Equal(t, "", refSchemaName(nil))
	assert.Equal(t, "", refSchemaName(map[string]any{}))
}

func TestQueryParamExample(t *testing.T) {
	for _, n := range []string{
		"page", "limit", "search", "search_term", "sort_by", "sort_order",
		"is_active", "period", "level", "start_time", "end_time",
	} {
		v, ok := queryParamExample(n)
		assert.True(t, ok, n)
		assert.NotNil(t, v, n)
	}
	_, ok := queryParamExample("unknown")
	assert.False(t, ok)
}

func TestFallbackPropertyExample(t *testing.T) {
	t.Run("example in schema", func(t *testing.T) {
		assert.Equal(t, 42, fallbackPropertyExample("x", map[string]any{"example": 42}))
	})
	t.Run("string uuid format", func(t *testing.T) {
		assert.Equal(t, "00000000-0000-0000-0000-000000000000",
			fallbackPropertyExample("x", map[string]any{"type": "string", "format": "uuid"}))
	})
	t.Run("string email", func(t *testing.T) {
		s := fallbackPropertyExample("user_email", map[string]any{"type": "string"})
		assert.Contains(t, s, "@")
	})
	t.Run("string date", func(t *testing.T) {
		s := fallbackPropertyExample("created_time", map[string]any{"type": "string"})
		assert.Contains(t, s, "2026")
	})
	t.Run("string enum", func(t *testing.T) {
		assert.Equal(t, "a", fallbackPropertyExample("e", map[string]any{
			"type": "string", "enum": []any{"a", "b"},
		}))
	})
	t.Run("number minimum", func(t *testing.T) {
		assert.Equal(t, float64(2), fallbackPropertyExample("n", map[string]any{
			"type": "number", "minimum": float64(2),
		}))
	})
	for _, st := range []string{"boolean", "array", "object", "other"} {
		if st == "other" {
			assert.Nil(t, fallbackPropertyExample("x", map[string]any{"type": st}))
		} else {
			assert.NotNil(t, fallbackPropertyExample("x", map[string]any{"type": st}), st)
		}
	}
	// number without minimum uses 1.0
	assert.Equal(t, float64(1), fallbackPropertyExample("n", map[string]any{"type": "number"}))
	assert.Equal(t, float64(1), fallbackPropertyExample("i", map[string]any{"type": "integer"}))
}

func TestBuildSchemaExample(t *testing.T) {
	schemas := map[string]any{
		"Child": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"val": map[string]any{"type": "string", "example": "ok"},
			},
		},
		"Parent": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"child": map[string]any{"$ref": "#/components/schemas/Child"},
			},
		},
		"A": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"b": map[string]any{"$ref": "#/components/schemas/B"},
			},
		},
		"B": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"a": map[string]any{"$ref": "#/components/schemas/A"},
			},
		},
		"List": map[string]any{
			"type":  "array",
			"items": map[string]any{"$ref": "#/components/schemas/Child"},
		},
		"ListBare": map[string]any{
			"type":  "array",
			"items": map[string]any{"type": "string"},
		},
		"EmptyObj": map[string]any{
			"type": "object",
		},
	}
	ex := buildSchemaExample("Parent", schemas, 0, map[string]bool{})
	require.NotNil(t, ex)
	m, _ := ex.(map[string]any)
	require.NotNil(t, m["child"])

	// Array with ref
	ex2 := buildSchemaExample("List", schemas, 0, map[string]bool{})
	require.NotNil(t, ex2)
	// Array with inline item type (not $ref) → fallbackPropertyExample
	exBare := buildSchemaExample("ListBare", schemas, 0, map[string]bool{})
	assert.Equal(t, []any{"string"}, exBare)

	// Array with nil items
	exNilItems := buildSchemaExample("EmptyArr", map[string]any{
		"EmptyArr": map[string]any{"type": "array"},
	}, 0, map[string]bool{})
	assert.Equal(t, []any{}, exNilItems)

	// Prop schema nil in object
	exNilProp := buildSchemaExample("NilProp", map[string]any{
		"NilProp": map[string]any{
			"type": "object", "properties": map[string]any{"x": any(nil)},
		},
	}, 0, map[string]bool{})
	require.NotNil(t, exNilProp)

	// default schema type in buildSchemaExample (e.g. number) via non-object, non-array
	assert.NotNil(t, buildSchemaExample("N", map[string]any{
		"N": map[string]any{"type": "number", "minimum": 3.0},
	}, 0, map[string]bool{}))

	// Unknown schema name
	assert.Nil(t, buildSchemaExample("Missing", schemas, 0, map[string]bool{}))
	assert.Nil(t, buildSchemaExample("T", map[string]any{
		"T": "not-a-map",
	}, 0, map[string]bool{}))

	// Array item $ref to missing schema => empty slice
	assert.Equal(t, []any{}, buildSchemaExample("OrphanArr", map[string]any{
		"OrphanArr": map[string]any{
			"type":  "array",
			"items": map[string]any{"$ref": "#/components/schemas/NoSuch"},
		},
	}, 0, map[string]bool{}))

	// Empty object
	ex3 := buildSchemaExample("EmptyObj", schemas, 0, map[string]bool{})
	assert.Equal(t, map[string]any{}, ex3)

	// depth limit
	seen := map[string]bool{}
	assert.Nil(t, buildSchemaExample("A", schemas, 4, seen))
}

func TestNormalizeOperationID(t *testing.T) {
	seen := map[string]int{}
	op1 := map[string]any{"summary": "List Things"}
	normalizeOperationID(op1, "get", "/a", seen)
	id1, _ := op1["operationId"].(string)
	op2 := map[string]any{"summary": "List Things"}
	normalizeOperationID(op2, "get", "/b", seen)
	id2, _ := op2["operationId"].(string)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id2, "2")
}

func TestNormalizeOperationID_punctuationFallsBack(t *testing.T) {
	seen := map[string]int{}
	op := map[string]any{"summary": "@@@"} // toLowerCamel yields ""
	normalizeOperationID(op, "get", "/api/v1/widgets", seen)
	assert.NotEmpty(t, op["operationId"])
}

func TestNormalizeOperationID_noSummaryFallsBack(t *testing.T) {
	seen := map[string]int{}
	op := map[string]any{}
	normalizeOperationID(op, "get", "/api/v1/z", seen)
	require.NotEmpty(t, op["operationId"])
}

func TestPostProcessSpec_marshalError(t *testing.T) {
	prev := specMarshalJSON
	t.Cleanup(func() { specMarshalJSON = prev })
	specMarshalJSON = func(any) ([]byte, error) { return nil, errors.New("encode failed") }
	dir := t.TempDir()
	p := filepath.Join(dir, "spec.json")
	require.NoError(t, os.WriteFile(p, []byte(`{"openapi":"3.0.0","paths":{}}`), 0o644))
	assert.Error(t, PostProcessSpec(p))
}

func TestAddParameterExamplesAndConstraints(t *testing.T) {
	op := map[string]any{
		"parameters": []any{
			map[string]any{
				"in": "query", "name": "page", "example": 1, "schema": map[string]any{"type": "string"},
			},
			map[string]any{
				"in": "query", "name": "page_size", "schema": map[string]any{"type": "string"},
			},
			map[string]any{
				"in": "query", "name": "sort_direction", "schema": map[string]any{"type": "string"},
			},
			map[string]any{
				"in": "query", "name": "period", "schema": map[string]any{"type": "string", "default": "7d"},
			},
			"not-a-map",
			map[string]any{
				"in": "path", "name": "id",
			},
			map[string]any{
				"in": "query", "name": "org_id", "schema": map[string]any{"type": "string"},
			},
			map[string]any{
				"in": "query", "name": "limit", "schema": map[string]any{"type": "string", "default": 5},
			},
			map[string]any{
				"in": "query", "name": "noschema", // schema nil → skip
			},
		},
	}
	addParameterExamplesAndConstraints(op)
}

func TestAddRequestExamples(t *testing.T) {
	schemas := map[string]any{
		"Input": map[string]any{
			"type": "object", "properties": map[string]any{"a": map[string]any{"type": "string"}},
		},
	}
	op := map[string]any{
		"requestBody": map[string]any{
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{"$ref": "#/components/schemas/Input"},
				},
				"text/plain": "x",
				"skipExample": map[string]any{
					"example": 1, "schema": map[string]any{"$ref": "#/components/schemas/Input"},
				},
				"noRef": map[string]any{"schema": map[string]any{"type": "string"}},
			},
		},
	}
	addRequestExamples(op, schemas)
}

func TestAddSuccessExamples_get(t *testing.T) {
	schemas := map[string]any{
		"Out": map[string]any{"type": "object", "properties": map[string]any{
			"v": map[string]any{"type": "string"},
		}},
	}
	addSuccessExamples(map[string]any{"responses": nil}, "get", schemas) // nil responses, early return

	op2 := map[string]any{
		"responses": map[string]any{
			"200": map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{
						"example": 1,
						"schema":  map[string]any{"$ref": "#/components/schemas/Out"},
					},
					"text":  9,
					"plain": map[string]any{"schema": map[string]any{"type": "string"}},
				},
			},
		},
	}
	addSuccessExamples(op2, "get", schemas)
}

func TestAddSuccessExamples_post_201(t *testing.T) {
	schemas := map[string]any{
		"Out": map[string]any{"type": "string"},
	}
	op := map[string]any{
		"responses": map[string]any{
			"201": map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/Out"},
					},
				},
			},
		},
	}
	addSuccessExamples(op, "post", schemas)

	op2 := map[string]any{
		"responses": map[string]any{
			"200": map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/Out"},
					},
				},
			},
		},
	}
	addSuccessExamples(op2, "post", schemas) // no 201, uses 200
}

func TestStandardizeErrorResponses(t *testing.T) {
	op := map[string]any{
		"responses": map[string]any{
			"400": map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"type": "object"},
					},
					"bad": "skip",
				},
			},
		},
	}
	standardizeErrorResponses(op)
}

func TestEnsureErrorEnvelopeSchema(t *testing.T) {
	ensureErrorEnvelopeSchema(nil) // no-op
	m := map[string]any{}
	ensureErrorEnvelopeSchema(m)
	require.NotNil(t, m["ErrorEnvelope"])
}

func TestPostProcessSpec(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.json")
	// minimal: empty paths, nil components
	raw := `{"openapi":"3.0.0","paths":{}}`
	require.NoError(t, os.WriteFile(path, []byte(raw), 0o644))
	require.NoError(t, PostProcessSpec(path))
	out, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(out), "openapi")

	// not json
	bad := filepath.Join(dir, "bad.json")
	require.NoError(t, os.WriteFile(bad, []byte(`{`), 0o644))
	assert.Error(t, PostProcessSpec(bad))

	// missing file
	assert.Error(t, PostProcessSpec(filepath.Join(dir, "nope.json")))

	// write failure: specPath is a directory
	assert.Error(t, PostProcessSpec(dir))
}

func TestPostProcessSpec_rich(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.json")
	raw := `{
		"openapi": "3.0.0",
		"components": { "schemas": { "T": { "type": "string" } } },
		"paths": {
			"/api/v1/x": {
				"get": {
					"summary": "    ",
					"parameters": [
						{ "in": "query", "name": "user_id", "schema": { "type": "string" } }
					],
					"responses": {
						"200": { "content": { "application/json": { "schema": { "$ref": "#/components/schemas/T" } } } },
						"400": { "content": { "application/json": { "schema": {} } } }
					}
				}
			}
		}
	}`
	require.NoError(t, os.WriteFile(path, []byte(raw), 0o644))
	require.NoError(t, PostProcessSpec(path))
}

func TestPostProcessSpec_skip_bad_path_nodes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.json")
	// string instead of object for path; get value not a map; post op not a map
	raw := `{"openapi":"3.0.0","paths":{"/bad":"not-obj","/ok":{"get":"x","post":{"summary":"S"}}}}`
	require.NoError(t, os.WriteFile(path, []byte(raw), 0o644))
	require.NoError(t, PostProcessSpec(path))
}

func TestPostProcessSpecWithRetry(t *testing.T) {
	t.Run("timeout without file and negative window uses synthetic timeout", func(t *testing.T) {
		err := PostProcessSpecWithRetry(filepath.Join(t.TempDir(), "nope"), -time.Hour)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "timed out")
	})
	t.Run("success when file is valid on first read", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "s.json")
		require.NoError(t, os.WriteFile(p, []byte(`{"openapi":"3.0.0","paths":{}}`), 0o644))
		assert.NoError(t, PostProcessSpecWithRetry(p, 2*time.Second))
	})
	t.Run("file invalid json then eventually gives error", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "e.json")
		require.NoError(t, os.WriteFile(p, []byte(`{`), 0o644))
		err := PostProcessSpecWithRetry(p, 200*time.Millisecond)
		require.Error(t, err)
	})
	t.Run("stat keeps failing then returns last stat error", func(t *testing.T) {
		err := PostProcessSpecWithRetry(filepath.Join(t.TempDir(), "missing", "a.json"), 80*time.Millisecond)
		require.Error(t, err)
	})
	t.Run("valid after retry once", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "late.json")
		// use short first windows - write after small delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = os.WriteFile(p, []byte(`{"openapi":"3.0.0","paths":{}}`), 0o644)
		}()
		err := PostProcessSpecWithRetry(p, 2*time.Second)
		assert.NoError(t, err)
	})
}
