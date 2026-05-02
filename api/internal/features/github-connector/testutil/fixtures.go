package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/google/uuid"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// GenerateTestPrivateKeyPEM returns a valid RSA private key in PEM PKCS1 form for tests.
func GenerateTestPrivateKeyPEM() string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("failed to generate test private key: %v", err))
	}
	b := x509.MarshalPKCS1PrivateKey(key)
	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: b}))
}

// MakeTestConnector returns a populated connector fixture with a valid PEM key.
func MakeTestConnector(userID uuid.UUID) shared_types.GithubConnector {
	return shared_types.GithubConnector{
		ID:             uuid.New(),
		AppID:          "12345",
		Slug:           "test-app",
		Pem:            GenerateTestPrivateKeyPEM(),
		ClientID:       "c",
		ClientSecret:   "cs",
		WebhookSecret:  "ws",
		InstallationID: "67890",
		UserID:         userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}
