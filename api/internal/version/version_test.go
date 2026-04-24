package version_manager

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersion(t *testing.T) {
	v := NewVersion("v2")
	assert.Equal(t, "v2", v.Version)
	assert.Equal(t, "/api/v2", v.Path)
}

func TestGetVersionFromRequest(t *testing.T) {
	cases := []struct {
		name  string
		req   *http.Request
		want  string
		notes string
	}{
		{
			"header wins over query and path",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "http://example.com/api/v3/foo?api-version=ignored", nil)
				r.Header.Set(VersionHeader, "v9")
				return r
			}(),
			"v9",
			"",
		},
		{
			"query wins over path when header absent",
			httptest.NewRequest(http.MethodGet, "http://example.com/api/v3/ignored?api-version=v8", nil),
			"v8",
			"",
		},
		{
			"header",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "http://example.com/any", nil)
				r.Header.Set(VersionHeader, "v9")
				return r
			}(),
			"v9",
			"",
		},
		{
			"query",
			httptest.NewRequest(http.MethodGet, "http://example.com?api-version=v8", nil),
			"v8",
			"",
		},
		{
			"path segment",
			httptest.NewRequest(http.MethodGet, "http://example.com/api/v3/foo", nil),
			"v3",
			"",
		},
		{
			"empty query value falls through to path",
			httptest.NewRequest(http.MethodGet, "/api/v2?api-version=", nil),
			"v2",
			"Get returns empty, path gives v2",
		},
		{
			"trailing slash path still parses version",
			httptest.NewRequest(http.MethodGet, "http://example.com/api/v2/", nil),
			"v2",
			"",
		},
		{
			"default when path is /api only (no /api/<ver>)",
			httptest.NewRequest(http.MethodGet, "http://example.com/api", nil),
			DefaultVersion,
			"",
		},
		{
			"default when not api path",
			httptest.NewRequest(http.MethodGet, "http://example.com/health", nil),
			DefaultVersion,
			"",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.notes != "" {
				t.Log(tc.notes)
			}
			assert.Equal(t, tc.want, GetVersionFromRequest(tc.req), tc.notes)
		})
	}
}

func TestIsValidVersion(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"", false},
		{"v1", true},
		{"   ", true}, // non-empty string is "valid" per current rule
		{"v0", true},
	} {
		assert.Equal(t, tc.want, IsValidVersion(tc.in), "in=%q", tc.in)
	}
}

func TestGetVersionedPath(t *testing.T) {
	for _, tc := range []struct {
		version, endpoint, want string
	}{
		{"v1", "/users", "/api/v1/users"},
		{"v1", "users", "/api/v1users"}, // endpoint is concatenated as-is
		{"v2", "/a/b", "/api/v2/a/b"},
	} {
		assert.Equal(t, tc.want, GetVersionedPath(tc.version, tc.endpoint),
			"GetVersionedPath(%q,%q)", tc.version, tc.endpoint)
	}
}

func TestVersionMiddleware(t *testing.T) {
	t.Run("invalid path version uses default in context and header", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			v, ok := r.Context().Value("api_version").(string)
			require.True(t, ok)
			assert.Equal(t, DefaultVersion, v)
		})
		req := httptest.NewRequest(http.MethodGet, "http://example.com/api/", nil)
		rr := httptest.NewRecorder()
		VersionMiddleware(next).ServeHTTP(rr, req)
		assert.True(t, called)
		assert.Equal(t, DefaultVersion, rr.Header().Get(VersionHeader))
	})

	t.Run("preserves explicit non-default version from header", func(t *testing.T) {
		var got string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got, _ = r.Context().Value("api_version").(string)
		})
		req := httptest.NewRequest(http.MethodGet, "http://example.com/any", nil)
		req.Header.Set(VersionHeader, "v2")
		rr := httptest.NewRecorder()
		VersionMiddleware(next).ServeHTTP(rr, req)
		assert.Equal(t, "v2", got)
		assert.Equal(t, "v2", rr.Header().Get(VersionHeader))
	})
}

func TestNewVersionDocumentation(t *testing.T) {
	d := NewVersionDocumentation()
	require.Len(t, d.Versions, 1)
	assert.Equal(t, CurrentVersion, d.Versions[0].Version)
	assert.Equal(t, "active", d.Versions[0].Status)
	assert.WithinDuration(t, time.Now(), d.Versions[0].ReleaseDate, 5*time.Second)
}

func TestVersionDocumentation_AddVersion(t *testing.T) {
	d := &VersionDocumentation{}
	require.NoError(t, d.AddVersion(VersionInfo{Version: "v1", Status: "active"}))
	err := d.AddVersion(VersionInfo{Version: "v1", Status: "retired"})
	require.Error(t, err)
	assert.Equal(t, "version v1 already exists", err.Error())
}

