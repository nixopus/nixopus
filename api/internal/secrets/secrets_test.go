package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestLoadSecretManagerConfig_minimal(t *testing.T) {
	t.Setenv("SECRET_MANAGER_ENABLED", "")
	t.Setenv("SECRET_MANAGER_TYPE", "")
	cfg := LoadSecretManagerConfig("mysvc")
	require.NotNil(t, cfg)
	assert.Equal(t, SecretManagerNone, cfg.Type)
	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.ServiceName)
}

func TestLoadSecretManagerConfig_fullWhenEnabled(t *testing.T) {
	t.Setenv("SECRET_MANAGER_ENABLED", "true")
	t.Setenv("SECRET_MANAGER_TYPE", "none")
	t.Setenv("SECRET_MANAGER_PROJECT_ID", "pid")
	t.Setenv("SECRET_MANAGER_ENVIRONMENT", "staging")
	t.Setenv("SECRET_MANAGER_SECRET_PATH", "/api")
	t.Setenv("INFISICAL_URL", "https://custom.example")
	t.Setenv("INFISICAL_TOKEN", "tok")

	cfg := LoadSecretManagerConfig("api")
	require.NotNil(t, cfg)
	assert.Equal(t, SecretManagerNone, cfg.Type)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "pid", cfg.ProjectID)
	assert.Equal(t, "staging", cfg.Environment)
	assert.Equal(t, "/api", cfg.SecretPath)
	assert.Equal(t, "api", cfg.ServiceName)
	assert.Equal(t, "https://custom.example", cfg.InfisicalURL)
	assert.Equal(t, "tok", cfg.InfisicalToken)
}

func TestLoadSecretManagerConfig_infisicalTypeNotEnabled(t *testing.T) {
	t.Setenv("SECRET_MANAGER_ENABLED", "false")
	t.Setenv("SECRET_MANAGER_TYPE", "INFISICAL")
	t.Setenv("SECRET_MANAGER_PROJECT_ID", "p")

	cfg := LoadSecretManagerConfig("svc")
	require.NotNil(t, cfg)
	assert.Equal(t, SecretManagerInfisical, cfg.Type)
	assert.False(t, cfg.Enabled)
	assert.Equal(t, "p", cfg.ProjectID)
	assert.Equal(t, "svc", cfg.ServiceName)
	assert.Equal(t, "prod", cfg.Environment)
	assert.Equal(t, "/", cfg.SecretPath)
	assert.Equal(t, "https://app.infisical.com", cfg.InfisicalURL)
}

func TestNewSecretManager(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		m, err := NewSecretManager(&SecretManagerConfig{Enabled: false, Type: SecretManagerInfisical})
		require.NoError(t, err)
		_, ok := m.(*NoOpSecretManager)
		assert.True(t, ok)
	})
	t.Run("type none", func(t *testing.T) {
		m, err := NewSecretManager(&SecretManagerConfig{Enabled: true, Type: SecretManagerNone})
		require.NoError(t, err)
		_, ok := m.(*NoOpSecretManager)
		assert.True(t, ok)
	})
	t.Run("infisical missing token", func(t *testing.T) {
		m, err := NewSecretManager(&SecretManagerConfig{
			Enabled: true, Type: SecretManagerInfisical, InfisicalToken: "",
		})
		assert.Nil(t, m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "INFISICAL_TOKEN")
	})
	t.Run("infisical ok", func(t *testing.T) {
		m, err := NewSecretManager(&SecretManagerConfig{
			Enabled: true, Type: SecretManagerInfisical, InfisicalToken: "t",
		})
		require.NoError(t, err)
		_, ok := m.(*InfisicalManager)
		assert.True(t, ok)
	})
	t.Run("unknown type noop", func(t *testing.T) {
		m, err := NewSecretManager(&SecretManagerConfig{Enabled: true, Type: SecretManagerType("other")})
		require.NoError(t, err)
		_, ok := m.(*NoOpSecretManager)
		assert.True(t, ok)
	})
}

func TestNormalizeEnvironmentName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"  DEV ", "dev"},
		{"development", "dev"},
		{"staging", "staging"},
		{"stage", "staging"},
		{"prod", "prod"},
		{"production", "prod"},
		{"custom", "custom"},
		{"", ""},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, normalizeEnvironmentName(tc.in))
		})
	}
}

func TestNoOpSecretManager(t *testing.T) {
	var n NoOpSecretManager
	ctx := context.Background()
	_, err := n.GetSecret(ctx, "k")
	require.Error(t, err)

	m, err := n.GetSecrets(ctx, "p")
	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Empty(t, m)
}

func TestNewInfisicalManager(t *testing.T) {
	cfg := &SecretManagerConfig{InfisicalURL: "https://x.example"}
	m := NewInfisicalManager(cfg)
	require.NotNil(t, m)
	assert.Equal(t, cfg, m.config)
	assert.Equal(t, "https://x.example", m.baseURL)
	require.NotNil(t, m.httpClient)
}

