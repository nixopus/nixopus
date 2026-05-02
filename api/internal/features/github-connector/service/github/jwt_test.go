package gh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func generateTestPrivateKey() string {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate test private key: %v", err))
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	return string(privateKeyPEM)
}

func TestInstallationToken_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	_, err := InstallationToken("a.b.c", "bad-install-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "installation not found")
}

func TestInstallationToken_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	_, err := InstallationToken("a.b.c", "install-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestInstallationToken_Forbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	_, err := InstallationToken("a.b.c", "install-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestInstallationToken_OtherError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("service unavailable"))
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	_, err := InstallationToken("a.b.c", "install-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to get installation token")
}

func TestInstallationToken_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	_, err := InstallationToken("a.b.c", "install-id")
	assert.Error(t, err)
}

func TestGenerateJwt_NilConnector_NoConfig(t *testing.T) {
	token := GenerateJwt(nil)
	assert.Equal(t, "", token)
}

func TestGenerateJwt_EmptyPem_NoConfig(t *testing.T) {
	token := GenerateJwt(&shared_types.GithubConnector{AppID: "123", Pem: ""})
	assert.Equal(t, "", token)
}

func TestGenerateJwt_InvalidPem(t *testing.T) {
	token := GenerateJwt(&shared_types.GithubConnector{AppID: "123", Pem: "not-a-pem"})
	assert.Equal(t, "", token)
}

func TestGenerateJwt_ValidConnector(t *testing.T) {
	token := GenerateJwt(&shared_types.GithubConnector{
		AppID: "123456",
		Pem:   generateTestPrivateKey(),
	})
	assert.NotEmpty(t, token)
}
