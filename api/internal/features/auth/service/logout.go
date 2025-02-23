package service

// Logout revokes the given refresh token.
//
// The function takes a refresh token as input and attempts to revoke it by
// updating the corresponding entry in the database. If the token is successfully
// revoked, it returns nil. Otherwise, it returns an error indicating the failure
// to revoke the token.
func (c *AuthService) Logout(refreshToken string) error {
	return c.storage.RevokeRefreshToken(refreshToken)
}