func TestInfisicalManager_GetSecrets_structured(t *testing.T) {
	payload := InfisicalSecretsResponse{
		Secrets: []InfisicalSecret{
			{SecretKey: "APP_FOO", SecretValue: "1"},
			{SecretKey: "APP_BAR", SecretValue: "2"},
			{SecretKey: "OTHER", SecretValue: "3"},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/secrets/raw", r.URL.Path)
		assert.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
		assert.Equal(t, "ws", r.URL.Query().Get("workspaceId"))
		assert.Equal(t, "prod", r.URL.Query().Get("environment"))
		assert.Equal(t, "/", r.URL.Query().Get("secretPath"))
		assert.Equal(t, "true", r.URL.Query().Get("recursive"))
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   srv.URL,
		InfisicalToken: "tok",
		ProjectID:      "ws",
		Environment:    "production",
		SecretPath:     "",
	})

	got, err := m.GetSecrets(context.Background(), "APP_")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"APP_FOO": "1", "APP_BAR": "2"}, got)

	all, err := m.GetSecrets(context.Background(), "")
	require.NoError(t, err)
	assert.Len(t, all, 3)
}

func TestInfisicalManager_GetSecrets_flatFallback(t *testing.T) {
	flat := map[string]string{"K1": "v1", "K2": "v2"}
	body, err := json.Marshal(flat)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   srv.URL,
		InfisicalToken: "tok",
		Environment:    "dev",
		SecretPath:     "/x",
	})
	m.config.ProjectID = ""

	got, err := m.GetSecrets(context.Background(), "K")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"K1": "v1", "K2": "v2"}, got)
}

func TestInfisicalManager_GetSecrets_flatOnlyJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"alpha":"beta"}`))
	}))
	defer srv.Close()

	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   srv.URL,
		InfisicalToken: "tok",
		Environment:    "staging",
		SecretPath:     "/",
	})

	got, err := m.GetSecrets(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"alpha": "beta"}, got)
}

func TestInfisicalManager_GetSecrets_notFoundEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   srv.URL,
		InfisicalToken: "tok",
		Environment:    "prod",
	})
	got, err := m.GetSecrets(context.Background(), "")
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestInfisicalManager_GetSecrets_nonOKError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()

	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   srv.URL,
		InfisicalToken: "tok",
		Environment:    "prod",
	})
	_, err := m.GetSecrets(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "boom")
}

func TestInfisicalManager_GetSecrets_requestCreateError(t *testing.T) {
	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   "http://a b",
		InfisicalToken: "tok",
		Environment:    "prod",
	})
	_, err := m.GetSecrets(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
}

func TestInfisicalManager_GetSecrets_doError(t *testing.T) {
	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   "http://example.com",
		InfisicalToken: "tok",
		Environment:    "prod",
	})
	m.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}

	_, err := m.GetSecrets(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch secrets")
	assert.Contains(t, err.Error(), "network down")
}

func TestInfisicalManager_GetSecrets_readBodyError(t *testing.T) {
	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   "http://example.com",
		InfisicalToken: "tok",
		Environment:    "prod",
	})
	m.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(iotest.ErrReader(errors.New("read fail"))),
		}, nil
	})}

	_, err := m.GetSecrets(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read response body")
}

func TestInfisicalManager_GetSecrets_invalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   srv.URL,
		InfisicalToken: "tok",
		Environment:    "prod",
	})
	_, err := m.GetSecrets(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse secrets response")
}

func TestInfisicalManager_GetSecrets_emptyEnvironment(t *testing.T) {
	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   "http://example.com",
		InfisicalToken: "tok",
		Environment:    "",
	})
	_, err := m.GetSecrets(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "environment is required")
}

func TestInfisicalManager_GetSecret_propagatesGetSecretsError(t *testing.T) {
	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   "http://example.com",
		InfisicalToken: "tok",
		Environment:    "prod",
	})
	m.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("upstream")
	})}

	_, err := m.GetSecret(context.Background(), "k")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upstream")
}

func TestInfisicalManager_GetSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"secrets":[{"secretKey":"X","secretValue":"y"}]}`))
	}))
	defer srv.Close()

	m := NewInfisicalManager(&SecretManagerConfig{
		InfisicalURL:   srv.URL,
		InfisicalToken: "tok",
		Environment:    "prod",
	})

	v, err := m.GetSecret(context.Background(), "X")
	require.NoError(t, err)
	assert.Equal(t, "y", v)

	_, err = m.GetSecret(context.Background(), "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestLoadSecretsIntoEnv(t *testing.T) {
	t.Run("nil manager", func(t *testing.T) {
		require.NoError(t, LoadSecretsIntoEnv(nil, ""))
	})

	t.Run("get secrets error", func(t *testing.T) {
		stub := &stubManager{err: errors.New("boom")}
		err := LoadSecretsIntoEnv(stub, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load secrets")
	})

	t.Run("success and setenv invalid key logs", func(t *testing.T) {
		stub := &stubManager{secrets: map[string]string{"VALID": "ok", "": "bad"}}
		t.Setenv("VALID", "")
		require.NoError(t, LoadSecretsIntoEnv(stub, ""))
		assert.Equal(t, "ok", os.Getenv("VALID"))
		_ = os.Unsetenv("VALID")
	})
}

type stubManager struct {
	secrets map[string]string
	err     error
}

func (s *stubManager) GetSecret(context.Context, string) (string, error) {
	return "", errors.New("not used")
}

func (s *stubManager) GetSecrets(context.Context, string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.secrets, nil
}
