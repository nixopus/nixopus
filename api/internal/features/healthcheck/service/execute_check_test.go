package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// getAppProvider
// ---------------------------------------------------------------------------

func TestGetAppProvider_ReturnsInjectedProvider(t *testing.T) {
	mock := &mockAppProvider{}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, mock)
	assert.Equal(t, mock, svc.getAppProvider())
}

func TestGetAppProvider_CreatesFromStoreWhenNil(t *testing.T) {
	svc := newTestService(&mockHealthCheckRepo{})
	svc.store = &shared_storage.Store{}
	p := svc.getAppProvider()
	assert.NotNil(t, p)
}

// ---------------------------------------------------------------------------
// ProcessHealthCheckResult
// ---------------------------------------------------------------------------

func TestProcessHealthCheckResult_AddResultError(t *testing.T) {
	addErr := errors.New("insert failed")
	repo := &mockHealthCheckRepo{
		addHealthCheckResult: func(r *shared_types.HealthCheckResult) error {
			return addErr
		},
	}
	svc := newTestService(repo)
	err := svc.ProcessHealthCheckResult(
		&shared_types.HealthCheck{},
		&shared_types.HealthCheckResult{Status: "healthy"},
	)
	assert.ErrorIs(t, err, addErr)
}

func TestProcessHealthCheckResult_UpdateStatusError(t *testing.T) {
	updateErr := errors.New("update failed")
	repo := &mockHealthCheckRepo{
		addHealthCheckResult: func(r *shared_types.HealthCheckResult) error { return nil },
		updateHealthCheckStatus: func(id uuid.UUID, fails int, at time.Time) error {
			return updateErr
		},
	}
	svc := newTestService(repo)
	err := svc.ProcessHealthCheckResult(
		&shared_types.HealthCheck{},
		&shared_types.HealthCheckResult{Status: "unhealthy"},
	)
	assert.ErrorIs(t, err, updateErr)
}

func TestProcessHealthCheckResult_HealthyResetsFailsAboveThreshold(t *testing.T) {
	var capturedFails int
	repo := &mockHealthCheckRepo{
		addHealthCheckResult: func(r *shared_types.HealthCheckResult) error { return nil },
		updateHealthCheckStatus: func(id uuid.UUID, fails int, at time.Time) error {
			capturedFails = fails
			return nil
		},
	}
	svc := newTestService(repo)
	hc := &shared_types.HealthCheck{ConsecutiveFails: 5, SuccessThreshold: 3}
	err := svc.ProcessHealthCheckResult(hc, &shared_types.HealthCheckResult{
		Status: string(shared_types.HealthCheckStatusHealthy),
	})
	require.NoError(t, err)
	assert.Equal(t, 0, capturedFails)
}

func TestProcessHealthCheckResult_HealthyResetsFailsBelowThreshold(t *testing.T) {
	var capturedFails int
	repo := &mockHealthCheckRepo{
		addHealthCheckResult: func(r *shared_types.HealthCheckResult) error { return nil },
		updateHealthCheckStatus: func(id uuid.UUID, fails int, at time.Time) error {
			capturedFails = fails
			return nil
		},
	}
	svc := newTestService(repo)
	hc := &shared_types.HealthCheck{ConsecutiveFails: 1, SuccessThreshold: 3}
	err := svc.ProcessHealthCheckResult(hc, &shared_types.HealthCheckResult{
		Status: string(shared_types.HealthCheckStatusHealthy),
	})
	require.NoError(t, err)
	assert.Equal(t, 0, capturedFails)
}

func TestProcessHealthCheckResult_UnhealthyIncrementsCounter(t *testing.T) {
	var capturedFails int
	repo := &mockHealthCheckRepo{
		addHealthCheckResult: func(r *shared_types.HealthCheckResult) error { return nil },
		updateHealthCheckStatus: func(id uuid.UUID, fails int, at time.Time) error {
			capturedFails = fails
			return nil
		},
	}
	svc := newTestService(repo)
	hc := &shared_types.HealthCheck{ConsecutiveFails: 2}
	err := svc.ProcessHealthCheckResult(hc, &shared_types.HealthCheckResult{
		Status: string(shared_types.HealthCheckStatusUnhealthy),
	})
	require.NoError(t, err)
	assert.Equal(t, 3, capturedFails)
}

