package gh

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/testutil"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// GetRepositoryByID
// ---------------------------------------------------------------------------

func TestGetRepositoryByID_NoConnectors(t *testing.T) {
	userID := uuid.New().String()
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{}, nil)

	api := newAPI(mockSt)
	_, err := api.GetRepositoryByID(userID, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no connectors found")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoryByID_GetConnectorsError(t *testing.T) {
	userID := uuid.New().String()
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return(nil, assert.AnError)

	api := newAPI(mockSt)
	_, err := api.GetRepositoryByID(userID, 42)
	assert.Error(t, err)
	mockSt.AssertExpectations(t)
}

func TestGetRepositoryByID_InvalidPEM(t *testing.T) {
	userID := uuid.New()
	connector := makeConnector(userID)
	connector.Pem = "invalid"
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, err := api.GetRepositoryByID(userID.String(), 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate app JWT")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoryByID_APIError(t *testing.T) {
	userID := uuid.New()
	connector := makeConnector(userID)

	srv := tokenServer(func(mux *http.ServeMux) {
		mux.HandleFunc("/repositories/42", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
		})
	})
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, err := api.GetRepositoryByID(userID.String(), 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub API error")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoryByID_BadJSON(t *testing.T) {
	userID := uuid.New()
	connector := makeConnector(userID)

	srv := tokenServer(func(mux *http.ServeMux) {
		mux.HandleFunc("/repositories/42", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("bad json"))
		})
	})
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, err := api.GetRepositoryByID(userID.String(), 42)
	assert.Error(t, err)
	mockSt.AssertExpectations(t)
}

func TestGetRepositoryByID_Success(t *testing.T) {
	userID := uuid.New()
	connector := makeConnector(userID)
	repo := shared_types.GithubRepository{ID: 42, Name: "my-repo"}

	srv := tokenServer(func(mux *http.ServeMux) {
		mux.HandleFunc("/repositories/42", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(repo)
		})
	})
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	got, err := api.GetRepositoryByID(userID.String(), 42)
	assert.NoError(t, err)
	assert.Equal(t, "my-repo", got.Name)
	mockSt.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetRepositoryBranches
// ---------------------------------------------------------------------------

func TestGetRepositoryBranches(t *testing.T) {
	userID := uuid.New().String()
	repoName := "test-user/test-repo"
	connector := makeConnectorAt(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	expectedBranches := []shared_types.GithubRepositoryBranch{
		{Name: "main", Commit: struct {
			Sha string `json:"sha"`
			URL string `json:"url"`
		}{Sha: "abc123", URL: "https://api.github.com/repos/test-user/test-repo/commits/abc123"}, Protected: true},
	}

	tests := []struct {
		name         string
		setup        func(*testutil.MockGithubConnectorStorage)
		server       func() *httptest.Server
		wantErr      bool
		errContains  string
		wantBranches []shared_types.GithubRepositoryBranch
	}{
		{
			name: "success",
			setup: func(m *testutil.MockGithubConnectorStorage) {
				m.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Once()
			},
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/app/installations/67890/access_tokens" {
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(map[string]string{"token": "tok"})
						return
					}
					if r.URL.Path == "/repos/test-user/test-repo/branches" {
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode(expectedBranches)
						return
					}
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			wantBranches: expectedBranches,
		},
		{
			name: "no connectors",
			setup: func(m *testutil.MockGithubConnectorStorage) {
				m.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{}, nil).Once()
			},
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(404) }))
			},
			wantBranches: []shared_types.GithubRepositoryBranch{},
		},
		{
			name: "storage error",
			setup: func(m *testutil.MockGithubConnectorStorage) {
				m.On("GetAllConnectors", userID).Return(nil, assert.AnError).Once()
			},
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(404) }))
			},
			wantErr: true,
		},
		{
			name: "token failure",
			setup: func(m *testutil.MockGithubConnectorStorage) {
				m.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Once()
			},
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(401) }))
			},
			wantErr:     true,
			errContains: "authentication failed",
		},
		{
			name: "github API error",
			setup: func(m *testutil.MockGithubConnectorStorage) {
				m.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Once()
			},
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/app/installations/67890/access_tokens" {
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(map[string]string{"token": "tok"})
						return
					}
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
				}))
			},
			wantErr:     true,
			errContains: "404",
		},
		{
			name: "invalid json response",
			setup: func(m *testutil.MockGithubConnectorStorage) {
				m.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Once()
			},
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/app/installations/67890/access_tokens" {
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(map[string]string{"token": "tok"})
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("invalid json"))
				}))
			},
			wantErr: true,
		},
		{
			name: "invalid PEM",
			setup: func(m *testutil.MockGithubConnectorStorage) {
				bad := connector
				bad.Pem = "invalid-pem-data"
				m.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{bad}, nil).Once()
			},
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(404) }))
			},
			wantErr:     true,
			errContains: "failed to generate app JWT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSt := testutil.NewMockGithubConnectorStorage()
			if tt.setup != nil {
				tt.setup(mockSt)
			}
			srv := tt.server()
			defer srv.Close()

			prev := APIBaseURL
			SetAPIBaseURL(srv.URL)
			defer SetAPIBaseURL(prev)

			api := newAPI(mockSt)
			branches, err := api.GetRepositoryBranches(userID, repoName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, branches)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBranches, branches)
			}
			mockSt.AssertExpectations(t)
		})
	}
}

