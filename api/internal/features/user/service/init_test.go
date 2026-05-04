package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/service"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserStorage struct {
	mock.Mock
}

func (m *mockUserStorage) GetUserOrganizationsWithRolesAndPermissions(userID string) ([]types.UserOrganizationsResponse, error) {
	args := m.Called(userID)
	return args.Get(0).([]types.UserOrganizationsResponse), args.Error(1)
}

func (m *mockUserStorage) GetUserById(id string) (*shared_types.User, error) {
	args := m.Called(id)
	return args.Get(0).(*shared_types.User), args.Error(1)
}

func (m *mockUserStorage) UpdateUserName(userID string, userName string, updatedAt time.Time) error {
	args := m.Called(userID, userName, updatedAt)
	return args.Error(0)
}

func (m *mockUserStorage) GetUserSettings(userID string) (*shared_types.UserSettings, error) {
	args := m.Called(userID)
	return args.Get(0).(*shared_types.UserSettings), args.Error(1)
}

func (m *mockUserStorage) UpdateUserSettings(userID string, updates map[string]interface{}) (*shared_types.UserSettings, error) {
	args := m.Called(userID, updates)
	return args.Get(0).(*shared_types.UserSettings), args.Error(1)
}

func (m *mockUserStorage) UpdateUserAvatar(ctx context.Context, userID string, avatarData string) error {
	args := m.Called(ctx, userID, avatarData)
	return args.Error(0)
}

func (m *mockUserStorage) GetUserPreferences(userID string) (*shared_types.UserPreferences, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared_types.UserPreferences), args.Error(1)
}

func (m *mockUserStorage) UpdateUserPreferences(userID string, preferences shared_types.UserPreferencesData) (*shared_types.UserPreferences, error) {
	args := m.Called(userID, preferences)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared_types.UserPreferences), args.Error(1)
}