// ---------------------------------------------------------------------------
// ExecuteHealthCheck — application lookup failure
// ---------------------------------------------------------------------------

func TestExecuteHealthCheck_ApplicationNotFound(t *testing.T) {
	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{}, errors.New("application not found")
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:            uuid.New(),
		ApplicationID: uuid.New(),
	}
	result, err := svc.ExecuteHealthCheck(hc)
	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get application")
}

// ---------------------------------------------------------------------------
// ExecuteHealthCheck — buildHealthCheckURL errors
// ---------------------------------------------------------------------------

func TestExecuteHealthCheck_NoDomainsAndLoadFails(t *testing.T) {
	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil // no domains
		},
		getApplicationDomains: func(appID uuid.UUID) ([]shared_types.ApplicationDomain, error) {
			return nil, errors.New("domains fetch failed")
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:            uuid.New(),
		ApplicationID: uuid.New(),
		Endpoint:      "/health",
	}
	result, err := svc.ExecuteHealthCheck(hc)
	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load domains")
}

func TestExecuteHealthCheck_EmptyDomainsAfterLoad(t *testing.T) {
	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil
		},
		getApplicationDomains: func(appID uuid.UUID) ([]shared_types.ApplicationDomain, error) {
			return []shared_types.ApplicationDomain{}, nil // empty list
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:            uuid.New(),
		ApplicationID: uuid.New(),
		Endpoint:      "/health",
	}
	result, err := svc.ExecuteHealthCheck(hc)
	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no domains configured")
}

// ---------------------------------------------------------------------------
// ExecuteHealthCheck — HTTP request creation failure
// ---------------------------------------------------------------------------

func TestExecuteHealthCheck_InvalidRequestMethod(t *testing.T) {
	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	// A null byte in the method makes http.NewRequest fail
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       "https://example.com/health",
		Method:         "INVALID\x00METHOD",
		ExpectedStatus: []int{200},
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err) // function does NOT return the error; it wraps it in result
	require.NotNil(t, result)
	assert.Equal(t, string(shared_types.HealthCheckStatusError), result.Status)
	assert.NotEmpty(t, result.ErrorMessage)
}

func TestExecuteHealthCheck_InvalidRequestViaPostBody(t *testing.T) {
	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	// NULL byte in URL triggers url.Parse error for the POST+body branch
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       "https://\x00invalid.example.com/health",
		Method:         "POST",
		Body:           `{"ping":true}`,
		ExpectedStatus: []int{200},
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, string(shared_types.HealthCheckStatusError), result.Status)
}

// ---------------------------------------------------------------------------
// ExecuteHealthCheck — successful HTTP round-trips
// ---------------------------------------------------------------------------

func TestExecuteHealthCheck_HealthyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       server.URL,
		Method:         "GET",
		ExpectedStatus: []int{200},
		TimeoutSeconds: 5,
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, string(shared_types.HealthCheckStatusHealthy), result.Status)
	assert.Equal(t, 200, result.StatusCode)
}

func TestExecuteHealthCheck_UnhealthyStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       server.URL,
		Method:         "GET",
		ExpectedStatus: []int{200},
		TimeoutSeconds: 5,
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, string(shared_types.HealthCheckStatusUnhealthy), result.Status)
	assert.Contains(t, result.ErrorMessage, "503")
}

func TestExecuteHealthCheck_PostWithBody(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 512)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       server.URL,
		Method:         "POST",
		Body:           `{"ping":true}`,
		ExpectedStatus: []int{200},
		TimeoutSeconds: 5,
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	assert.Equal(t, string(shared_types.HealthCheckStatusHealthy), result.Status)
	assert.Contains(t, receivedBody, "ping")
}

func TestExecuteHealthCheck_CustomHeaders(t *testing.T) {
	var receivedHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       server.URL,
		Method:         "GET",
		Headers:        map[string]string{"X-Custom": "test-value"},
		ExpectedStatus: []int{200},
		TimeoutSeconds: 5,
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	assert.Equal(t, string(shared_types.HealthCheckStatusHealthy), result.Status)
	assert.Equal(t, "test-value", receivedHeader)
}

func TestExecuteHealthCheck_DefaultUserAgentSet(t *testing.T) {
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       server.URL,
		Method:         "GET",
		ExpectedStatus: []int{200},
		TimeoutSeconds: 5,
	}
	_, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	assert.Equal(t, "Nixopus-HealthCheck/1.0", receivedUA)
}

