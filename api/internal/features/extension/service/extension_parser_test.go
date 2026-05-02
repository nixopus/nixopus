package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	exttypes "github.com/nixopus/nixopus/api/internal/features/extension/types"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validMetadataYAML(id string) string {
	return `metadata:
  id: ` + id + `
  name: Test Extension
  description: A test description
  author: tester
  icon: icon.png
  category: Utility
  type: install
  version: 1.0.0
  isVerified: false
  featured: false
variables:
  port:
    type: integer
    description: Port number
    default: 8080
    is_required: true
    validation_pattern: ""
`
}

func TestExtensionYAMLParser_parseExtensionFile_readError(t *testing.T) {
	p := newExtensionYAMLParser()
	_, _, err := p.parseExtensionFile("/nonexistent/path/metadata.yaml")
	require.Error(t, err)
}

func TestExtensionYAMLParser_parseExtensionFile_invalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(path, []byte("not: yaml: [[["), 0o600))

	p := newExtensionYAMLParser()
	_, _, err := p.parseExtensionFile(path)
	require.Error(t, err)
}

func TestExtensionYAMLParser_parseExtensionFile_validationError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`metadata:
  id: ""
  name: x
`), 0o600))

	p := newExtensionYAMLParser()
	_, _, err := p.parseExtensionFile(path)
	require.Error(t, err)
}

func TestExtensionYAMLParser_parseExtensionFile_success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.yaml")
	require.NoError(t, os.WriteFile(path, []byte(validMetadataYAML("svc-test-parse-001")), 0o600))

	p := newExtensionYAMLParser()
	ext, vars, err := p.parseExtensionFile(path)
	require.NoError(t, err)
	require.NotNil(t, ext)
	assert.Equal(t, "svc-test-parse-001", ext.ExtensionID)
	assert.Equal(t, types.ExtensionCategoryUtility, ext.Category)
	require.Len(t, vars, 1)
	assert.Equal(t, "port", vars[0].VariableName)
}

func TestExtensionYAMLParser_convertToExtension_marshalFallback(t *testing.T) {
	p := newExtensionYAMLParser()
	extYAML := &exttypes.ExtensionYAML{
		Metadata: exttypes.ExtensionMetadata{
			ID:          "svc-marshal-fail",
			Name:        "N",
			Description: "D",
			Author:      "A",
			Icon:        "I",
			Category:    "Utility",
			Type:        "run",
			Version:     "",
		},
		Variables: map[string]exttypes.ExtensionVariable{
			"x": {Type: "string", Default: make(chan int)},
		},
	}
	ext := p.convertToExtension(extYAML, "raw")
	require.NotNil(t, ext)
	assert.JSONEq(t, "{}", ext.ParsedContent)
	assert.Equal(t, types.ValidationStatusValid, ext.ValidationStatus)
}

func TestExtensionYAMLParser_convertToVariables_marshalDefaultFallback(t *testing.T) {
	p := newExtensionYAMLParser()
	extYAML := &exttypes.ExtensionYAML{
		Metadata: exttypes.ExtensionMetadata{
			ID: "id", Name: "n", Description: "d", Author: "a", Icon: "i", Category: "Utility", Type: "install",
		},
		Variables: map[string]exttypes.ExtensionVariable{
			"baddef": {Type: "string", Default: make(chan int), Description: "d"},
		},
	}
	vars := p.convertToVariables(extYAML, "id")
	require.Len(t, vars, 1)
	assert.Equal(t, "null", string(vars[0].DefaultValue))
}

func TestExtensionYAMLParser_loadExtensionsFromDirectory_readError(t *testing.T) {
	p := newExtensionYAMLParser()
	_, _, err := p.loadExtensionsFromDirectory("/this/path/should/not/exist/for/test")
	require.Error(t, err)
}

func TestExtensionYAMLParser_loadExtensionsFromDirectory_empty(t *testing.T) {
	p := newExtensionYAMLParser()
	dir := t.TempDir()
	exts, vars, err := p.loadExtensionsFromDirectory(dir)
	require.NoError(t, err)
	assert.Empty(t, exts)
	assert.Empty(t, vars)
}

func TestExtensionYAMLParser_loadExtensionsFromDirectory_parseFailure(t *testing.T) {
	base := t.TempDir()
	sub := filepath.Join(base, "broken")
	require.NoError(t, os.Mkdir(sub, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "metadata.yaml"), []byte("metadata: not-a-map"), 0o600))

	p := newExtensionYAMLParser()
	_, _, err := p.loadExtensionsFromDirectory(base)
	require.Error(t, err)
}

func TestExtensionYAMLParser_loadExtensionsFromDirectory_success(t *testing.T) {
	base := t.TempDir()
	sub := filepath.Join(base, "good")
	require.NoError(t, os.Mkdir(sub, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "metadata.yaml"), []byte(validMetadataYAML("svc-dir-ok-001")), 0o600))

	p := newExtensionYAMLParser()
	exts, varss, err := p.loadExtensionsFromDirectory(base)
	require.NoError(t, err)
	require.Len(t, exts, 1)
	require.Len(t, varss, 1)
	assert.Equal(t, "svc-dir-ok-001", exts[0].ExtensionID)
}

