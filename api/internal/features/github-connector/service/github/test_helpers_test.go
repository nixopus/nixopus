package gh

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/testutil"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// newAPI returns an *API wired to a fresh mock storage.
func newAPI(mock *testutil.MockGithubConnectorStorage) *API {
	return NewAPI(mock, testutil.NewLogger())
}

// makeConnector returns a ready-made connector for the given user, with a valid PEM key.
func makeConnector(userID uuid.UUID) shared_types.GithubConnector {
	return testutil.MakeTestConnector(userID)
}

// connectorWithInstall returns a connector whose InstallationID is "67890".
func connectorWithInstall(userID uuid.UUID) shared_types.GithubConnector {
	c := makeConnector(userID)
	c.InstallationID = "67890"
	return c
}

// tokenServer returns a test server that grants installation tokens for installation 67890.
// extra routes are added after the default token route.
func tokenServer(extraRoutes func(mux *http.ServeMux)) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/app/installations/67890/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"token": "test-access-token"})
	})
	if extraRoutes != nil {
		extraRoutes(mux)
	}
	return httptest.NewServer(mux)
}

// reposServer builds a test server that issues installation tokens and
// serves a fixed /installation/repositories payload.
func reposServer(perPage int, totalCount int, repos []shared_types.GithubRepository) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/app/installations/67890/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"token": "tok"})
	})
	mux.HandleFunc("/installation/repositories", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count":  totalCount,
			"repositories": repos,
		})
	})
	_ = perPage
	return httptest.NewServer(mux)
}

// makeConnectorWithTime mirrors the branches-test fixture with explicit timestamps.
func makeConnectorAt(userID uuid.UUID) shared_types.GithubConnector {
	c := makeConnector(userID)
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return c
}
