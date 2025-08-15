package tests

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestGetGithubRepositoryBranches(t *testing.T) {
	userID := uuid.New().String()
	repositoryName := "test-user/test-repo"

	validGithubConnector := shared_types.GithubConnector{
		ID:             uuid.New(),
		AppID:          "12345",
		Slug:           "test-app",
		Pem:            generateTestPrivateKey(),
		ClientID:       "test-client-id",
		ClientSecret:   "test-client-secret",
		WebhookSecret:  "test-webhook-secret",
		InstallationID: "67890",
		UserID:         uuid.MustParse(userID),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	expectedBranches := []shared_types.GithubRepositoryBranch{
		{
			Name: "main",
			Commit: struct {
				Sha string `json:"sha"`
				URL string `json:"url"`
			}{
				Sha: "abc123def456",
				URL: "https://api.github.com/repos/test-user/test-repo/commits/abc123def456",
			},
			Protected: true,
		},
		{
			Name: "develop",
			Commit: struct {
				Sha string `json:"sha"`
				URL string `json:"url"`
			}{
				Sha: "def456ghi789",
				URL: "https://api.github.com/repos/test-user/test-repo/commits/def456ghi789",
			},
			Protected: false,
		},
	}

	tests := []struct {
		name             string
		userID           string
		repositoryName   string
		mockStorageSetup func(*MockGithubConnectorStorage)
		mockServerSetup  func() *httptest.Server
		expectedBranches []shared_types.GithubRepositoryBranch
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name:           "Successfully retrieve repository branches",
			userID:         userID,
			repositoryName: repositoryName,
			mockStorageSetup: func(mockStorage *MockGithubConnectorStorage) {
				mockStorage.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{validGithubConnector}, nil).Once()
			},
			mockServerSetup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/app/installations/67890/access_tokens" && r.Method == "POST" {
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(map[string]string{"token": "test-access-token"})
						return
					}
					if r.URL.Path == "/repos/test-user/test-repo/branches" && r.Method == "GET" {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode(expectedBranches)
						return
					}
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectedBranches: expectedBranches,
			expectedError:    false,
		},
		{
			name:           "No connectors found for user",
			userID:         userID,
			repositoryName: repositoryName,
			mockStorageSetup: func(mockStorage *MockGithubConnectorStorage) {
				mockStorage.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{}, nil).Once()
			},
			mockServerSetup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectedBranches: []shared_types.GithubRepositoryBranch{},
			expectedError:    false,
		},
		{
			name:           "Storage error when getting connectors",
			userID:         userID,
			repositoryName: repositoryName,
			mockStorageSetup: func(mockStorage *MockGithubConnectorStorage) {
				mockStorage.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{}, fmt.Errorf("database connection error")).Once()
			},
			mockServerSetup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectedBranches: nil,
			expectedError:    true,
			expectedErrorMsg: "database connection error",
		},
		{
			name:           "Failed to get installation token",
			userID:         userID,
			repositoryName: repositoryName,
			mockStorageSetup: func(mockStorage *MockGithubConnectorStorage) {
				mockStorage.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{validGithubConnector}, nil).Once()
			},
			mockServerSetup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/app/installations/67890/access_tokens" && r.Method == "POST" {
						w.WriteHeader(http.StatusUnauthorized)
						json.NewEncoder(w).Encode(map[string]string{"message": "Bad credentials"})
						return
					}
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectedBranches: nil,
			expectedError:    true,
			expectedErrorMsg: "Failed to get installation token",
		},
		{
			name:           "GitHub API error for branches",
			userID:         userID,
			repositoryName: repositoryName,
			mockStorageSetup: func(mockStorage *MockGithubConnectorStorage) {
				mockStorage.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{validGithubConnector}, nil).Once()
			},
			mockServerSetup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/app/installations/67890/access_tokens" && r.Method == "POST" {
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(map[string]string{"token": "test-access-token"})
						return
					}
					if r.URL.Path == "/repos/test-user/test-repo/branches" && r.Method == "GET" {
						w.WriteHeader(http.StatusNotFound)
						json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
						return
					}
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectedBranches: nil,
			expectedError:    true,
			expectedErrorMsg: "GitHub API error: 404 Not Found",
		},
		{
			name:           "Invalid JSON response from GitHub",
			userID:         userID,
			repositoryName: repositoryName,
			mockStorageSetup: func(mockStorage *MockGithubConnectorStorage) {
				mockStorage.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{validGithubConnector}, nil).Once()
			},
			mockServerSetup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/app/installations/67890/access_tokens" && r.Method == "POST" {
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(map[string]string{"token": "test-access-token"})
						return
					}
					if r.URL.Path == "/repos/test-user/test-repo/branches" && r.Method == "GET" {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("invalid json"))
						return
					}
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectedBranches: nil,
			expectedError:    true,
		},
		{
			name:           "Empty branches response",
			userID:         userID,
			repositoryName: repositoryName,
			mockStorageSetup: func(mockStorage *MockGithubConnectorStorage) {
				mockStorage.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{validGithubConnector}, nil).Once()
			},
			mockServerSetup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/app/installations/67890/access_tokens" && r.Method == "POST" {
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(map[string]string{"token": "test-access-token"})
						return
					}
					if r.URL.Path == "/repos/test-user/test-repo/branches" && r.Method == "GET" {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode([]shared_types.GithubRepositoryBranch{})
						return
					}
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectedBranches: []shared_types.GithubRepositoryBranch{},
			expectedError:    false,
		},
		{
			name:           "JWT generation fails with invalid PEM",
			userID:         userID,
			repositoryName: repositoryName,
			mockStorageSetup: func(mockStorage *MockGithubConnectorStorage) {
				invalidConnector := validGithubConnector
				invalidConnector.Pem = "invalid-pem-data"
				mockStorage.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{invalidConnector}, nil).Once()
			},
			mockServerSetup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectedBranches: nil,
			expectedError:    true,
			expectedErrorMsg: "failed to generate app JWT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockGithubConnectorStorage()
			if tt.mockStorageSetup != nil {
				tt.mockStorageSetup(mockStorage)
			}

			mockServer := tt.mockServerSetup()
			defer mockServer.Close()

			originalURL := "https://api.github.com"
			defer func() {
				service.SetGithubAPIBaseURL(originalURL)
			}()
			service.SetGithubAPIBaseURL(mockServer.URL)

			svc := service.NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockStorage)

			branches, err := svc.GetGithubRepositoryBranches(tt.userID, tt.repositoryName)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
				assert.Nil(t, branches)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBranches, branches)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestGetGithubRepositoryBranchesRequestHeaders(t *testing.T) {
	userID := uuid.New().String()
	repositoryName := "test-user/test-repo"

	validGithubConnector := shared_types.GithubConnector{
		ID:             uuid.New(),
		AppID:          "12345",
		Slug:           "test-app",
		Pem:            generateTestPrivateKey(),
		ClientID:       "test-client-id",
		ClientSecret:   "test-client-secret",
		WebhookSecret:  "test-webhook-secret",
		InstallationID: "67890",
		UserID:         uuid.MustParse(userID),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	var capturedHeaders http.Header
	var capturedURL string
	var capturedMethod string

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app/installations/67890/access_tokens" && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"token": "test-access-token"})
			return
		}
		if r.URL.Path == "/repos/test-user/test-repo/branches" && r.Method == "GET" {
			capturedHeaders = r.Header
			capturedURL = r.URL.String()
			capturedMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]shared_types.GithubRepositoryBranch{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	originalURL := "https://api.github.com"
	defer func() {
		service.SetGithubAPIBaseURL(originalURL)
	}()
	service.SetGithubAPIBaseURL(mockServer.URL)

	mockStorage := NewMockGithubConnectorStorage()
	mockStorage.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{validGithubConnector}, nil).Once()

	svc := service.NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockStorage)

	_, err := svc.GetGithubRepositoryBranches(userID, repositoryName)

	assert.NoError(t, err)
	assert.Equal(t, "GET", capturedMethod)
	assert.Equal(t, "/repos/test-user/test-repo/branches", capturedURL)
	assert.Equal(t, "token test-access-token", capturedHeaders.Get("Authorization"))
	assert.Equal(t, "application/vnd.github.v3+json", capturedHeaders.Get("Accept"))
	assert.Equal(t, "nixopus", capturedHeaders.Get("User-Agent"))

	mockStorage.AssertExpectations(t)
}

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
