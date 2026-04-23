package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ufixture() *types.User {
	return &types.User{
		ID:        uuid.New(),
		Name:      "N",
		Email:     "n@e.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestGetUser(t *testing.T) {
	t.Run("returns user from context", func(t *testing.T) {
		u := ufixture()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(context.WithValue(r.Context(), types.UserContextKey, u))
		rr := httptest.NewRecorder()
		out := GetUser(rr, r)
		require.Same(t, u, out)
		assert.Equal(t, 200, rr.Code) // no error written
	})

	t.Run("wrong type sends error and returns nil", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(context.WithValue(r.Context(), types.UserContextKey, "not-a-user"))
		rr := httptest.NewRecorder()
		assert.Nil(t, GetUser(rr, r))
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestGetOrganizationID(t *testing.T) {
	t.Run("nil in context", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, uuid.Nil, GetOrganizationID(r))
	})
	t.Run("valid string uuid", func(t *testing.T) {
		id := uuid.New()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(context.WithValue(r.Context(), types.OrganizationIDKey, id.String()))
		assert.Equal(t, id, GetOrganizationID(r))
	})
	t.Run("invalid string", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(context.WithValue(r.Context(), types.OrganizationIDKey, "not-a-uuid"))
		assert.Equal(t, uuid.Nil, GetOrganizationID(r))
	})
	t.Run("uuid value", func(t *testing.T) {
		id := uuid.New()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(context.WithValue(r.Context(), types.OrganizationIDKey, id))
		assert.Equal(t, id, GetOrganizationID(r))
	})
	t.Run("other type is nil", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(context.WithValue(r.Context(), types.OrganizationIDKey, 123))
		assert.Equal(t, uuid.Nil, GetOrganizationID(r))
	})
}