func TestExtensionYAMLParser_loadExtensionsFromDirectory_skipsNondirEntries(t *testing.T) {
	base := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(base, "readme.txt"), []byte("x"), 0o600))
	sub := filepath.Join(base, "good")
	require.NoError(t, os.MkdirAll(sub, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "metadata.yaml"), []byte(validMetadataYAML("svc-skip-file-001")), 0o600))

	p := newExtensionYAMLParser()
	exts, varss, err := p.loadExtensionsFromDirectory(base)
	require.NoError(t, err)
	require.Len(t, exts, 1)
	require.Len(t, varss, 1)
}

func TestExtensionYAMLParser_validateExtension_errors(t *testing.T) {
	p := newExtensionYAMLParser()
	mk := func() exttypes.ExtensionYAML {
		return exttypes.ExtensionYAML{
			Metadata: exttypes.ExtensionMetadata{
				ID: "abc", Name: "n", Description: "d", Author: "a", Icon: "i", Category: "Utility", Type: "install", Version: "",
			},
			Variables: nil,
		}
	}

	cases := []struct {
		name string
		mut  func(*exttypes.ExtensionYAML)
	}{
		{"empty id", func(y *exttypes.ExtensionYAML) { y.Metadata.ID = "" }},
		{"empty name", func(y *exttypes.ExtensionYAML) { y.Metadata.Name = "" }},
		{"empty description", func(y *exttypes.ExtensionYAML) { y.Metadata.Description = "" }},
		{"empty author", func(y *exttypes.ExtensionYAML) { y.Metadata.Author = "" }},
		{"empty icon", func(y *exttypes.ExtensionYAML) { y.Metadata.Icon = "" }},
		{"empty category", func(y *exttypes.ExtensionYAML) { y.Metadata.Category = "" }},
		{"empty type", func(y *exttypes.ExtensionYAML) { y.Metadata.Type = "" }},
		{"bad type", func(y *exttypes.ExtensionYAML) { y.Metadata.Type = "wrong" }},
		{"invalid category", func(y *exttypes.ExtensionYAML) { y.Metadata.Category = "NotACat" }},
		{"bad extension id short", func(y *exttypes.ExtensionYAML) { y.Metadata.ID = "ab" }},
		{"bad extension id upper", func(y *exttypes.ExtensionYAML) { y.Metadata.ID = "abcD" }},
		{"bad version", func(y *exttypes.ExtensionYAML) { y.Metadata.Version = "v1" }},
		{"bad var name", func(y *exttypes.ExtensionYAML) {
			y.Variables = map[string]exttypes.ExtensionVariable{"9bad": {Type: "string"}}
		}},
		{"bad var type", func(y *exttypes.ExtensionYAML) {
			y.Variables = map[string]exttypes.ExtensionVariable{"ok": {Type: "blob"}}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			y := mk()
			tc.mut(&y)
			err := p.validateExtension(&y)
			require.Error(t, err, tc.name)
		})
	}
}

func TestExtensionYAMLParser_validateExtension_success(t *testing.T) {
	p := newExtensionYAMLParser()
	y := exttypes.ExtensionYAML{
		Metadata: exttypes.ExtensionMetadata{
			ID: "valid-id", Name: "n", Description: "d", Author: "a", Icon: "i", Category: "Social", Type: "run", Version: "1.2.3",
		},
		Variables: map[string]exttypes.ExtensionVariable{
			"_Port": {Type: "boolean", Description: "", Default: true, IsRequired: false, ValidationPattern: ""},
		},
	}
	require.NoError(t, p.validateExtension(&y))
}

func TestExtensionYAMLParser_helpers(t *testing.T) {
	p := newExtensionYAMLParser()

	assert.False(t, p.isValidCategory("__none__"))
	assert.False(t, p.isValidExtensionID("ab"))
	assert.False(t, p.isValidExtensionID(strings.Repeat("a", 51)))
	assert.False(t, p.isValidExtensionID("-abc"))
	assert.False(t, p.isValidExtensionID("abc-"))
	assert.False(t, p.isValidVersion("1.2"))
	assert.False(t, p.isValidVersion("1.2.x"))
	assert.False(t, p.isValidVersion("1..3"))
	assert.False(t, p.isValidVariableName(""))
	assert.False(t, p.isValidVariableName(string(make([]byte, 101))))
	assert.False(t, p.isValidVariableName("9x"))
	assert.False(t, p.isValidVariableName("a-b"))
	assert.False(t, p.isValidVariableType("float"))

	assert.True(t, p.isValidVariableName("A9_x"))
	assert.True(t, p.isValidVariableType("array"))
}
