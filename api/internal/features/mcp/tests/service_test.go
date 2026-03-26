package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/service"
	"github.com/nixopus/nixopus/api/internal/features/mcp/storage"
	"github.com/nixopus/nixopus/api/internal/features/mcp/validation"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ─── Mock repository ──────────────────────────────────────────────────────────

type mockMCPRepo struct {
	mock.Mock
}

func (m *mockMCPRepo) CreateServer(s *shared_types.MCPServer) error {
	return m.Called(s).Error(0)
}

func (m *mockMCPRepo) ListServers(orgID uuid.UUID, p storage.ListServersParams) ([]shared_types.MCPServer, int, error) {
	args := m.Called(orgID, p)
	list, _ := args.Get(0).([]shared_types.MCPServer)
	return list, args.Int(1), args.Error(2)
}

func (m *mockMCPRepo) GetServerByID(id uuid.UUID) (*shared_types.MCPServer, error) {
	args := m.Called(id)
	s, _ := args.Get(0).(*shared_types.MCPServer)
	return s, args.Error(1)
}

func (m *mockMCPRepo) GetServerByName(orgID uuid.UUID, name string) (*shared_types.MCPServer, error) {
	args := m.Called(orgID, name)
	s, _ := args.Get(0).(*shared_types.MCPServer)
	return s, args.Error(1)
}

func (m *mockMCPRepo) UpdateServer(s *shared_types.MCPServer) error {
	return m.Called(s).Error(0)
}

func (m *mockMCPRepo) DeleteServer(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func newTestService(repo storage.MCPRepository) *service.MCPService {
	return service.NewMCPService(nil, context.Background(), logger.NewLogger(), repo)
}

// ─── AddServer ───────────────────────────────────────────────────────────────

func TestAddServer(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	req := &validation.CreateServerRequest{
		ProviderID:  "supabase",
		Name:        "My Supabase",
		Credentials: map[string]string{"access_token": "tok"},
		Enabled:     true,
	}

	t.Run("creates server when name is unique", func(t *testing.T) {
		repo := &mockMCPRepo{}
		repo.On("GetServerByName", orgID, req.Name).Return((*shared_types.MCPServer)(nil), nil)
		repo.On("CreateServer", mock.AnythingOfType("*types.MCPServer")).Return(nil)

		svc := newTestService(repo)
		server, err := svc.AddServer(req, orgID, userID)

		require.NoError(t, err)
		require.NotNil(t, server)
		assert.Equal(t, req.Name, server.Name)
		assert.Equal(t, req.ProviderID, server.ProviderID)
		assert.Equal(t, orgID, server.OrgID)
		assert.Equal(t, userID, server.CreatedBy)
		assert.WithinDuration(t, time.Now(), server.CreatedAt, 5*time.Second)
		repo.AssertExpectations(t)
	})

	t.Run("returns ErrDuplicateName when name already exists", func(t *testing.T) {
		existing := &shared_types.MCPServer{ID: uuid.New(), Name: req.Name}
		repo := &mockMCPRepo{}
		repo.On("GetServerByName", orgID, req.Name).Return(existing, nil)

		svc := newTestService(repo)
		_, err := svc.AddServer(req, orgID, userID)

		require.ErrorIs(t, err, service.ErrDuplicateName)
		repo.AssertNotCalled(t, "CreateServer")
		repo.AssertExpectations(t)
	})

	t.Run("propagates storage error from GetServerByName", func(t *testing.T) {
		repo := &mockMCPRepo{}
		repo.On("GetServerByName", orgID, req.Name).Return((*shared_types.MCPServer)(nil), assert.AnError)

		svc := newTestService(repo)
		_, err := svc.AddServer(req, orgID, userID)

		require.Error(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("sets custom_url only when provided", func(t *testing.T) {
		customReq := &validation.CreateServerRequest{
			ProviderID:  "custom",
			Name:        "My Custom",
			Credentials: map[string]string{},
			CustomURL:   "https://mcp.example.com/sse",
			Enabled:     true,
		}
		repo := &mockMCPRepo{}
		repo.On("GetServerByName", orgID, customReq.Name).Return((*shared_types.MCPServer)(nil), nil)
		repo.On("CreateServer", mock.AnythingOfType("*types.MCPServer")).Return(nil)

		svc := newTestService(repo)
		server, err := svc.AddServer(customReq, orgID, userID)

		require.NoError(t, err)
		require.NotNil(t, server.CustomURL)
		assert.Equal(t, "https://mcp.example.com/sse", *server.CustomURL)
		repo.AssertExpectations(t)
	})
}

// ─── DeleteServer ─────────────────────────────────────────────────────────────

func TestDeleteServer(t *testing.T) {
	orgID := uuid.New()
	serverID := uuid.New()

	t.Run("deletes server that belongs to org", func(t *testing.T) {
		server := &shared_types.MCPServer{ID: serverID, OrgID: orgID}
		repo := &mockMCPRepo{}
		repo.On("GetServerByID", serverID).Return(server, nil)
		repo.On("DeleteServer", serverID).Return(nil)

		svc := newTestService(repo)
		err := svc.DeleteServer(serverID.String(), orgID)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("returns ErrServerNotFound when server doesn't exist", func(t *testing.T) {
		repo := &mockMCPRepo{}
		repo.On("GetServerByID", serverID).Return((*shared_types.MCPServer)(nil), nil)

		svc := newTestService(repo)
		err := svc.DeleteServer(serverID.String(), orgID)

		require.ErrorIs(t, err, service.ErrServerNotFound)
		repo.AssertNotCalled(t, "DeleteServer")
	})

	t.Run("returns ErrServerNotFound when server belongs to different org", func(t *testing.T) {
		otherOrgServer := &shared_types.MCPServer{ID: serverID, OrgID: uuid.New()}
		repo := &mockMCPRepo{}
		repo.On("GetServerByID", serverID).Return(otherOrgServer, nil)

		svc := newTestService(repo)
		err := svc.DeleteServer(serverID.String(), orgID)

		require.ErrorIs(t, err, service.ErrServerNotFound)
		repo.AssertNotCalled(t, "DeleteServer")
	})

	t.Run("returns error on invalid UUID", func(t *testing.T) {
		repo := &mockMCPRepo{}
		svc := newTestService(repo)
		err := svc.DeleteServer("not-a-valid-uuid", orgID)
		require.Error(t, err)
	})
}

// ─── ListServers ─────────────────────────────────────────────────────────────

func TestListServers(t *testing.T) {
	orgID := uuid.New()

	t.Run("returns servers from storage", func(t *testing.T) {
		servers := []shared_types.MCPServer{
			{ID: uuid.New(), Name: "Alpha", OrgID: orgID},
			{ID: uuid.New(), Name: "Beta", OrgID: orgID},
		}
		params := storage.ListServersParams{Page: 1, Limit: 10}
		repo := &mockMCPRepo{}
		repo.On("ListServers", orgID, params).Return(servers, 2, nil)

		svc := newTestService(repo)
		got, total, err := svc.ListServers(orgID, params)

		require.NoError(t, err)
		assert.Len(t, got, 2)
		assert.Equal(t, 2, total)
		repo.AssertExpectations(t)
	})

	t.Run("forwards search and sort params to storage unchanged", func(t *testing.T) {
		params := storage.ListServersParams{Q: "sup", SortBy: "name", SortDir: "desc", Page: 2, Limit: 5}
		repo := &mockMCPRepo{}
		repo.On("ListServers", orgID, params).Return([]shared_types.MCPServer{}, 0, nil)

		svc := newTestService(repo)
		_, _, err := svc.ListServers(orgID, params)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
