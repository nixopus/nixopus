package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"golang.org/x/crypto/bcrypt"
)

// TestLogin is a unit test for the Login function of the AuthService.
// It ensures that the service returns the correct user details, access token,
// and refresh token for valid login credentials. The test iterates over a
// table of user emails and passwords, setting up mock storage with hashed
// passwords and refresh tokens. It then verifies that the login response
// contains the expected user email and non-empty tokens, and asserts that
// all mock expectations are met.
func TestLogin(t *testing.T) {
	successTable := []struct {
		email    string
		password string
	}{
		{"nixopus_user1@nixopus.com", "password"},
		{"nixopus_user2@nixopus.com", "password"},
	}

	errorTable := []struct {
		name          string
		email         string
		password      string
		setupMocks    func(mockStorage *MockAuthStorage, email, password string)
		expectedError error
	}{
		{
			name:          "user not found",
			email:         "nonexistent@nixopus.com",
			password:      "password",
			setupMocks:    func(mockStorage *MockAuthStorage, email, password string) {
				mockStorage.WithUserByEmail(email, nil, types.ErrUserNotFound)
			},
			expectedError: types.ErrUserNotFound,
		},
		{
			name:          "invalid password",
			email:         "nixopus_user3@nixopus.com",
			password:      "wrongpassword",
			setupMocks:    func(mockStorage *MockAuthStorage, email, password string) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
				testUser := CreateTestUser(
					uuid.New().String(),
					email,
					string(hashedPassword),
				)
				mockStorage.WithUserByEmail(email, testUser, nil)
			},
			expectedError: types.ErrInvalidPassword,
		},
	}

	mockStorage := NewMockAuthStorage()
	mockLogger := logger.NewLogger()

	for _, test := range successTable {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(test.password), bcrypt.DefaultCost)
		
		userID := uuid.New()
		testUser := CreateTestUser(
			userID.String(),
			test.email,
			string(hashedPassword),
		)
		
		mockStorage.WithUserByEmail(test.email, testUser, nil)
		
		refreshToken := CreateTestRefreshToken(
			uuid.New().String(),
			userID.String(),
			"test-refresh-token-"+test.email,
			30,
		)
		mockStorage.WithCreateRefreshToken(userID, refreshToken, nil)
	}
	
	authService := service.NewAuthService(mockStorage, mockLogger, nil, nil, nil, context.Background())

	for _, test := range successTable {
		t.Run("Success_"+test.email, func(t *testing.T) {
			response, err := authService.Login(test.email, test.password)
			if err != nil {
				t.Fatalf("Login failed for %s: %v", test.email, err)
			}
			
			if response.User.Email != test.email {
				t.Errorf("Expected user email %s, got %s", test.email, response.User.Email)
			}
			
			if response.AccessToken == "" {
				t.Errorf("Expected access token but got empty string")
			}
			
			if response.RefreshToken == "" {
				t.Errorf("Expected refresh token but got empty string")
			}
			
			if response.ExpiresIn != 900 {
				t.Errorf("Expected ExpiresIn to be 900, got %d", response.ExpiresIn)
			}
		})
	}
	
	for _, test := range errorTable {
		t.Run(test.name, func(t *testing.T) {
			mockStorage := NewMockAuthStorage()
			mockLogger := logger.NewLogger()
			
			test.setupMocks(mockStorage, test.email, test.password)
			
			authService := service.NewAuthService(mockStorage, mockLogger, nil, nil, nil, context.Background())
			
			_, err := authService.Login(test.email, test.password)
			if err != test.expectedError {
				t.Errorf("Expected error %v, got %v", test.expectedError, err)
			}
		})
	}
}