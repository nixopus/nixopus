package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// Logout revokes the given refresh token.
//
// The function takes a refresh token as input and attempts to revoke it by
// updating the corresponding entry in the database. If the token is successfully
// revoked, it returns nil. Otherwise, it returns an error indicating the failure
// to revoke the token.
func (c *AuthService) Logout(refreshToken string) error {
	c.logger.Log(logger.Info, "Revoking refresh token", refreshToken)
	if refreshToken == "" {
		return types.ErrRefreshTokenIsRequired
	}

	token, err := c.storage.GetRefreshToken(refreshToken)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to get refresh token", err.Error())
		return err
	}
	return c.storage.RevokeRefreshToken(token.Token)
}
