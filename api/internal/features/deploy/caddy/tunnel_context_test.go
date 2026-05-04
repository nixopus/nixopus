package caddy

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_orgIDFromContext(t *testing.T) {
	org := uuid.MustParse("33333333-4444-5555-6666-777777777777")

	assert.Equal(t, uuid.Nil, orgIDFromContext(context.Background()))

	s := org.String()
	ctx := context.WithValue(context.Background(), shared_types.OrganizationIDKey, s)
	assert.Equal(t, org, orgIDFromContext(ctx))

	ctx = context.WithValue(context.Background(), shared_types.OrganizationIDKey, org)
	assert.Equal(t, org, orgIDFromContext(ctx))

	ctx = context.WithValue(context.Background(), shared_types.OrganizationIDKey, "not-a-uuid")
	assert.Equal(t, uuid.Nil, orgIDFromContext(ctx))

	ctx = context.WithValue(context.Background(), shared_types.OrganizationIDKey, 42)
	assert.Equal(t, uuid.Nil, orgIDFromContext(ctx))
}

func Test_getCaddyPort_defaultsAndValidation(t *testing.T) {
	orig := config.AppConfig
	t.Cleanup(func() { config.AppConfig = orig })

	config.AppConfig.Proxy.CaddyPort = ""
	port, err := getCaddyPort()
	require.NoError(t, err)
	assert.Equal(t, defaultCaddyPort, port)

	config.AppConfig.Proxy.CaddyPort = "9555"
	port, err = getCaddyPort()
	require.NoError(t, err)
	assert.Equal(t, "9555", port)

	config.AppConfig.Proxy.CaddyPort = "not-int"
	_, err = getCaddyPort()
	require.Error(t, err)
}
