package gh

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/testutil"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestGetRepositoriesPaginated_NoConnectors(t *testing.T) {
	userID := uuid.New().String()
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{}, nil)

	api := newAPI(mockSt)
	repos, total, err := api.GetRepositoriesPaginated(userID, 1, 10, "", "", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, repos)
	mockSt.AssertExpectations(t)
}

func TestGetRepositoriesPaginated_GetConnectorsError(t *testing.T) {
	userID := uuid.New().String()
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return(nil, assert.AnError)

	api := newAPI(mockSt)
	_, _, err := api.GetRepositoriesPaginated(userID, 1, 10, "", "", "", "")
	assert.Error(t, err)
	mockSt.AssertExpectations(t)
}

func TestGetRepositoriesPaginated_ConnectorIDNotFound(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, _, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, "non-existent-id", "", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connector not found")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoriesPaginated_NoValidInstallation(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)
	connector.InstallationID = ""
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, _, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, "", "", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no connector with valid installation found")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoriesPaginated_InvalidPEM(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)
	connector.Pem = "invalid"
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, _, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, "", "", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate app JWT")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoriesPaginated_InstallationTokenFailure(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, _, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, "", "", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoriesPaginated_InstallationTokenFailure_NotFound(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, _, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, "", "", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GitHub installation")
	mockSt.AssertExpectations(t)
}

func TestGetRepositoriesPaginated_PaginatedSuccess(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)
	desc := "a repo"
	repos := []shared_types.GithubRepository{{ID: 1, Name: "repo1", Description: &desc}}

	srv := reposServer(10, 1, repos)
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	got, total, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, "", "", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, got, 1)
	mockSt.AssertExpectations(t)
}

func TestGetRepositoriesPaginated_WithSearch(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)
	desc := "awesome library"
	repos := []shared_types.GithubRepository{
		{ID: 1, Name: "awesome-lib", Description: &desc},
		{ID: 2, Name: "other-repo"},
	}

	srv := reposServer(10, 2, repos)
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	got, total, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, "", "awesome", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, got, 1)
	assert.Equal(t, "awesome-lib", got[0].Name)
	mockSt.AssertExpectations(t)
}

func TestGetRepositoriesPaginated_WithSortBy(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)
	repos := []shared_types.GithubRepository{
		{ID: 2, Name: "bravo", StargazersCount: 5},
		{ID: 1, Name: "alpha", StargazersCount: 10},
	}

	srv := reposServer(10, 2, repos)
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	got, _, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, "", "", "name", "asc")
	assert.NoError(t, err)
	assert.Equal(t, "alpha", got[0].Name)
	mockSt.AssertExpectations(t)
}

func TestGetRepositoriesPaginated_WithConnectorID(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)
	repos := []shared_types.GithubRepository{{ID: 1, Name: "repo1"}}

	srv := reposServer(10, 1, repos)
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	got, total, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, connector.ID.String(), "", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, got, 1)
	mockSt.AssertExpectations(t)
}

func TestFetchAllAndFilter_StartBeyondTotal(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)
	repos := []shared_types.GithubRepository{{ID: 1, Name: "only-repo"}}

	srv := reposServer(10, 1, repos)
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	got, total, err := api.GetRepositoriesPaginated(userID.String(), 100, 10, "", "only", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Empty(t, got)
	mockSt.AssertExpectations(t)
}

func TestFetchPaginatedRepositories_APIError(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app/installations/67890/access_tokens" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"token": "tok"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, _, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, "", "", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub API error")
	mockSt.AssertExpectations(t)
}

func TestFetchPaginatedRepositories_BadJSON(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app/installations/67890/access_tokens" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"token": "tok"})
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("bad json"))
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, _, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, "", "", "", "")
	assert.Error(t, err)
	mockSt.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// sortRepositories
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// fetchAllAndFilter — additional coverage
// ---------------------------------------------------------------------------

func TestFetchAllAndFilter_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	_, _, err := api.fetchAllAndFilter("tok", 1, 10, "search", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub API error")
}

func TestFetchAllAndFilter_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json {"))
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	_, _, err := api.fetchAllAndFilter("tok", 1, 10, "search", "", "")
	assert.Error(t, err)
}

// TestFetchAllAndFilter_MultiPage triggers the currentPage++ path by returning a full
// page (100 repos) with total_count > 100, forcing a second request.
func TestFetchAllAndFilter_MultiPage(t *testing.T) {
	repos := make([]shared_types.GithubRepository, 100)
	for i := range repos {
		repos[i] = shared_types.GithubRepository{ID: uint64(i + 1), Name: "repo"}
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count":  101,
			"repositories": repos,
		})
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	got, total, err := api.fetchAllAndFilter("tok", 1, 10, "repo", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 200, total) // two pages × 100 repos, all match "repo"
	assert.Len(t, got, 10)
}

// TestFetchAllAndFilter_DescriptionMatch covers the description-based filter branch.
func TestFetchAllAndFilter_DescriptionMatch(t *testing.T) {
	desc := "find-me"
	repos := []shared_types.GithubRepository{
		{ID: 1, Name: "no-match-name", Description: &desc},
		{ID: 2, Name: "also-no-match"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count":  2,
			"repositories": repos,
		})
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	got, total, err := api.fetchAllAndFilter("tok", 1, 10, "find-me", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, got, 1)
	assert.Equal(t, "no-match-name", got[0].Name)
}

// TestFetchAllAndFilter_RequestCreationError covers the http.NewRequest error path
// by using a URL with a control byte, which url.Parse rejects.
func TestFetchAllAndFilter_RequestCreationError(t *testing.T) {
	prev := APIBaseURL
	SetAPIBaseURL("\x00://invalid")
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	_, _, err := api.fetchAllAndFilter("tok", 1, 10, "search", "", "")
	assert.Error(t, err)
}

