package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	auth "github.com/raghavyuva/nixopus-api/internal/features/auth/controller"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/stretchr/testify/assert"
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
			name:     "user not found",
			email:    "nonexistent@nixopus.com",
			password: "password",
			setupMocks: func(mockStorage *MockAuthStorage, email, password string) {
				mockStorage.WithUserByEmail(email, nil, types.ErrUserNotFound)
			},
			expectedError: types.ErrUserNotFound,
		},
		{
			name:     "invalid password",
			email:    "nixopus_user3@nixopus.com",
			password: "wrongpassword",
			setupMocks: func(mockStorage *MockAuthStorage, email, password string) {
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

// TestLoginEndpoint tests the Login endpoint of the AuthController.
// It sets up a mock AuthController and tests the following scenarios:
// - Success case: valid email and password
// - Invalid Password: invalid password
// - User Not Found: user not found in storage
func TestLoginEndpoint(t *testing.T) {
	l := logger.NewLogger()
	notificationManager := notification.NewNotificationManager(notification.NewNotificationChannels(), nil)

	t.Run("Success case", func(t *testing.T) {
		authService := setupAuthService(t, l)
		authController := auth.NewAuthController(context.Background(), l, notificationManager, *authService)

		ctx := fuego.NewMockContext(types.LoginRequest{
			Email:    "nixopus_user1@nixopus.com",
			Password: "password",
		})

		response, err := authController.Login(ctx)
		assert.Nil(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "User logged in successfully", response.Message)
		assert.NotNil(t, response.Data)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		authService := setupAuthService(t, l)
		authController := auth.NewAuthController(context.Background(), l, notificationManager, *authService)

		ctx := fuego.NewMockContext(types.LoginRequest{
			Email:    "nixopus_user1@nixopus.com",
			Password: "wrong-password",
		})

		response, err := authController.Login(ctx)
		assert.NotNil(t, err)
		assert.Nil(t, response)
	})

	t.Run("User Not Found", func(t *testing.T) {
		userStorage := &MockAuthStorage{}
		userStorage.WithUserByEmail("nixopus_user2@nixopus.com", nil, errors.New("user not found"))

		authService := service.NewAuthService(userStorage, l, nil, nil, nil, context.Background())
		authController := auth.NewAuthController(context.Background(), l, notificationManager, *authService)

		ctx := fuego.NewMockContext(types.LoginRequest{
			Email:    "nixopus_user2@nixopus.com",
			Password: "password",
		})

		response, err := authController.Login(ctx)
		assert.NotNil(t, err)
		assert.Nil(t, response)
	})
}

func setupAuthService(t *testing.T, l logger.Logger) *service.AuthService {
	userStorage := &MockAuthStorage{}
	userId := uuid.New()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	testUser := CreateTestUser(
		userId.String(),
		"nixopus_user1@nixopus.com",
		string(hashedPassword),
	)

	refreshToken := CreateTestRefreshToken(
		uuid.New().String(),
		userId.String(),
		"test-refresh-token-nixopus_user1@nixopus.com",
		30,
	)

	userStorage.WithCreateRefreshToken(userId, refreshToken, nil)
	userStorage.WithUserByEmail("nixopus_user1@nixopus.com", testUser, nil)

	authService := service.NewAuthService(userStorage, l, nil, nil, nil, context.Background())
	return authService
}
