package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// Deprecated: Use SupertokensLogout instead
func (c *AuthService) Logout(refreshToken string) error {
	c.logger.Log(logger.Info, "Revoking refresh token", refreshToken)
	if refreshToken == "" {
		return types.ErrRefreshTokenIsRequired
	}

	tx, err := c.storage.BeginTx()
	if err != nil {
		c.logger.Log(logger.Error, "failed to start transaction", err.Error())
		return types.ErrFailedToUpdateUser
	}
	defer tx.Rollback()

	txStorage := c.storage.WithTx(tx)

	token, err := txStorage.GetRefreshToken(refreshToken)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to get refresh token", err.Error())
		return types.ErrInvalidRefreshToken
	}

	if token.RevokedAt != nil {
		c.logger.Log(logger.Error, "Refresh token is already revoked", refreshToken)
		return types.ErrRefreshTokenAlreadyRevoked
	}

	if err := txStorage.RevokeRefreshToken(token.Token); err != nil {
		c.logger.Log(logger.Error, "Failed to revoke refresh token", err.Error())
		return err
	}

	if err := tx.Commit(); err != nil {
		c.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return types.ErrFailedToUpdateUser
	}

	return nil
}
