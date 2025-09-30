package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// Deprecated: Use SupertokensGenerateVerificationToken instead
func (s *AuthService) GenerateVerificationToken(userID string) (string, error) {
	s.logger.Log(logger.Info, "Generating verification token for user", userID)
	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	err := s.storage.StoreVerificationToken(userID, token, expiresAt)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to store verification token", err.Error())
		return "", errors.New("failed to generate verification token")
	}

	s.logger.Log(logger.Info, "Successfully generated verification token", token)
	return token, nil
}

// Deprecated: Use SupertokensVerifyToken instead
func (s *AuthService) VerifyToken(token string) (string, error) {
	if token == "" {
		return "", errors.New("verification token is required")
	}

	s.logger.Log(logger.Info, "Verifying token", token)
	userID, expiresAt, err := s.storage.GetVerificationToken(token)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get verification token", err.Error())
		return "", errors.New("verification token is already used")
	}

	if time.Now().After(expiresAt) {
		s.logger.Log(logger.Error, "Verification token expired", userID)
		return "", errors.New("verification token expired")
	}

	err = s.storage.DeleteVerificationToken(token)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to delete verification token", err.Error())
		return "", errors.New("verification token is already used")
	}

	err = s.MarkEmailAsVerified(userID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to mark email as verified", err.Error())
		return "", errors.New("failed to mark email as verified")
	}

	s.logger.Log(logger.Info, "Successfully verified token for user", userID)
	return userID, nil
}

// Deprecated: Use SupertokensMarkEmailAsVerified instead
func (s *AuthService) MarkEmailAsVerified(userID string) error {
	s.logger.Log(logger.Info, "Marking email as verified for user", userID)
	err := s.storage.UpdateUserEmailVerification(userID, true)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to update email verification status", err.Error())
		return errors.New("failed to update email verification status")
	}

	s.logger.Log(logger.Info, "Successfully marked email as verified for user", userID)
	return nil
}

// Deprecated: Use SupertokensGetUserByID instead
func (s *AuthService) GetUserByID(userID string) (*shared_types.User, error) {
	s.logger.Log(logger.Info, "Getting user by ID", userID)
	user, err := s.storage.FindUserByID(userID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get user by ID", err.Error())
		return nil, errors.New("user not found")
	}
	return user, nil
}