func TestVersionDocumentation_UpdateVersion(t *testing.T) {
	d := &VersionDocumentation{Versions: []VersionInfo{{Version: "v1", Status: "active"}}}
	updated := VersionInfo{Version: "v1", Status: "deprecated"}
	require.NoError(t, d.UpdateVersion(updated))
	assert.Equal(t, "deprecated", d.Versions[0].Status)
	err := d.UpdateVersion(VersionInfo{Version: "missing", Status: "active"})
	require.Error(t, err)
	assert.Equal(t, "version missing not found", err.Error())
}

func TestVersionDocumentation_GetVersion(t *testing.T) {
	d := &VersionDocumentation{Versions: []VersionInfo{{Version: "v1", Status: "active"}}}
	got, err := d.GetVersion("v1")
	require.NoError(t, err)
	assert.Equal(t, "v1", got.Version)
	_, err = d.GetVersion("nope")
	require.Error(t, err)
	assert.Equal(t, "version nope not found", err.Error())
}

func TestVersionDocumentation_SaveLoad(t *testing.T) {
	d := &VersionDocumentation{Versions: []VersionInfo{
		{
			Version:     "v1",
			Status:      "active",
			ReleaseDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			EndOfLife:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Changes:     []string{"a", "b"},
		},
	}}
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "doc.json")
	require.NoError(t, d.Save(path))
	loaded := &VersionDocumentation{}
	require.NoError(t, loaded.Load(path))
	assert.Equal(t, d.Versions, loaded.Versions)
}

func TestVersionDocumentation_Load_errors(t *testing.T) {
	d := &VersionDocumentation{}
	err := d.Load(filepath.Join(t.TempDir(), "missing.json"))
	require.Error(t, err)

	bad := filepath.Join(t.TempDir(), "bad.json")
	require.NoError(t, os.WriteFile(bad, []byte("not json"), 0o644))
	err = d.Load(bad)
	require.Error(t, err)
}

func TestVersionDocumentation_Save_writeError(t *testing.T) {
	d := &VersionDocumentation{Versions: []VersionInfo{{Version: "v1", Status: "active"}}}
	// path is a directory: WriteFile should fail
	err := d.Save(t.TempDir())
	require.Error(t, err)
}

func TestVersionDocumentation_Save_marshalError(t *testing.T) {
	d := &VersionDocumentation{
		Versions:       []VersionInfo{{Version: "v1", Status: "active"}},
		ForceJSONError: new(JSONMarshalFailer),
	}
	err := d.Save(filepath.Join(t.TempDir(), "doc.json"))
	require.Error(t, err)
}

func TestVersionDocumentation_Save_mkdirError(t *testing.T) {
	d := &VersionDocumentation{Versions: []VersionInfo{{Version: "v1", Status: "active"}}}
	td := t.TempDir()
	filePath := filepath.Join(td, "block")
	require.NoError(t, os.WriteFile(filePath, []byte("x"), 0o644))
	// MkdirAll cannot create a directory at filePath/child
	path := filepath.Join(filePath, "doc.json")
	err := d.Save(path)
	require.Error(t, err)
}

func TestVersionDocumentation_GetVersionStatus(t *testing.T) {
	d := &VersionDocumentation{Versions: []VersionInfo{{Version: "v1", Status: "retired"}}}
	st, err := d.GetVersionStatus("v1")
	require.NoError(t, err)
	assert.Equal(t, "retired", st)
	_, err = d.GetVersionStatus("missing")
	require.Error(t, err)
}

func TestVersionDocumentation_statusHelpers(t *testing.T) {
	d := &VersionDocumentation{Versions: []VersionInfo{
		{Version: "a", Status: "active"},
		{Version: "b", Status: "deprecated"},
		{Version: "c", Status: "retired"},
	}}
	assert.True(t, d.IsVersionActive("a"))
	assert.False(t, d.IsVersionActive("b"))
	assert.False(t, d.IsVersionActive("unknown"))

	assert.True(t, d.IsVersionDeprecated("b"))
	assert.False(t, d.IsVersionDeprecated("a"))
	assert.False(t, d.IsVersionDeprecated("unknown"))

	assert.True(t, d.IsVersionRetired("c"))
	assert.False(t, d.IsVersionRetired("a"))
	assert.False(t, d.IsVersionRetired("unknown"))
}

func TestMigrationHandler_passthrough(t *testing.T) {
	m := NewMigrationHandler()
	r := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.Same(t, r, m.MigrateRequest(r, "v0", "v1"))
	rr := httptest.NewRecorder()
	assert.Same(t, rr, m.MigrateResponse(rr, "v0", "v1"))
}

func TestMigrationMiddleware_sameVersion(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/"+CurrentVersion+"/x", nil)
	MigrationMiddleware(next).ServeHTTP(httptest.NewRecorder(), req)
	assert.True(t, called)
}

func TestMigrationMiddleware_differentVersion(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	req := httptest.NewRequest(http.MethodGet, "http://example.com/any", nil)
	req.Header.Set(VersionHeader, "v0")
	MigrationMiddleware(next).ServeHTTP(httptest.NewRecorder(), req)
	assert.True(t, called)
}
