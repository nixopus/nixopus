package tests

import (
	"errors"
	"testing"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGithubConnectorStorage implements GithubConnectorRepository for successful operations
type MockGithubConnectorStorage struct {
	mock.Mock
	connectors     map[string]*shared_types.GithubConnector
	appMapping     map[string]string
	userConnectors map[string][]string
	lastConnector  *shared_types.GithubConnector
	methodCalls    map[string]int
}

// NewMockGithubConnectorStorage creates a new instance of MockGithubConnectorStorage
func NewMockGithubConnectorStorage() *MockGithubConnectorStorage {
	return &MockGithubConnectorStorage{
		connectors:     make(map[string]*shared_types.GithubConnector),
		appMapping:     make(map[string]string),
		userConnectors: make(map[string][]string),
		methodCalls:    make(map[string]int),
	}
}

func (m *MockGithubConnectorStorage) CreateConnector(connector *shared_types.GithubConnector) error {
	m.methodCalls["CreateConnector"] = m.methodCalls["CreateConnector"] + 1
	m.lastConnector = connector

	args := m.Called(connector)
	if args.Get(0) != nil {
		return args.Error(0)
	}

	connectorID := connector.ID.String()
	userID := connector.UserID.String()

	m.connectors[connectorID] = connector
	m.appMapping[connector.AppID] = connectorID

	if _, exists := m.userConnectors[userID]; !exists {
		m.userConnectors[userID] = []string{}
	}
	m.userConnectors[userID] = append(m.userConnectors[userID], connectorID)

	return nil
}

func (m *MockGithubConnectorStorage) UpdateConnector(ConnectorID, InstallationID string) error {
	m.methodCalls["UpdateConnector"] = m.methodCalls["UpdateConnector"] + 1

	args := m.Called(ConnectorID, InstallationID)
	if args.Get(0) != nil {
		return args.Error(0)
	}

	if connector, exists := m.connectors[ConnectorID]; exists {
		connector.InstallationID = InstallationID
		return nil
	}
	return errors.New("connector not found")
}

func (m *MockGithubConnectorStorage) GetConnector(ConnectorID string) (*shared_types.GithubConnector, error) {
	m.methodCalls["GetConnector"] = m.methodCalls["GetConnector"] + 1

	args := m.Called(ConnectorID)
	if args.Get(0) != nil {
		return args.Get(0).(*shared_types.GithubConnector), args.Error(1)
	}

	if connector, exists := m.connectors[ConnectorID]; exists {
		return connector, nil
	}
	return nil, errors.New("connector not found")
}

func (m *MockGithubConnectorStorage) GetAllConnectors(UserID string) ([]shared_types.GithubConnector, error) {
	m.methodCalls["GetAllConnectors"] = m.methodCalls["GetAllConnectors"] + 1

	args := m.Called(UserID)
	if len(args) > 0 && args.Get(0) != nil {
		return args.Get(0).([]shared_types.GithubConnector), args.Error(1)
	}

	var result []shared_types.GithubConnector

	if connectorIDs, exists := m.userConnectors[UserID]; exists {
		for _, id := range connectorIDs {
			if connector, found := m.connectors[id]; found {
				result = append(result, *connector)
			}
		}
	}

	return result, nil
}

func (m *MockGithubConnectorStorage) GetConnectorByAppID(AppID string) (*shared_types.GithubConnector, error) {
	m.methodCalls["GetConnectorByAppID"] = m.methodCalls["GetConnectorByAppID"] + 1

	args := m.Called(AppID)
	if args.Get(0) != nil {
		return args.Get(0).(*shared_types.GithubConnector), args.Error(1)
	}

	if connectorID, exists := m.appMapping[AppID]; exists {
		return m.connectors[connectorID], nil
	}
	return nil, errors.New("connector not found")
}

// MockGithubConnectorStorageWithErr implements GithubConnectorRepository with error responses
type MockGithubConnectorStorageWithErr struct {
	mock.Mock
}

// NewMockGithubConnectorStorageWithErr creates a new instance of MockGithubConnectorStorageWithErr
func NewMockGithubConnectorStorageWithErr() *MockGithubConnectorStorageWithErr {
	return &MockGithubConnectorStorageWithErr{}
}

func (m *MockGithubConnectorStorageWithErr) CreateConnector(connector *shared_types.GithubConnector) error {
	args := m.Called(connector)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return errors.New("failed to create connector")
}

func (m *MockGithubConnectorStorageWithErr) UpdateConnector(ConnectorID, InstallationID string) error {
	args := m.Called(ConnectorID, InstallationID)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return errors.New("failed to update connector")
}

func (m *MockGithubConnectorStorageWithErr) GetConnector(ConnectorID string) (*shared_types.GithubConnector, error) {
	args := m.Called(ConnectorID)
	if args.Get(0) != nil {
		return args.Get(0).(*shared_types.GithubConnector), args.Error(1)
	}
	return nil, errors.New("failed to get connector")
}

func (m *MockGithubConnectorStorageWithErr) GetAllConnectors(UserID string) ([]shared_types.GithubConnector, error) {
	args := m.Called(UserID)
	if args.Get(0) != nil {
		return args.Get(0).([]shared_types.GithubConnector), args.Error(1)
	}
	return nil, errors.New("failed to get all connectors")
}

func (m *MockGithubConnectorStorageWithErr) GetConnectorByAppID(AppID string) (*shared_types.GithubConnector, error) {
	args := m.Called(AppID)
	if args.Get(0) != nil {
		return args.Get(0).(*shared_types.GithubConnector), args.Error(1)
	}
	return nil, errors.New("failed to get connector by app ID")
}

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

func (m *CustomMockStorage) ExpectUpdateConnector(connectorID, installationID string, err error) {
	m.updateConnectorID = connectorID
	m.updateConnectorInstallID = installationID
	m.updateConnectorError = err
}

func (m *CustomMockStorage) GetAllConnectors(userID string) ([]shared_types.GithubConnector, error) {
	m.getAllConnectorsCalled = true
	if userID == m.getAllConnectorsUserID {
		return m.getAllConnectorsResult, m.getAllConnectorsError
	}
	return nil, errors.New("unexpected userID")
}

func (m *CustomMockStorage) UpdateConnector(connectorID, installationID string) error {
	m.updateConnectorCalled = true
	if connectorID == m.updateConnectorID && installationID == m.updateConnectorInstallID {
		return m.updateConnectorError
	}
	return errors.New("unexpected connector or installation ID")
}

func (m *CustomMockStorage) CreateConnector(connector *shared_types.GithubConnector) error {
	return errors.New("not implemented for this test")
}

func (m *CustomMockStorage) GetConnector(connectorID string) (*shared_types.GithubConnector, error) {
	return nil, errors.New("not implemented for this test")
}

func (m *CustomMockStorage) GetConnectorByAppID(appID string) (*shared_types.GithubConnector, error) {
	return nil, errors.New("not implemented for this test")
}

func (m *CustomMockStorage) VerifyExpectations(t *testing.T) {
	assert.True(t, m.getAllConnectorsCalled, "GetAllConnectors was not called")
	assert.True(t, m.updateConnectorCalled, "UpdateConnector was not called")
}
