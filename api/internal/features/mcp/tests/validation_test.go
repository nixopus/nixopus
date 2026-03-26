package tests

import (
	"testing"

	"github.com/nixopus/nixopus/api/internal/features/mcp/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── ValidateURL ─────────────────────────────────────────────────────────────

func TestValidateURL(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		wantErr error
	}{
		// Happy path
		{"valid https public host", "https://mcp.supabase.com/sse", nil},
		{"valid https with path and query", "https://api.example.com/mcp?ref=abc", nil},

		// Scheme enforcement
		{"http is rejected", "http://example.com/mcp", validation.ErrInvalidURL},
		{"no scheme", "example.com/mcp", validation.ErrInvalidURL},
		{"empty string", "", validation.ErrInvalidURL},

		// Private / loopback address SSRF mitigations
		{"loopback 127.0.0.1", "https://127.0.0.1/mcp", validation.ErrPrivateURL},
		{"private 10.x", "https://10.0.0.1/mcp", validation.ErrPrivateURL},
		{"private 192.168.x", "https://192.168.1.100/mcp", validation.ErrPrivateURL},
		{"private 172.16.x", "https://172.16.0.1/mcp", validation.ErrPrivateURL},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validation.ValidateURL(tc.url)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

// ─── ValidateCreateRequest ───────────────────────────────────────────────────

func TestValidateCreateRequest(t *testing.T) {
	validSupabaseCreds := map[string]string{"access_token": "tok_abc123"}

	cases := []struct {
		name    string
		req     *validation.CreateServerRequest
		wantErr error
	}{
		{
			name: "valid supabase server",
			req: &validation.CreateServerRequest{
				ProviderID:  "supabase",
				Name:        "My Supabase",
				Credentials: validSupabaseCreds,
				Enabled:     true,
			},
		},
		{
			name: "valid custom server",
			req: &validation.CreateServerRequest{
				ProviderID:  "custom",
				Name:        "My Custom MCP",
				Credentials: map[string]string{},
				CustomURL:   "https://mcp.mycompany.com/sse",
				Enabled:     true,
			},
		},
		{
			name:    "missing name",
			req:     &validation.CreateServerRequest{ProviderID: "supabase", Credentials: validSupabaseCreds},
			wantErr: validation.ErrNameRequired,
		},
		{
			name:    "whitespace-only name",
			req:     &validation.CreateServerRequest{Name: "   ", ProviderID: "supabase", Credentials: validSupabaseCreds},
			wantErr: validation.ErrNameRequired,
		},
		{
			name:    "missing provider_id",
			req:     &validation.CreateServerRequest{Name: "test", Credentials: validSupabaseCreds},
			wantErr: validation.ErrProviderRequired,
		},
		{
			name:    "unknown provider_id",
			req:     &validation.CreateServerRequest{Name: "test", ProviderID: "unknown_xyz", Credentials: map[string]string{}},
			wantErr: validation.ErrUnknownProvider,
		},
		{
			name: "custom without custom_url",
			req: &validation.CreateServerRequest{
				ProviderID:  "custom",
				Name:        "test",
				Credentials: map[string]string{},
			},
			wantErr: validation.ErrCustomURLRequired,
		},
		{
			name: "custom with non-https custom_url",
			req: &validation.CreateServerRequest{
				ProviderID:  "custom",
				Name:        "test",
				Credentials: map[string]string{},
				CustomURL:   "http://my-mcp.com/sse",
			},
			wantErr: validation.ErrInvalidURL,
		},
		{
			name: "supabase missing required access_token",
			req: &validation.CreateServerRequest{
				ProviderID:  "supabase",
				Name:        "test",
				Credentials: map[string]string{},
			},
			wantErr: validation.ErrMissingRequiredField,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validation.ValidateCreateRequest(tc.req)
			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

// ─── ValidateUpdateRequest ───────────────────────────────────────────────────

func TestValidateUpdateRequest(t *testing.T) {
	t.Run("valid update", func(t *testing.T) {
		err := validation.ValidateUpdateRequest(&validation.UpdateServerRequest{
			ID: "uuid-abc", Name: "Renamed Server",
		})
		assert.NoError(t, err)
	})

	t.Run("empty name rejected", func(t *testing.T) {
		err := validation.ValidateUpdateRequest(&validation.UpdateServerRequest{ID: "uuid-abc", Name: ""})
		assert.ErrorIs(t, err, validation.ErrNameRequired)
	})

	t.Run("custom_url http rejected", func(t *testing.T) {
		err := validation.ValidateUpdateRequest(&validation.UpdateServerRequest{
			ID: "uuid-abc", Name: "test", CustomURL: "http://bad.com/mcp",
		})
		assert.ErrorIs(t, err, validation.ErrInvalidURL)
	})

	t.Run("valid custom_url accepted", func(t *testing.T) {
		err := validation.ValidateUpdateRequest(&validation.UpdateServerRequest{
			ID: "uuid-abc", Name: "test", CustomURL: "https://ok.com/mcp",
		})
		assert.NoError(t, err)
	})
}
