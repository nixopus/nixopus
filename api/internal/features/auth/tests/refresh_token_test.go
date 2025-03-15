package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func TestRefreshToken(t *testing.T) {
	table := []struct {
		token string
	}{
		{"test-refresh-token-nixopus_user1@nixopus.com"},
		{"test-refresh-token-nixopus_user2@nixopus.com"},
		{"test-refresh-token-nixopus_user3@nixopus.com"},
	}

	mockStorage := NewMockAuthStorage()
	mockLogger := logger.NewLogger()

	for _, test := range table {
		userId := uuid.New().String()
		refreshToken := CreateTestRefreshToken(
			uuid.New().String(),
			userId,
			test.token,
			30)
		user:=CreateTestUser(userId, "nixopus_user1@nixopus.com", "password")
		mockStorage.WithGetRefreshToken(test.token, refreshToken, nil)
		mockStorage.WithUserByID(userId, user, nil)
		mockStorage.WithRevokeRefreshToken(test.token,nil)
		mockStorage.WithCreateRefreshToken(user.ID, refreshToken, nil)
	}

	authService := service.NewAuthService(mockStorage, mockLogger, nil, nil, nil, context.Background())

	for _, test := range table {
		_, err := authService.RefreshToken(types.RefreshTokenRequest{RefreshToken: test.token})
		if err != nil {
			t.Errorf("Error refreshing token: %v", err)
		}
	}
}