func TestExecuteHealthCheck_UserAgentNotOverriddenIfProvided(t *testing.T) {
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       server.URL,
		Method:         "GET",
		Headers:        map[string]string{"User-Agent": "MyAgent/2.0"},
		ExpectedStatus: []int{200},
		TimeoutSeconds: 5,
	}
	_, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	assert.Equal(t, "MyAgent/2.0", receivedUA)
}

func TestExecuteHealthCheck_ConnectionRefused(t *testing.T) {
	// Use a port that is not listening
	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       "http://127.0.0.1:19999/health",
		Method:         "GET",
		ExpectedStatus: []int{200},
		TimeoutSeconds: 2,
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err) // wrapped in result, not returned as error
	require.NotNil(t, result)
	assert.Equal(t, string(shared_types.HealthCheckStatusError), result.Status)
}

func TestExecuteHealthCheck_Timeout(t *testing.T) {
	// Server that hangs indefinitely
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: uuid.New()}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       server.URL,
		Method:         "GET",
		ExpectedStatus: []int{200},
		TimeoutSeconds: 1, // 1 second timeout
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, string(shared_types.HealthCheckStatusTimeout), result.Status)
}

// ---------------------------------------------------------------------------
// ExecuteHealthCheck — buildHealthCheckURL domain path
// ---------------------------------------------------------------------------

func TestExecuteHealthCheck_UsesPreloadedDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Extract host from server URL (e.g. "127.0.0.1:PORT")
	host := server.URL[len("http://"):]

	domainID := uuid.New()
	domain := shared_types.ApplicationDomain{ID: domainID, Domain: host}

	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{
				ID:      uuid.New(),
				Domains: []*shared_types.ApplicationDomain{&domain},
			}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       "/",
		Method:         "GET",
		ExpectedStatus: []int{200},
		TimeoutSeconds: 5,
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, string(shared_types.HealthCheckStatusHealthy), result.Status)
}

func TestExecuteHealthCheck_LoadsDomainsFromStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	host := server.URL[len("http://"):]
	appID := uuid.New()

	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{ID: appID}, nil // no pre-loaded domains
		},
		getApplicationDomains: func(id uuid.UUID) ([]shared_types.ApplicationDomain, error) {
			assert.Equal(t, appID, id)
			return []shared_types.ApplicationDomain{
				{ID: uuid.New(), Domain: host},
			}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       "/",
		Method:         "GET",
		ExpectedStatus: []int{200},
		TimeoutSeconds: 5,
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, string(shared_types.HealthCheckStatusHealthy), result.Status)
}

func TestExecuteHealthCheck_LocalhostUsesHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// server.URL is "http://127.0.0.1:PORT" — extract just the host+port
	host := server.URL[len("http://"):]

	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{
				ID: uuid.New(),
				Domains: []*shared_types.ApplicationDomain{
					{Domain: host},
				},
			}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       "/ping",
		Method:         "GET",
		ExpectedStatus: []int{200},
		TimeoutSeconds: 5,
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	require.NotNil(t, result)
	// 127.0.0.1 triggers http protocol — the httptest server handles it fine
	assert.Equal(t, string(shared_types.HealthCheckStatusHealthy), result.Status)
}

func TestExecuteHealthCheck_LocalhostDomainName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Replace 127.0.0.1 with localhost in the URL
	host := "localhost" + server.URL[len("http://127.0.0.1"):]

	appProvider := &mockAppProvider{
		getApplicationById: func(id string, orgID uuid.UUID) (shared_types.Application, error) {
			return shared_types.Application{
				ID: uuid.New(),
				Domains: []*shared_types.ApplicationDomain{
					{Domain: host},
				},
			}, nil
		},
	}
	svc := newTestServiceWithApp(&mockHealthCheckRepo{}, appProvider)
	hc := &shared_types.HealthCheck{
		ID:             uuid.New(),
		ApplicationID:  uuid.New(),
		Endpoint:       "/",
		Method:         "GET",
		ExpectedStatus: []int{200},
		TimeoutSeconds: 5,
	}
	result, err := svc.ExecuteHealthCheck(hc)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, string(shared_types.HealthCheckStatusHealthy), result.Status)
}
