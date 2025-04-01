package service

import (
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/xlzd/gotp"
)

func (s *AuthService) SetupTwoFactor(user *shared_types.User) (types.TwoFactorSetupResponse, error) {
	s.logger.Log(logger.Info, "setting up 2FA for user", user.Email)

	secret := gotp.RandomSecret(16)
	totp := gotp.NewDefaultTOTP(secret)
	uri := totp.ProvisioningUri(user.Email, "Nixopus")

	user.TwoFactorSecret = secret
	user.TwoFactorEnabled = false

	if err := s.storage.UpdateUser(user); err != nil {
		s.logger.Log(logger.Error, "failed to update user with 2FA secret", err.Error())
		return types.TwoFactorSetupResponse{}, types.ErrFailedToSetup2FA
	}

	return types.TwoFactorSetupResponse{
		Secret: secret,
		QRCode: uri,
	}, nil
}

func (s *AuthService) VerifyTwoFactor(user *shared_types.User, code string) error {
	s.logger.Log(logger.Info, "verifying 2FA code for user", user.Email)

	totp := gotp.NewDefaultTOTP(user.TwoFactorSecret)
	valid := totp.Verify(code, time.Now().Unix())

	if !valid {
		s.logger.Log(logger.Error, "invalid 2FA code", "")
		return types.ErrInvalid2FACode
	}

	user.TwoFactorEnabled = true
	if err := s.storage.UpdateUser(user); err != nil {
		s.logger.Log(logger.Error, "failed to enable 2FA for user", err.Error())
		return types.ErrFailedToEnable2FA
	}

	return nil
}

func (s *AuthService) DisableTwoFactor(user *shared_types.User) error {
	s.logger.Log(logger.Info, "disabling 2FA for user", user.Email)

	user.TwoFactorEnabled = false
	user.TwoFactorSecret = ""
	if err := s.storage.UpdateUser(user); err != nil {
		s.logger.Log(logger.Error, "failed to disable 2FA for user", err.Error())
		return types.ErrFailedToDisable2FA
	}

	return nil
}

func (s *AuthService) VerifyTwoFactorCode(user *shared_types.User, code string) error {
	s.logger.Log(logger.Info, "verifying 2FA code for login", user.Email)

	totp := gotp.NewDefaultTOTP(user.TwoFactorSecret)
	valid := totp.Verify(code, time.Now().Unix())

	if !valid {
		s.logger.Log(logger.Error, "invalid 2FA code", "")
		return types.ErrInvalid2FACode
	}

	return nil
}