// TestFetchAllAndFilter_ClientDoError covers the client.Do network-error path.
func TestFetchAllAndFilter_ClientDoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close immediately so all connections are refused

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	_, _, err := api.fetchAllAndFilter("tok", 1, 10, "search", "", "")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// fetchAllSortAndPaginate — additional coverage
// ---------------------------------------------------------------------------

func TestFetchAllSortAndPaginate_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	_, _, err := api.fetchAllSortAndPaginate("tok", 1, 10, "name", "asc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub API error")
}

func TestFetchAllSortAndPaginate_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json {"))
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	_, _, err := api.fetchAllSortAndPaginate("tok", 1, 10, "name", "asc")
	assert.Error(t, err)
}

// TestFetchAllSortAndPaginate_MultiPage triggers the currentPage++ path.
func TestFetchAllSortAndPaginate_MultiPage(t *testing.T) {
	repos := make([]shared_types.GithubRepository, 100)
	for i := range repos {
		repos[i] = shared_types.GithubRepository{ID: uint64(i + 1), Name: "repo"}
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count":  101,
			"repositories": repos,
		})
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	got, total, err := api.fetchAllSortAndPaginate("tok", 1, 10, "name", "asc")
	assert.NoError(t, err)
	assert.Equal(t, 200, total)
	assert.Len(t, got, 10)
}

// TestFetchAllSortAndPaginate_StartBeyondTotal covers the start>totalCount clamping.
func TestFetchAllSortAndPaginate_StartBeyondTotal(t *testing.T) {
	repos := []shared_types.GithubRepository{{ID: 1, Name: "alpha"}, {ID: 2, Name: "beta"}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count":  2,
			"repositories": repos,
		})
	}))
	defer srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	got, total, err := api.fetchAllSortAndPaginate("tok", 100, 10, "name", "asc")
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Empty(t, got)
}

func TestFetchAllSortAndPaginate_RequestCreationError(t *testing.T) {
	prev := APIBaseURL
	SetAPIBaseURL("\x00://invalid")
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	_, _, err := api.fetchAllSortAndPaginate("tok", 1, 10, "name", "asc")
	assert.Error(t, err)
}

func TestFetchAllSortAndPaginate_ClientDoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	_, _, err := api.fetchAllSortAndPaginate("tok", 1, 10, "name", "asc")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// fetchPaginatedRepositories — additional coverage
// ---------------------------------------------------------------------------

func TestFetchPaginatedRepositories_RequestCreationError(t *testing.T) {
	prev := APIBaseURL
	SetAPIBaseURL("\x00://invalid")
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	_, _, err := api.fetchPaginatedRepositories("tok", 1, 10)
	assert.Error(t, err)
}

func TestFetchPaginatedRepositories_ClientDoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	prev := APIBaseURL
	SetAPIBaseURL(srv.URL)
	defer SetAPIBaseURL(prev)

	api := &API{Logger: testutil.NewLogger()}
	_, _, err := api.fetchPaginatedRepositories("tok", 1, 10)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetRepositoriesPaginated — connector-found-but-empty-installation edge case
// ---------------------------------------------------------------------------

// TestGetRepositoriesPaginated_ConnectorIDFoundButEmptyInstallation covers the path
// where a connector is matched by specific ID but has an empty InstallationID.
func TestGetRepositoriesPaginated_ConnectorIDFoundButEmptyInstallation(t *testing.T) {
	userID := uuid.New()
	connector := connectorWithInstall(userID)
	connector.InstallationID = "" // matched by ID but has no installation

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil)

	api := newAPI(mockSt)
	_, _, err := api.GetRepositoriesPaginated(userID.String(), 1, 10, connector.ID.String(), "", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connector has no installation_id")
	mockSt.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// sortRepositories
// ---------------------------------------------------------------------------

func TestSortRepositories_Empty(t *testing.T) {
	a := &API{}
	assert.Empty(t, a.sortRepositories(nil, "name", "asc"))
}

func TestSortRepositories_ByName_Asc(t *testing.T) {
	a := &API{}
	repos := []shared_types.GithubRepository{{Name: "zebra"}, {Name: "apple"}}
	sorted := a.sortRepositories(repos, "name", "asc")
	assert.Equal(t, "apple", sorted[0].Name)
}

func TestSortRepositories_ByName_Desc(t *testing.T) {
	a := &API{}
	repos := []shared_types.GithubRepository{{Name: "apple"}, {Name: "zebra"}}
	sorted := a.sortRepositories(repos, "name", "desc")
	assert.Equal(t, "zebra", sorted[0].Name)
}

func TestSortRepositories_ByStars_Asc(t *testing.T) {
	a := &API{}
	repos := []shared_types.GithubRepository{{Name: "a", StargazersCount: 100}, {Name: "b", StargazersCount: 1}}
	sorted := a.sortRepositories(repos, "stars", "asc")
	assert.Equal(t, uint64(1), sorted[0].StargazersCount)
}

func TestSortRepositories_ByStargazers_Desc(t *testing.T) {
	a := &API{}
	repos := []shared_types.GithubRepository{{Name: "a", StargazersCount: 1}, {Name: "b", StargazersCount: 100}}
	sorted := a.sortRepositories(repos, "stargazers_count", "desc")
	assert.Equal(t, uint64(100), sorted[0].StargazersCount)
}

func TestSortRepositories_DefaultSort(t *testing.T) {
	a := &API{}
	repos := []shared_types.GithubRepository{{Name: "z"}, {Name: "a"}}
	sorted := a.sortRepositories(repos, "unknown_field", "")
	assert.Equal(t, "a", sorted[0].Name)
}
