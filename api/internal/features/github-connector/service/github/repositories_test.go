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