func TestGetRepositoryBranchesRequestHeaders(t *testing.T) {
	userID := uuid.New().String()
	connector := makeConnectorAt(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	var capturedAuth, capturedAccept, capturedAgent, capturedMethod string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app/installations/67890/access_tokens" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"token": "test-access-token"})
			return
		}
		if r.URL.Path == "/repos/test-user/test-repo/branches" {
			capturedAuth = r.Header.Get("Authorization")
			capturedAccept = r.Header.Get("Accept")
			capturedAgent = r.Header.Get("User-Agent")
			capturedMethod = r.Method
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]shared_types.GithubRepositoryBranch{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Once()

	api := newAPI(mockSt)
	_, err := api.GetRepositoryBranches(userID, "test-user/test-repo")
	assert.NoError(t, err)
	assert.Equal(t, "GET", capturedMethod)
	assert.Equal(t, "token test-access-token", capturedAuth)
	assert.Equal(t, "application/vnd.github.v3+json", capturedAccept)
	assert.Equal(t, "nixopus", capturedAgent)
	mockSt.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetRepositoryFileContent
// ---------------------------------------------------------------------------

func TestGetRepositoryFileContent_NoConnectors(t *testing.T) {
	userID := uuid.New().String()
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{}, nil)

	api := newAPI(mockSt)
	_, err := api.GetRepositoryFileContent(userID, "owner/repo", "main", "file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no GitHub connectors found")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoryFileContent_GetConnectorsError(t *testing.T) {
	userID := uuid.New().String()
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return(nil, assert.AnError)

	api := newAPI(mockSt)
	_, err := api.GetRepositoryFileContent(userID, "owner/repo", "main", "file.txt")
	assert.Error(t, err)
	mockSt.AssertExpectations(t)
}

func TestGetRepositoryFileContent_InvalidPEM(t *testing.T) {
	userID := uuid.New()
	connector := makeConnector(userID)
	connector.Pem = "bad"
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, err := api.GetRepositoryFileContent(userID.String(), "owner/repo", "main", "file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate GitHub App JWT")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoryFileContent_APIError(t *testing.T) {
	userID := uuid.New()
	connector := makeConnector(userID)

	srv := tokenServer(func(mux *http.ServeMux) {
		mux.HandleFunc("/repos/owner/repo/contents/file.txt", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
	})
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, err := api.GetRepositoryFileContent(userID.String(), "owner/repo", "main", "file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub API error")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoryFileContent_UnexpectedEncoding(t *testing.T) {
	userID := uuid.New()
	connector := makeConnector(userID)

	srv := tokenServer(func(mux *http.ServeMux) {
		mux.HandleFunc("/repos/owner/repo/contents/file.txt", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"content": "data", "encoding": "utf-8"})
		})
	})
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, err := api.GetRepositoryFileContent(userID.String(), "owner/repo", "main", "file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected content encoding")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoryFileContent_Success(t *testing.T) {
	userID := uuid.New()
	connector := makeConnector(userID)
	b64 := "aGVsbG8gd29ybGQ=" // "hello world"

	srv := tokenServer(func(mux *http.ServeMux) {
		mux.HandleFunc("/repos/owner/repo/contents/file.txt", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"content": b64, "encoding": "base64"})
		})
	})
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	data, err := api.GetRepositoryFileContent(userID.String(), "owner/repo", "main", "file.txt")
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world"), data)
	mockSt.AssertExpectations(t)
}

func TestGetRepositoryFileContent_NumericRepoID(t *testing.T) {
	userID := uuid.New()
	connector := makeConnector(userID)
	repo := shared_types.GithubRepository{ID: 99, FullName: "owner/numeric-repo"}
	b64 := "aGVsbG8=" // "hello"

	srv := tokenServer(func(mux *http.ServeMux) {
		mux.HandleFunc("/repositories/99", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(repo)
		})
		mux.HandleFunc("/repos/owner/numeric-repo/contents/readme.md", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"content": b64, "encoding": "base64"})
		})
	})
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil).Times(2)

	api := newAPI(mockSt)
	data, err := api.GetRepositoryFileContent(userID.String(), "99", "main", "readme.md")
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), data)
	mockSt.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// LatestCommitHash
// ---------------------------------------------------------------------------

func TestLatestCommitHash_InvalidURL(t *testing.T) {
	_, err := LatestCommitHash(logger.NewLogger(), "notaurl", "tok")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repository URL format")
}

func TestLatestCommitHash_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"Forbidden"}`)
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	_, err := LatestCommitHash(logger.NewLogger(), "https://github.com/owner/repo.git", "tok")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestLatestCommitHash_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	_, err := LatestCommitHash(logger.NewLogger(), "https://github.com/owner/repo", "tok")
	assert.Error(t, err)
}

func TestLatestCommitHash_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"sha":"abc123"}`)
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	sha, err := LatestCommitHash(logger.NewLogger(), "https://github.com/owner/repo", "tok")
	assert.NoError(t, err)
	assert.Equal(t, "abc123", sha)
}