func (m *mockUserStorage) GetIsOnboarded(userID string) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserStorage) MarkOnboardingComplete(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func TestGetUserOrganizations(t *testing.T) {
	t.Run("user does not exist", func(t *testing.T) {
		st := &mockUserStorage{}
		st.On("GetUserOrganizationsWithRolesAndPermissions", "non-existent-user").Return([]types.UserOrganizationsResponse{}, types.ErrUserDoesNotExist)

		s := service.NewUserService(nil, context.Background(), logger.NewLogger(), st)
		orgs, err := s.GetUserOrganizations("non-existent-user")

		assert.Empty(t, orgs)
		assert.Equal(t, types.ErrUserDoesNotExist, err)
	})

	t.Run("storage returns error", func(t *testing.T) {
		st := &mockUserStorage{}
		st.On("GetUserOrganizationsWithRolesAndPermissions", "user-id").Return([]types.UserOrganizationsResponse{}, errors.New("storage error"))

		s := service.NewUserService(nil, context.Background(), logger.NewLogger(), st)
		orgs, err := s.GetUserOrganizations("user-id")

		assert.Empty(t, orgs)
		assert.NotNil(t, err)
	})

	t.Run("storage returns empty organizations", func(t *testing.T) {
		st := &mockUserStorage{}
		st.On("GetUserOrganizationsWithRolesAndPermissions", "user-id").Return([]types.UserOrganizationsResponse{}, nil)

		s := service.NewUserService(nil, context.Background(), logger.NewLogger(), st)
		orgs, err := s.GetUserOrganizations("user-id")

		assert.Empty(t, orgs)
		assert.Nil(t, err)
	})

	t.Run("storage returns organizations", func(t *testing.T) {
		st := &mockUserStorage{}
		expected := []types.UserOrganizationsResponse{{}}
		st.On("GetUserOrganizationsWithRolesAndPermissions", "user-id").Return(expected, nil)

		s := service.NewUserService(nil, context.Background(), logger.NewLogger(), st)
		orgs, err := s.GetUserOrganizations("user-id")

		assert.NotEmpty(t, orgs)
		assert.Nil(t, err)
		assert.Equal(t, expected, orgs)
	})
}

func TestUpdateUsername(t *testing.T) {
	st := &mockUserStorage{}
	s := service.NewUserService(nil, context.Background(), logger.NewLogger(), st)

	validUUID := uuid.New()
	nilUUID := uuid.Nil
	errGetUser := errors.New("db: get user failed")

	tests := []struct {
		name          string
		id            string
		req           *types.UpdateUserNameRequest
		wantErr       error
		mockUser      *shared_types.User
		mockGetErr    error
		mockUpdateErr error
	}{
		{
			name:          "User exists and update is successful",
			id:            validUUID.String(),
			req:           &types.UpdateUserNameRequest{Name: "new-username"},
			wantErr:       nil,
			mockUser:      &shared_types.User{ID: validUUID},
			mockGetErr:    nil,
			mockUpdateErr: nil,
		},
		{
			name:          "User does not exist",
			id:            nilUUID.String(),
			req:           &types.UpdateUserNameRequest{Name: "new-username"},
			wantErr:       types.ErrUserDoesNotExist,
			mockUser:      &shared_types.User{ID: nilUUID},
			mockGetErr:    nil,
			mockUpdateErr: nil,
		},
		{
			name:          "GetUserById returns storage error",
			id:            validUUID.String(),
			req:           &types.UpdateUserNameRequest{Name: "new-username"},
			wantErr:       errGetUser,
			mockUser:      &shared_types.User{ID: validUUID},
			mockGetErr:    errGetUser,
			mockUpdateErr: nil,
		},
		{
			name:          "Update fails due to storage error",
			id:            validUUID.String(),
			req:           &types.UpdateUserNameRequest{Name: "new-username"},
			wantErr:       types.ErrFailedToUpdateUser,
			mockUser:      &shared_types.User{ID: validUUID},
			mockGetErr:    nil,
			mockUpdateErr: errors.New("storage error"),
		},
		{
			name:          "Empty request body",
			id:            validUUID.String(),
			req:           &types.UpdateUserNameRequest{},
			wantErr:       nil,
			mockUser:      &shared_types.User{ID: validUUID},
			mockGetErr:    nil,
			mockUpdateErr: nil,
		},
		{
			name:          "Nil request",
			id:            validUUID.String(),
			req:           nil,
			wantErr:       types.ErrInvalidRequestType,
			mockUser:      &shared_types.User{ID: validUUID},
			mockGetErr:    nil,
			mockUpdateErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st.ExpectedCalls = nil

			if tt.req != nil {
				st.On("GetUserById", tt.id).Return(tt.mockUser, tt.mockGetErr)
			}

			if tt.req != nil && tt.mockUser.ID != uuid.Nil && tt.mockGetErr == nil {
				st.On("UpdateUserName", tt.mockUser.ID.String(), tt.req.Name, mock.Anything).Return(tt.mockUpdateErr)
			}

			err := s.UpdateUsername(tt.id, tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UpdateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateAvatar(t *testing.T) {
	ctx := context.Background()

	t.Run("nil request returns ErrInvalidRequestType", func(t *testing.T) {
		st := &mockUserStorage{}
		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)

		err := s.UpdateAvatar(ctx, "user-id", nil)
		assert.Equal(t, types.ErrInvalidRequestType, err)
	})

	t.Run("storage error is propagated", func(t *testing.T) {
		st := &mockUserStorage{}
		storageErr := errors.New("avatar update failed")
		st.On("UpdateUserAvatar", mock.Anything, "user-id", "data:image/png;base64,abc").Return(storageErr)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		err := s.UpdateAvatar(ctx, "user-id", &types.UpdateAvatarRequest{AvatarData: "data:image/png;base64,abc"})
		assert.Equal(t, storageErr, err)
	})

	t.Run("success returns nil", func(t *testing.T) {
		st := &mockUserStorage{}
		st.On("UpdateUserAvatar", mock.Anything, "user-id", "data:image/png;base64,abc").Return(nil)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		err := s.UpdateAvatar(ctx, "user-id", &types.UpdateAvatarRequest{AvatarData: "data:image/png;base64,abc"})
		assert.NoError(t, err)
	})
}

func TestIsOnboarded(t *testing.T) {
	ctx := context.Background()

	t.Run("storage error returns false and error", func(t *testing.T) {
		st := &mockUserStorage{}
		storageErr := errors.New("db error")
		st.On("GetIsOnboarded", "user-id").Return(false, storageErr)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		result, err := s.IsOnboarded("user-id")
		assert.False(t, result)
		assert.Equal(t, storageErr, err)
	})

	t.Run("user is onboarded", func(t *testing.T) {
		st := &mockUserStorage{}
		st.On("GetIsOnboarded", "user-id").Return(true, nil)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		result, err := s.IsOnboarded("user-id")
		assert.True(t, result)
		assert.NoError(t, err)
	})

	t.Run("user is not onboarded", func(t *testing.T) {
		st := &mockUserStorage{}
		st.On("GetIsOnboarded", "user-id").Return(false, nil)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		result, err := s.IsOnboarded("user-id")
		assert.False(t, result)
		assert.NoError(t, err)
	})
}

func TestMarkOnboardingComplete(t *testing.T) {
	ctx := context.Background()

	t.Run("storage error is propagated", func(t *testing.T) {
		st := &mockUserStorage{}
		storageErr := errors.New("db error")
		st.On("MarkOnboardingComplete", "user-id").Return(storageErr)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		err := s.MarkOnboardingComplete("user-id")
		assert.Equal(t, storageErr, err)
	})

	t.Run("success returns nil", func(t *testing.T) {
		st := &mockUserStorage{}
		st.On("MarkOnboardingComplete", "user-id").Return(nil)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		err := s.MarkOnboardingComplete("user-id")
		assert.NoError(t, err)
	})
}

func TestGetSettings(t *testing.T) {
	ctx := context.Background()

	t.Run("returns settings on success", func(t *testing.T) {
		st := &mockUserStorage{}
		expected := &shared_types.UserSettings{FontFamily: "system-ui"}
		st.On("GetUserSettings", "user-id").Return(expected, nil)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		settings, err := s.GetSettings("user-id")
		require.NoError(t, err)
		assert.Equal(t, expected, settings)
	})

	t.Run("returns error from storage", func(t *testing.T) {
		st := &mockUserStorage{}
		storageErr := errors.New("db error")
		st.On("GetUserSettings", "user-id").Return((*shared_types.UserSettings)(nil), storageErr)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		settings, err := s.GetSettings("user-id")
		assert.Nil(t, settings)
		assert.Equal(t, storageErr, err)
	})
}

func TestUpdateFont(t *testing.T) {
	ctx := context.Background()

	t.Run("returns updated settings", func(t *testing.T) {
		st := &mockUserStorage{}
		expected := &shared_types.UserSettings{FontFamily: "monospace", FontSize: 14}
		st.On("UpdateUserSettings", "user-id", mock.Anything).Return(expected, nil)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		settings, err := s.UpdateFont("user-id", "monospace", 14)
		require.NoError(t, err)
		assert.Equal(t, expected, settings)
	})

	t.Run("returns error from storage", func(t *testing.T) {
		st := &mockUserStorage{}
		storageErr := errors.New("db error")
		st.On("UpdateUserSettings", "user-id", mock.Anything).Return((*shared_types.UserSettings)(nil), storageErr)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		_, err := s.UpdateFont("user-id", "monospace", 14)
		assert.Equal(t, storageErr, err)
	})
}

func TestUpdateTheme(t *testing.T) {
	ctx := context.Background()

	t.Run("returns updated settings", func(t *testing.T) {
		st := &mockUserStorage{}
		expected := &shared_types.UserSettings{Theme: "dark"}
		st.On("UpdateUserSettings", "user-id", mock.Anything).Return(expected, nil)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		settings, err := s.UpdateTheme("user-id", "dark")
		require.NoError(t, err)
		assert.Equal(t, expected, settings)
	})

	t.Run("returns error from storage", func(t *testing.T) {
		st := &mockUserStorage{}
		storageErr := errors.New("db error")
		st.On("UpdateUserSettings", "user-id", mock.Anything).Return((*shared_types.UserSettings)(nil), storageErr)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		_, err := s.UpdateTheme("user-id", "dark")
		assert.Equal(t, storageErr, err)
	})
}

func TestUpdateLanguage(t *testing.T) {
	ctx := context.Background()

	t.Run("returns updated settings", func(t *testing.T) {
		st := &mockUserStorage{}
		expected := &shared_types.UserSettings{Language: "fr"}
		st.On("UpdateUserSettings", "user-id", mock.Anything).Return(expected, nil)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		settings, err := s.UpdateLanguage("user-id", "fr")
		require.NoError(t, err)
		assert.Equal(t, expected, settings)
	})

	t.Run("returns error from storage", func(t *testing.T) {
		st := &mockUserStorage{}
		storageErr := errors.New("db error")
		st.On("UpdateUserSettings", "user-id", mock.Anything).Return((*shared_types.UserSettings)(nil), storageErr)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		_, err := s.UpdateLanguage("user-id", "fr")
		assert.Equal(t, storageErr, err)
	})
}

func TestUpdateAutoUpdate(t *testing.T) {
	ctx := context.Background()

	t.Run("returns updated settings", func(t *testing.T) {
		st := &mockUserStorage{}
		expected := &shared_types.UserSettings{AutoUpdate: true}
		st.On("UpdateUserSettings", "user-id", mock.Anything).Return(expected, nil)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		settings, err := s.UpdateAutoUpdate("user-id", true)
		require.NoError(t, err)
		assert.Equal(t, expected, settings)
	})

	t.Run("returns error from storage", func(t *testing.T) {
		st := &mockUserStorage{}
		storageErr := errors.New("db error")
		st.On("UpdateUserSettings", "user-id", mock.Anything).Return((*shared_types.UserSettings)(nil), storageErr)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		_, err := s.UpdateAutoUpdate("user-id", false)
		assert.Equal(t, storageErr, err)
	})
}

func TestGetUserPreferences(t *testing.T) {
	ctx := context.Background()

	t.Run("returns preferences on success", func(t *testing.T) {
		st := &mockUserStorage{}
		expected := &shared_types.UserPreferences{}
		st.On("GetUserPreferences", "user-id").Return(expected, nil)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		prefs, err := s.GetUserPreferences("user-id")
		require.NoError(t, err)
		assert.Equal(t, expected, prefs)
	})

	t.Run("returns error from storage", func(t *testing.T) {
		st := &mockUserStorage{}
		storageErr := errors.New("db error")
		st.On("GetUserPreferences", "user-id").Return(nil, storageErr)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		prefs, err := s.GetUserPreferences("user-id")
		assert.Nil(t, prefs)
		assert.Equal(t, storageErr, err)
	})
}

func TestUpdateUserPreferences(t *testing.T) {
	ctx := context.Background()
	prefsData := shared_types.UserPreferencesData{DebugMode: true}

	t.Run("returns updated preferences", func(t *testing.T) {
		st := &mockUserStorage{}
		expected := &shared_types.UserPreferences{}
		st.On("UpdateUserPreferences", "user-id", prefsData).Return(expected, nil)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		prefs, err := s.UpdateUserPreferences("user-id", prefsData)
		require.NoError(t, err)
		assert.Equal(t, expected, prefs)
	})

	t.Run("returns error from storage", func(t *testing.T) {
		st := &mockUserStorage{}
		storageErr := errors.New("db error")
		st.On("UpdateUserPreferences", "user-id", prefsData).Return(nil, storageErr)

		s := service.NewUserService(nil, ctx, logger.NewLogger(), st)
		prefs, err := s.UpdateUserPreferences("user-id", prefsData)
		assert.Nil(t, prefs)
		assert.Equal(t, storageErr, err)
	})
}
