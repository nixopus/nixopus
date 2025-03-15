package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func TestLogoutSuccess(t *testing.T) {
	mockStorage := NewMockAuthStorage()
	mockLogger := logger.NewLogger()
	authService := service.NewAuthService(mockStorage, mockLogger, nil, nil, nil, context.Background())
	refreshToken := "test-refresh-token"
	userID := uuid.New()
	token := CreateTestRefreshToken(uuid.New().String(), userID.String(), refreshToken, 30)
	mockStorage.WithGetRefreshToken(refreshToken, token, nil)
	mockStorage.WithRevokeRefreshToken(refreshToken, nil)

	err := authService.Logout(refreshToken)

	if err != nil {
		t.Errorf("Logout failed: %v", err)
	}
	mockStorage.AssertExpectations(t)
}

func TestLogoutInvalidToken(t *testing.T) {
	mockStorage := NewMockAuthStorage()
	mockLogger := logger.NewLogger()
	authService := service.NewAuthService(mockStorage, mockLogger, nil, nil, nil, context.Background())
	refreshToken := "invalid-refresh-token"
	mockStorage.WithGetRefreshTokenError(refreshToken, errors.New("invalid token"))
	err := authService.Logout(refreshToken)

	if err == nil {
		t.Errorf("Expected error for invalid token")
	}
	mockStorage.AssertExpectations(t)
}

func TestLogoutGetRefreshTokenError(t *testing.T) {
	mockStorage := NewMockAuthStorage()
	mockLogger := logger.NewLogger()
	authService := service.NewAuthService(mockStorage, mockLogger, nil, nil, nil, context.Background())
	refreshToken := "test-refresh-token"
	userID := uuid.New()
	token := CreateTestRefreshToken(uuid.New().String(), userID.String(), refreshToken, 30)
	mockStorage.WithGetRefreshToken(refreshToken, token, errors.New("get refresh token error"))

	err := authService.Logout(refreshToken)

	if err == nil {
		t.Errorf("Expected error for GetRefreshToken")
	}
	mockStorage.AssertExpectations(t)
}

func TestLogoutRevokeRefreshTokenError(t *testing.T) {
	mockStorage := NewMockAuthStorage()
	mockLogger := logger.NewLogger()
	authService := service.NewAuthService(mockStorage, mockLogger, nil, nil, nil, context.Background())
	refreshToken := "test-refresh-token"
	userID := uuid.New()
	token := CreateTestRefreshToken(uuid.New().String(), userID.String(), refreshToken, 30)
	mockStorage.WithGetRefreshToken(refreshToken, token, nil)
	mockStorage.WithRevokeRefreshToken(refreshToken, errors.New("revoke refresh token error"))

	err := authService.Logout(refreshToken)

	if err == nil {
		t.Errorf("Expected error for RevokeRefreshToken")
	}
	mockStorage.AssertExpectations(t)
}
