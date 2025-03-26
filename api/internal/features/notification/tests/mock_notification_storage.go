package tests

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

type MockNotificationStorage struct {
	mock.Mock
}

func (m *MockNotificationStorage) AddSmtp(config *shared_types.SMTPConfigs) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockNotificationStorage) UpdateSmtp(config *notification.UpdateSMTPConfigRequest) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockNotificationStorage) DeleteSmtp(ID string) error {
	args := m.Called(ID)
	return args.Error(0)
}

func (m *MockNotificationStorage) GetPreferences(context context.Context, userID uuid.UUID) (*notification.GetPreferencesResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(*notification.GetPreferencesResponse), args.Error(1)
}

func (m *MockNotificationStorage) GetSmtp(ID string) (*shared_types.SMTPConfigs, error) {
	args := m.Called(ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared_types.SMTPConfigs), args.Error(1)
}

func (m *MockNotificationStorage) GetOrganizationsSmtp(organizationID string) ([]shared_types.SMTPConfigs, error) {
	args := m.Called(organizationID)
	return args.Get(0).([]shared_types.SMTPConfigs), args.Error(1)
}

func (m *MockNotificationStorage) UpdatePreference(ctx context.Context, req notification.UpdatePreferenceRequest, userID uuid.UUID) error {
	args := m.Called(ctx, req, userID)
	return args.Error(0)
}
