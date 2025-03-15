package tests

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/mock"
)

// MockAuthStorage is a mock implementation of the AuthRepository interface
type MockAuthStorage struct {
	mock.Mock
}

// FindUserByEmail mocks the FindUserByEmail method
func (m *MockAuthStorage) FindUserByEmail(email string) (*types.User, error) {
	args := m.Called(email)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*types.User), args.Error(1)
}

// FindUserByID mocks the FindUserByID method
func (m *MockAuthStorage) FindUserByID(id string) (*types.User, error) {
	args := m.Called(id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*types.User), args.Error(1)
}

// CreateUser mocks the CreateUser method
func (m *MockAuthStorage) CreateUser(user *types.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// UpdateUser mocks the UpdateUser method
func (m *MockAuthStorage) UpdateUser(user *types.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// CreateRefreshToken mocks the CreateRefreshToken method
func (m *MockAuthStorage) CreateRefreshToken(userID uuid.UUID) (*types.RefreshToken, error) {
	args := m.Called(userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*types.RefreshToken), args.Error(1)
}

// GetRefreshToken mocks the GetRefreshToken method
func (m *MockAuthStorage) GetRefreshToken(refreshToken string) (*types.RefreshToken, error) {
	args := m.Called(refreshToken)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*types.RefreshToken), args.Error(1)
}

// GetResetToken mocks the GetResetToken method
func (m *MockAuthStorage) GetResetToken(token string) (*types.User, error) {
	args := m.Called(token)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*types.User), args.Error(1)
}

// RevokeRefreshToken mocks the RevokeRefreshToken method
func (m *MockAuthStorage) RevokeRefreshToken(refreshToken string) error {
	args := m.Called(refreshToken)
	return args.Error(0)
}

// Helper functions to make test setup easier

// WithUserByEmail sets up the mock to return a specific user for FindUserByEmail
func (m *MockAuthStorage) WithUserByEmail(email string, user *types.User, err error) *MockAuthStorage {
	m.On("FindUserByEmail", email).Return(user, err)
	return m
}

// WithUserByID sets up the mock to return a specific user for FindUserByID
func (m *MockAuthStorage) WithUserByID(id string, user *types.User, err error) *MockAuthStorage {
	m.On("FindUserByID", id).Return(user, err)
	return m
}

// WithCreateUser sets up the mock for CreateUser
func (m *MockAuthStorage) WithCreateUser(user *types.User, err error) *MockAuthStorage {
	m.On("CreateUser", user).Return(err)
	return m
}

// WithUpdateUser sets up the mock for UpdateUser
func (m *MockAuthStorage) WithUpdateUser(user *types.User, err error) *MockAuthStorage {
	m.On("UpdateUser", user).Return(err)
	return m
}

// WithCreateRefreshToken sets up the mock for CreateRefreshToken
func (m *MockAuthStorage) WithCreateRefreshToken(userID uuid.UUID, token *types.RefreshToken, err error) *MockAuthStorage {
	m.On("CreateRefreshToken", userID).Return(token, err)
	return m
}

// WithGetRefreshToken sets up the mock for GetRefreshToken
func (m *MockAuthStorage) WithGetRefreshToken(token string, refreshToken *types.RefreshToken, err error) *MockAuthStorage {
	m.On("GetRefreshToken", token).Return(refreshToken, err)
	return m
}

func (m *MockAuthStorage) WithGetRefreshTokenError(token string, err error) *MockAuthStorage {
	m.On("GetRefreshToken", token).Return(nil, err)
	return m
}

// WithGetResetToken sets up the mock for GetResetToken
func (m *MockAuthStorage) WithGetResetToken(token string, user *types.User, err error) *MockAuthStorage {
	m.On("GetResetToken", token).Return(user, err)
	return m
}

// WithRevokeRefreshToken sets up the mock for RevokeRefreshToken
func (m *MockAuthStorage) WithRevokeRefreshToken(token string, err error) *MockAuthStorage {
	m.On("RevokeRefreshToken", token).Return(err)
	return m
}

// CreateTestUser creates a test user object for testing
func CreateTestUser(id string, email string, password string) *types.User {
	userID, _ := uuid.Parse(id)
	return &types.User{
		ID:        userID,
		Email:     email,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestRefreshToken creates a test refresh token for testing
func CreateTestRefreshToken(id string, userID string, token string, expiresInDays int) *types.RefreshToken {
	tokenID, _ := uuid.Parse(id)
	userUUID, _ := uuid.Parse(userID)

	return &types.RefreshToken{
		ID:        tokenID,
		UserID:    userUUID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour * 24 * time.Duration(expiresInDays)),
		CreatedAt: time.Now(),
	}
}

// NewMockAuthStorage creates a new mock auth storage
func NewMockAuthStorage() *MockAuthStorage {
	return &MockAuthStorage{}
}
