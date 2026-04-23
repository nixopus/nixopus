package openapi

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-fuego/fuego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionForMethod(t *testing.T) {
	for _, tc := range []struct{ m, want string }{
		{"get", "Get"},
		{"POST", "Create"},
		{"put", "Update"},
		{"pAtCh", "Patch"},
		{"delete", "Delete"},
		{"HEAD", "Use"},
	} {
		assert.Equal(t, tc.want, actionForMethod(tc.m), "method %q", tc.m)
	}
}

func TestFallbackSummary(t *testing.T) {
	s := fallbackSummary("GET", "/api/v1/machines")
	assert.Contains(t, s, "Get")
	assert.Contains(t, strings.ToLower(s), "machines")
	s2 := fallbackSummary("POST", "/api/v1/{id}")
	assert.NotEmpty(t, s2)
	// /api/v1 with no further segment leaves no resource; uses "endpoint"
	assert.Contains(t, strings.ToLower(fallbackSummary("GET", "/api/v1")), "endpoint")
}

func TestInferResourceName(t *testing.T) {
	assert.Equal(t, "", inferResourceName(""))
	assert.Equal(t, "machines", inferResourceName("/api/v1/machines"))
	assert.Equal(t, "items", inferResourceName("/api/v1/items/{id}"))
	// /api is a single path segment, last segment is "api" (not empty)
	assert.Equal(t, "api", inferResourceName("/api"))
	assert.Equal(t, "a", inferResourceName("/a/{id}"))
}

func TestInferPrimaryTag(t *testing.T) {
	assert.Equal(t, "machines", inferPrimaryTag("/api/v1/machines"))
	assert.Equal(t, "foo", inferPrimaryTag("foo"))
	assert.Equal(t, "api", inferPrimaryTag("/api"))
}

func TestBuildDescription(t *testing.T) {
	d := buildDescription("List widgets", "GET", "/api/v1/w")
	assert.Contains(t, d, "List widgets")
	assert.Contains(t, d, "bearer")
	assert.Contains(t, d, "Read-only")
	d2 := buildDescription("Create", "POST", "/api/v1/w")
	assert.Contains(t, d2, "mutate")
	pub := buildDescription("Health", "GET", "/api/v1/health/live")
	assert.Contains(t, pub, "Public")
}

func TestIsPublicPath(t *testing.T) {
	assert.True(t, isPublicPath("/api/v1/health"))
	assert.True(t, isPublicPath("/api/v1/health/ready"))
	assert.True(t, isPublicPath("/api/v1/webhook/hook"))
	assert.True(t, isPublicPath("/api/v1/auth/is-admin-registered"))
	assert.True(t, isPublicPath("/ws"))
	assert.False(t, isPublicPath("/api/v1/private"))
}

func TestSplitPath(t *testing.T) {
	assert.Nil(t, splitPath(""))
	assert.Nil(t, splitPath("///"))
	seg := splitPath("/api/v1/x")
	assert.Equal(t, []string{"api", "v1", "x"}, seg)
}

func TestIsPathParam(t *testing.T) {
	assert.True(t, isPathParam("{id}"))
	assert.False(t, isPathParam("id"))
}

func TestToLowerCamel(t *testing.T) {
	assert.Equal(t, "helloWorld", toLowerCamel("Hello World"))
	assert.Equal(t, "abc", toLowerCamel("Abc"))
	assert.Equal(t, "", toLowerCamel("   "))
}

func TestEnsureUniqueOperationID(t *testing.T) {
	a := ensureUniqueOperationID("once" + t.Name())
	b := ensureUniqueOperationID("once" + t.Name())
	assert.Equal(t, "once"+t.Name(), a)
	assert.Equal(t, "once"+t.Name()+"2", b)
}

func TestRouteContractOption(t *testing.T) {
	r := &fuego.BaseRoute{
		Method: "GET",
		Path:   "/api/v1/labels",
		Operation: &openapi3.Operation{
			Summary: "",
		},
	}
	opt := RouteContractOption()
	require.NotPanics(t, func() { opt(r) })
	require.NotEmpty(t, r.Operation.Summary)
	require.NotEmpty(t, r.Operation.OperationID)
	assert.Contains(t, r.Operation.Description, "Auth:")
	assert.Contains(t, r.Operation.Description, "Read-only")
	require.NotEmpty(t, r.Operation.Tags)
}

func TestRouteContractOption_addsTagWhenEmpty(t *testing.T) {
	r := &fuego.BaseRoute{
		Method:    "GET",
		Path:      "/api/v1/tagged",
		Operation: &openapi3.Operation{Summary: "Already set", Tags: nil},
	}
	RouteContractOption()(r)
	assert.Equal(t, "tagged", r.Operation.Tags[0])
}

func TestRouteContractOption_doesNotOverrideExistingTags(t *testing.T) {
	r := &fuego.BaseRoute{
		Method:    "GET",
		Path:      "/api/v1/x",
		Operation: &openapi3.Operation{Summary: "X", Tags: []string{"keepMe"}},
	}
	RouteContractOption()(r)
	require.Equal(t, []string{"keepMe"}, r.Operation.Tags)
}

func TestRouteContractOption_punctuationSummaryFallsBackToPath(t *testing.T) {
	r := &fuego.BaseRoute{
		Method:    "GET",
		Path:      "/api/v1/widgets",
		Operation: &openapi3.Operation{Summary: "@@@"},
	}
	RouteContractOption()(r)
	// toLowerCamel("@@@") is empty; opID comes from fallbackSummary(Get + resource)
	require.NotEmpty(t, r.Operation.OperationID)
	assert.Regexp(t, `^getWidgets\d*$`, r.Operation.OperationID) // may be suffixed if duplicate global counts
}

func TestRouteContractOption_noInferredTagForBareRoot(t *testing.T) {
	r := &fuego.BaseRoute{
		Method:    "GET",
		Path:      "/",
		Operation: &openapi3.Operation{Summary: "Root list"},
	}
	RouteContractOption()(r)
	assert.Empty(t, r.Operation.Tags)
}
