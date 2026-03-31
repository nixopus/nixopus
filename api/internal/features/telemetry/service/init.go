package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/storage"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/types"
)

type TelemetryService struct {
	storage storage.TelemetryRepository
	ctx     context.Context
	logger  logger.Logger
	ipSalt  string
}

func NewTelemetryService(repo storage.TelemetryRepository, ctx context.Context, l logger.Logger) *TelemetryService {
	salt := os.Getenv("TELEMETRY_IP_SALT")
	if salt == "" {
		salt = "nixopus-telemetry-default-salt"
	}
	return &TelemetryService{
		storage: repo,
		ctx:     ctx,
		logger:  l,
		ipSalt:  salt,
	}
}

func (s *TelemetryService) TrackInstall(req *types.TrackInstallRequest, clientIP string) error {
	ipHash := s.hashIP(clientIP)

	event := &types.CliInstallation{
		ID:        uuid.New(),
		EventType: req.EventType,
		OS:        req.OS,
		Arch:      req.Arch,
		Version:   req.Version,
		Duration:  req.Duration,
		Error:     req.Error,
		IPHash:    ipHash,
	}

	if err := s.storage.CreateInstallEvent(event); err != nil {
		s.logger.Log(logger.Error, "failed to store install event", err.Error())
		return err
	}

	return nil
}

func (s *TelemetryService) hashIP(ip string) string {
	h := sha256.New()
	h.Write([]byte(s.ipSalt + ip))
	return fmt.Sprintf("%x", h.Sum(nil))
}
