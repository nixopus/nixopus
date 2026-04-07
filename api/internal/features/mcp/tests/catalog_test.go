package tests

import (
	"testing"
	"time"

	mcp "github.com/nixopus/nixopus/api/internal/features/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tenSeconds = 10 * time.Second

// ─── GetProvider ─────────────────────────────────────────────────────────────

func TestGetProvider(t *testing.T) {
	t.Run("returns supabase provider", func(t *testing.T) {
		p := mcp.GetProvider("supabase")
		require.NotNil(t, p)
		assert.Equal(t, "supabase", p.ID)
		assert.Equal(t, "Supabase", p.Name)
	})

	t.Run("returns github provider", func(t *testing.T) {
		p := mcp.GetProvider("github")
		require.NotNil(t, p)
		assert.Equal(t, "github", p.ID)
	})

	t.Run("returns custom provider", func(t *testing.T) {
		p := mcp.GetProvider("custom")
		require.NotNil(t, p)
		assert.Equal(t, "custom", p.ID)
	})

	t.Run("returns nil for unknown id", func(t *testing.T) {
		p := mcp.GetProvider("does_not_exist")
		assert.Nil(t, p)
	})
}

// ─── Transport correctness ────────────────────────────────────────────────────

func TestCatalogTransportFields(t *testing.T) {
	t.Run("supabase uses HTTP transport", func(t *testing.T) {
		p := mcp.GetProvider("supabase")
		require.NotNil(t, p)
		assert.Equal(t, "http", p.Transport,
			"supabase must use streamable HTTP transport")
	})

	t.Run("github uses HTTP transport", func(t *testing.T) {
		p := mcp.GetProvider("github")
		require.NotNil(t, p)
		assert.Equal(t, "http", p.Transport)
	})

	t.Run("custom uses HTTP transport", func(t *testing.T) {
		p := mcp.GetProvider("custom")
		require.NotNil(t, p)
		assert.Equal(t, "http", p.Transport)
	})
}

// ─── Provider field definitions ───────────────────────────────────────────────

func TestCatalogProviderFields(t *testing.T) {
	t.Run("supabase access_token is required and mapped to Authorization header", func(t *testing.T) {
		p := mcp.GetProvider("supabase")
		require.NotNil(t, p)

		var tokenField *mcp.ProviderField
		for i := range p.Fields {
			if p.Fields[i].Key == "access_token" {
				tokenField = &p.Fields[i]
				break
			}
		}
		require.NotNil(t, tokenField, "access_token field must exist")
		assert.True(t, tokenField.Required)
		assert.Equal(t, "Authorization", tokenField.HeaderName)
		assert.Equal(t, "Bearer", tokenField.HeaderPrefix)
		assert.True(t, tokenField.Sensitive)
	})

	t.Run("supabase project_ref is an optional query param", func(t *testing.T) {
		p := mcp.GetProvider("supabase")
		require.NotNil(t, p)

		var refField *mcp.ProviderField
		for i := range p.Fields {
			if p.Fields[i].Key == "project_ref" {
				refField = &p.Fields[i]
				break
			}
		}
		require.NotNil(t, refField, "project_ref field must exist")
		assert.False(t, refField.Required)
		assert.True(t, refField.IsQueryParam)
	})

	t.Run("github token is required and mapped to Authorization header", func(t *testing.T) {
		p := mcp.GetProvider("github")
		require.NotNil(t, p)
		require.Len(t, p.Fields, 1)
		assert.Equal(t, "token", p.Fields[0].Key)
		assert.True(t, p.Fields[0].Required)
		assert.Equal(t, "Authorization", p.Fields[0].HeaderName)
		assert.Equal(t, "Bearer", p.Fields[0].HeaderPrefix)
	})

	t.Run("custom provider has no predefined fields", func(t *testing.T) {
		p := mcp.GetProvider("custom")
		require.NotNil(t, p)
		assert.Empty(t, p.Fields,
			"custom provider requires no pre-defined fields; URL is set by the user")
	})
}

// ─── Catalog completeness ─────────────────────────────────────────────────────

func TestCatalogCompleteness(t *testing.T) {
	knownProviders := []string{"supabase", "github", "custom"}

	for _, id := range knownProviders {
		t.Run(id+" is in catalog", func(t *testing.T) {
			p := mcp.GetProvider(id)
			require.NotNil(t, p)
			assert.NotEmpty(t, p.Name)
			assert.NotEmpty(t, p.Description)
		})
	}

	t.Run("all catalog entries have a non-empty transport", func(t *testing.T) {
		for _, p := range mcp.Catalog {
			assert.NotEmpty(t, p.Transport, "provider %q has no transport set", p.ID)
		}
	})
}
