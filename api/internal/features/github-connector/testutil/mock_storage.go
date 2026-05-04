package testutil

import (
	"errors"
	"testing"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// MockGithubConnectorStorage — full-featured testify mock
// ---------------------------------------------------------------------------

// MockGithubConnectorStorage implements storage.GithubConnectorRepository.
type MockGithubConnectorStorage struct {
	mock.Mock
	connectors     map[string]*shared_types.GithubConnector
	appMapping     map[string]string
	userConnectors map[string][]string
	lastConnector  *shared_types.GithubConnector
	methodCalls    map[string]int
}

// NewMockGithubConnectorStorage creates a MockGithubConnectorStorage.
func NewMockGithubConnectorStorage() *MockGithubConnectorStorage {
	return &MockGithubConnectorStorage{
		connectors:     make(map[string]*shared_types.GithubConnector),
		appMapping:     make(map[string]string),
		userConnectors: make(map[string][]string),
		methodCalls:    make(map[string]int),
	}
}

// LastConnector returns the most-recently created connector (nil if none).
func (m *MockGithubConnectorStorage) LastConnector() *shared_types.GithubConnector {
	return m.lastConnector
}

// MethodCalls returns a copy of the call-count map.
func (m *MockGithubConnectorStorage) MethodCalls() map[string]int { return m.methodCalls }

func (m *MockGithubConnectorStorage) CreateConnector(connector *shared_types.GithubConnector) error {
	m.methodCalls["CreateConnector"]++
	m.lastConnector = connector

	args := m.Called(connector)
	if args.Get(0) != nil {
		return args.Error(0)
	}

	id := connector.ID.String()
	uid := connector.UserID.String()
	m.connectors[id] = connector
	m.appMapping[connector.AppID] = id
	if m.userConnectors[uid] == nil {
		m.userConnectors[uid] = []string{}
	}
	m.userConnectors[uid] = append(m.userConnectors[uid], id)
	return nil
}

func (m *MockGithubConnectorStorage) UpdateConnector(connectorID, installationID string) error {
	m.methodCalls["UpdateConnector"]++
	args := m.Called(connectorID, installationID)
	if len(args) > 0 {
		return args.Error(0)
	}
	if c, ok := m.connectors[connectorID]; ok {
		c.InstallationID = installationID
		return nil
	}
	return errors.New("connector not found")
}

func (m *MockGithubConnectorStorage) GetConnector(connectorID string) (*shared_types.GithubConnector, error) {
	m.methodCalls["GetConnector"]++
	args := m.Called(connectorID)
	if len(args) > 0 {
		if err := args.Error(1); err != nil {
			return nil, err
		}
		if v := args.Get(0); v != nil {
			return v.(*shared_types.GithubConnector), nil
		}
		return nil, nil
	}
	if c, ok := m.connectors[connectorID]; ok {
		return c, nil
	}
	return nil, errors.New("connector not found")
}

func (m *MockGithubConnectorStorage) GetAllConnectors(userID string) ([]shared_types.GithubConnector, error) {
	m.methodCalls["GetAllConnectors"]++
	args := m.Called(userID)
	if len(args) > 0 {
		if err := args.Error(1); err != nil {
			return nil, err
		}
		if v := args.Get(0); v != nil {
			return v.([]shared_types.GithubConnector), nil
		}
		return []shared_types.GithubConnector{}, nil
	}
	var result []shared_types.GithubConnector
	for _, id := range m.userConnectors[userID] {
		if c, ok := m.connectors[id]; ok {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *MockGithubConnectorStorage) GetConnectorByAppID(appID string) (*shared_types.GithubConnector, error) {
	m.methodCalls["GetConnectorByAppID"]++
	args := m.Called(appID)
	if len(args) > 0 {
		if err := args.Error(1); err != nil {
			return nil, err
		}
		if v := args.Get(0); v != nil {
			return v.(*shared_types.GithubConnector), nil
		}
		return nil, nil
	}
	if id, ok := m.appMapping[appID]; ok {
		return m.connectors[id], nil
	}
	return nil, errors.New("connector not found")
}

func (m *MockGithubConnectorStorage) DeleteConnector(connectorID, userID string) error {
	m.methodCalls["DeleteConnector"]++
	args := m.Called(connectorID, userID)
	if len(args) > 0 {
		return args.Error(0)
	}
	if c, ok := m.connectors[connectorID]; ok {
		if c.UserID.String() != userID {
			return errors.New("permission denied")
		}
		delete(m.connectors, connectorID)
		return nil
	}
	return errors.New("connector not found")
}

// ---------------------------------------------------------------------------
// MockGithubConnectorStorageWithErr — always errors unless stub overrides
// ---------------------------------------------------------------------------

// MockGithubConnectorStorageWithErr returns errors from all methods unless overridden.
type MockGithubConnectorStorageWithErr struct{ mock.Mock }

// NewMockGithubConnectorStorageWithErr creates a MockGithubConnectorStorageWithErr.
func NewMockGithubConnectorStorageWithErr() *MockGithubConnectorStorageWithErr {
	return &MockGithubConnectorStorageWithErr{}
}

func (m *MockGithubConnectorStorageWithErr) CreateConnector(c *shared_types.GithubConnector) error {
	args := m.Called(c)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return errors.New("failed to create connector")
}
func (m *MockGithubConnectorStorageWithErr) UpdateConnector(id, inst string) error {
	args := m.Called(id, inst)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return errors.New("failed to update connector")
}
func (m *MockGithubConnectorStorageWithErr) GetConnector(id string) (*shared_types.GithubConnector, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*shared_types.GithubConnector), args.Error(1)
	}
	return nil, errors.New("failed to get connector")
}
func (m *MockGithubConnectorStorageWithErr) GetAllConnectors(uid string) ([]shared_types.GithubConnector, error) {
	args := m.Called(uid)
	if args.Get(0) != nil {
		return args.Get(0).([]shared_types.GithubConnector), args.Error(1)
	}
	return nil, errors.New("failed to get all connectors")
}
func (m *MockGithubConnectorStorageWithErr) GetConnectorByAppID(appID string) (*shared_types.GithubConnector, error) {
	args := m.Called(appID)
	if args.Get(0) != nil {
		return args.Get(0).(*shared_types.GithubConnector), args.Error(1)
	}
	return nil, errors.New("failed to get connector by app ID")
}
func (m *MockGithubConnectorStorageWithErr) DeleteConnector(id, uid string) error {
	args := m.Called(id, uid)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return errors.New("failed to delete connector")
}

// ---------------------------------------------------------------------------
// CustomMockStorage — minimal hand-rolled mock for update-connector flows
// ---------------------------------------------------------------------------

// CustomMockStorage is a minimal hand-rolled mock for update-connector test flows.
type CustomMockStorage struct {
	getAllConnectorsUserID string
	getAllConnectorsResult []shared_types.GithubConnector
	getAllConnectorsError  error

	updateConnectorID        string
	updateConnectorInstallID string
	updateConnectorError     error

	getAllConnectorsCalled bool
	updateConnectorCalled  bool
}

func (m *CustomMockStorage) ExpectGetAllConnectors(userID string, result []shared_types.GithubConnector, err error) {
	m.getAllConnectorsUserID = userID
	m.getAllConnectorsResult = result
	m.getAllConnectorsError = err
}
func (m *CustomMockStorage) ExpectUpdateConnector(connectorID, installID string, err error) {
	m.updateConnectorID = connectorID
	m.updateConnectorInstallID = installID
	m.updateConnectorError = err
}
func (m *CustomMockStorage) GetAllConnectors(userID string) ([]shared_types.GithubConnector, error) {
	m.getAllConnectorsCalled = true
	if userID == m.getAllConnectorsUserID {
		return m.getAllConnectorsResult, m.getAllConnectorsError
	}
	return nil, errors.New("unexpected userID")
}
func (m *CustomMockStorage) UpdateConnector(connectorID, installID string) error {
	m.updateConnectorCalled = true
	if connectorID == m.updateConnectorID && installID == m.updateConnectorInstallID {
		return m.updateConnectorError
	}
	return errors.New("unexpected connector or installation ID")
}
func (m *CustomMockStorage) CreateConnector(_ *shared_types.GithubConnector) error {
	return errors.New("not implemented for this test")
}
func (m *CustomMockStorage) GetConnector(_ string) (*shared_types.GithubConnector, error) {
	return nil, errors.New("not implemented for this test")
}
func (m *CustomMockStorage) GetConnectorByAppID(_ string) (*shared_types.GithubConnector, error) {
	return nil, errors.New("not implemented for this test")
}
func (m *CustomMockStorage) DeleteConnector(_, _ string) error {
	return errors.New("not implemented for this test")
}
func (m *CustomMockStorage) VerifyExpectations(t *testing.T) {
	assert.True(t, m.getAllConnectorsCalled, "GetAllConnectors was not called")
	assert.True(t, m.updateConnectorCalled, "UpdateConnector was not called")
}
