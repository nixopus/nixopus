package tests

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	"github.com/stretchr/testify/mock"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"time"
)

type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) GetUserOrganizationsWithRolesAndPermissions(userID string) ([]types.UserOrganizationsResponse, error) {
	args := m.Called(userID)
	return args.Get(0).([]types.UserOrganizationsResponse), args.Error(1)
}

func (m *MockUserStorage) GetUserById(id string) (*shared_types.User, error) {
	args := m.Called(id)
	return args.Get(0).(*shared_types.User), args.Error(1)
}

func (m *MockUserStorage) UpdateUserName(userID string, userName string, updatedAt time.Time) error {
	args := m.Called(userID, userName, updatedAt)
	return args.Error(0)
}

func (m *MockUserStorage) GetUserSettings(userID string) (*shared_types.UserSettings, error) {
	args := m.Called(userID)
	return args.Get(0).(*shared_types.UserSettings), args.Error(1)
}

func (m *MockUserStorage) UpdateUserSettings(userID string, updates map[string]interface{}) (*shared_types.UserSettings, error) {
	args := m.Called(userID, updates)
	return args.Get(0).(*shared_types.UserSettings), args.Error(1)
}

func (m *MockUserStorage) UpdateUserAvatar(ctx context.Context, userID string, avatarData string) error {
	args := m.Called(ctx, userID, avatarData)
	return args.Error(0)
}
