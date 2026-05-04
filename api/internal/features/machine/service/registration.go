package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	ff_service "github.com/nixopus/nixopus/api/internal/features/feature-flags/service"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	api_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
	cryptossh "golang.org/x/crypto/ssh"
)

type MachineBillingChecker interface {
	CanProvision(orgID uuid.UUID) error
}

type NoOpBillingChecker struct{}

func (n *NoOpBillingChecker) CanProvision(orgID uuid.UUID) error {
	return nil
}

// RegistrationRepository is the minimal storage interface required by RegistrationService.
// It is satisfied by *storage.RegistrationStorage in production and by a mock in tests.
type RegistrationRepository interface {
	HostPortExists(orgID uuid.UUID, host string, port int) (bool, error)
	RunInTx(fn func(bun.Tx) error) error
	InsertSSHKeyTx(tx bun.Tx, key *api_types.SSHKey) error
	InsertProvisionDetailsTx(tx bun.Tx, userID, orgID, sshKeyID uuid.UUID, provType string, step api_types.ProvisionStep) error
	GetSSHKeyByID(id, orgID uuid.UUID) (*api_types.SSHKey, error)
	GetSSHKeyStatus(id, orgID uuid.UUID) (bool, *time.Time, error)
	HasActiveAppServers(sshKeyID uuid.UUID) (bool, error)
	SoftDeleteSSHKey(sshKeyID uuid.UUID) error
	UpdateMachineName(sshKeyID uuid.UUID, name string) error
	MarkMachineActive(sshKeyID uuid.UUID) error
	MarkMachineInactive(sshKeyID uuid.UUID) error
}

type RegistrationService struct {
	storage            RegistrationRepository
	featureFlagService *ff_service.FeatureFlagService
	billingChecker     MachineBillingChecker
	logger             logger.Logger
	ctx                context.Context
	parsePrivateKeyFn  func(privateKey []byte) (cryptossh.Signer, error)                                         // nil -> cryptossh.ParsePrivateKey
	dialSSHFn          func(network, addr string, config *cryptossh.ClientConfig) (RegistrationSSHClient, error) // nil -> cryptossh.Dial
}

// RegistrationSSHSession abstracts SSH session operations for testing VerifyMachine.
type RegistrationSSHSession interface {
	Run(cmd string) error
	Close() error
}

// RegistrationSSHClient abstracts SSH client operations for testing VerifyMachine.
type RegistrationSSHClient interface {
	NewSession() (RegistrationSSHSession, error)
	Close() error
}

type registrationSSHClientAdapter struct {
	client *cryptossh.Client
}

func (a *registrationSSHClientAdapter) NewSession() (RegistrationSSHSession, error) {
	sess, err := a.client.NewSession()
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (a *registrationSSHClientAdapter) Close() error {
	return a.client.Close()
}

func NewRegistrationService(
	s *storage.RegistrationStorage,
	ffs *ff_service.FeatureFlagService,
	bc MachineBillingChecker,
	l logger.Logger,
	ctx context.Context,
) *RegistrationService {
	if bc == nil {
		bc = &NoOpBillingChecker{}
	}
	return &RegistrationService{
		storage:            s,
		featureFlagService: ffs,
		billingChecker:     bc,
		logger:             l,
		ctx:                ctx,
	}
}

// NewRegistrationServiceWith creates a RegistrationService with an injectable storage
// interface, intended for use in tests.
func NewRegistrationServiceWith(
	s RegistrationRepository,
	ffs *ff_service.FeatureFlagService,
	bc MachineBillingChecker,
	l logger.Logger,
	ctx context.Context,
) *RegistrationService {
	if bc == nil {
		bc = &NoOpBillingChecker{}
	}
	return &RegistrationService{
		storage:            s,
		featureFlagService: ffs,
		billingChecker:     bc,
		logger:             l,
		ctx:                ctx,
	}
}

func (s *RegistrationService) getParsePrivateKey() func(privateKey []byte) (cryptossh.Signer, error) {
	if s.parsePrivateKeyFn != nil {
		return s.parsePrivateKeyFn
	}
	return cryptossh.ParsePrivateKey
}

func (s *RegistrationService) getDialSSH() func(network, addr string, config *cryptossh.ClientConfig) (RegistrationSSHClient, error) {
	if s.dialSSHFn != nil {
		return s.dialSSHFn
	}
	return func(network, addr string, config *cryptossh.ClientConfig) (RegistrationSSHClient, error) {
		client, err := cryptossh.Dial(network, addr, config)
		if err != nil {
			return nil, err
		}
		return &registrationSSHClientAdapter{client: client}, nil
	}
}

// SetParsePrivateKeyFnForTest injects a private key parser for tests.
func (s *RegistrationService) SetParsePrivateKeyFnForTest(fn func(privateKey []byte) (cryptossh.Signer, error)) {
	s.parsePrivateKeyFn = fn
}

// SetDialSSHFnForTest injects an SSH dial function for tests.
func (s *RegistrationService) SetDialSSHFnForTest(fn func(network, addr string, config *cryptossh.ClientConfig) (RegistrationSSHClient, error)) {
	s.dialSSHFn = fn
}

func (s *RegistrationService) CreateMachine(orgID uuid.UUID, userID uuid.UUID, req types.CreateMachineRequest) (*types.CreateMachineResponse, error) {
	port := req.Port
	if port == 0 {
		port = 22
	}
	user := req.User
	if user == "" {
		user = "root"
	}

	exists, err := s.storage.HostPortExists(orgID, req.Host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to check host uniqueness: %w", err)
	}
	if exists {
		return nil, types.ErrDuplicateHost
	}

	privateKeyPEM, publicKeyStr, fingerprint, err := generateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	keyType := "rsa"
	keySize := 4096
	authMethod := "key"
	sshKey := &api_types.SSHKey{
		ID:                  uuid.New(),
		OrganizationID:      orgID,
		Name:                req.Name,
		Host:                &req.Host,
		User:                &user,
		Port:                &port,
		PublicKey:           &publicKeyStr,
		PrivateKeyEncrypted: &privateKeyPEM,
		KeyType:             &keyType,
		KeySize:             &keySize,
		Fingerprint:         &fingerprint,
		AuthMethod:          authMethod,
		IsActive:            false,
		IsDefault:           false,
	}

	if err := s.storage.RunInTx(func(tx bun.Tx) error {
		if err := s.storage.InsertSSHKeyTx(tx, sshKey); err != nil {
			return fmt.Errorf("failed to insert ssh key: %w", err)
		}
		if err := s.storage.InsertProvisionDetailsTx(tx, userID, orgID, sshKey.ID, "user_owned", api_types.ProvisionStepCompleted); err != nil {
			return fmt.Errorf("failed to insert provision details: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &types.CreateMachineResponse{
		ID:        sshKey.ID.String(),
		Name:      req.Name,
		Host:      req.Host,
		Port:      port,
		User:      user,
		PublicKey: publicKeyStr,
	}, nil
}

func (s *RegistrationService) VerifyMachine(orgID uuid.UUID, machineID uuid.UUID) (*types.VerifyMachineResponse, error) {
	sshKey, err := s.storage.GetSSHKeyByID(machineID, orgID)
	if err != nil {
		return nil, fmt.Errorf("machine not found: %w", err)
	}

	if sshKey.PrivateKeyEncrypted == nil || sshKey.Host == nil {
		return &types.VerifyMachineResponse{Status: "failed", IsActive: false}, nil
	}

	signer, err := s.getParsePrivateKey()([]byte(*sshKey.PrivateKeyEncrypted))
	if err != nil {
		return &types.VerifyMachineResponse{Status: "failed", IsActive: false}, nil
	}

	port := 22
	if sshKey.Port != nil {
		port = *sshKey.Port
	}
	user := "root"
	if sshKey.User != nil {
		user = *sshKey.User
	}

	config := &cryptossh.ClientConfig{
		User:            user,
		Auth:            []cryptossh.AuthMethod{cryptossh.PublicKeys(signer)},
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", *sshKey.Host, port)
	client, err := s.getDialSSH()("tcp", addr, config)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("SSH dial failed for machine %s: %v", machineID, err), orgID.String())
		if dbErr := s.storage.MarkMachineInactive(machineID); dbErr != nil {
			s.logger.Log(logger.Error, fmt.Sprintf("Failed to mark machine %s inactive after SSH dial error: %v", machineID, dbErr), orgID.String())
		}
		return &types.VerifyMachineResponse{Status: "failed", IsActive: false}, nil
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("SSH session failed for machine %s: %v", machineID, err), orgID.String())
		if dbErr := s.storage.MarkMachineInactive(machineID); dbErr != nil {
			s.logger.Log(logger.Error, fmt.Sprintf("Failed to mark machine %s inactive after SSH session error: %v", machineID, dbErr), orgID.String())
		}
		return &types.VerifyMachineResponse{Status: "failed", IsActive: false}, nil
	}
	defer session.Close()

	if err := session.Run("echo ok"); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("SSH command failed for machine %s: %v", machineID, err), orgID.String())
		if dbErr := s.storage.MarkMachineInactive(machineID); dbErr != nil {
			s.logger.Log(logger.Error, fmt.Sprintf("Failed to mark machine %s inactive after SSH command error: %v", machineID, dbErr), orgID.String())
		}
		return &types.VerifyMachineResponse{Status: "failed", IsActive: false}, nil
	}

	if err := s.storage.MarkMachineActive(machineID); err != nil {
		return nil, fmt.Errorf("failed to update machine status: %w", err)
	}

	return &types.VerifyMachineResponse{Status: "success", IsActive: true}, nil
}

func (s *RegistrationService) RenameMachine(orgID uuid.UUID, machineID uuid.UUID, name string) (*types.RenameMachineResponse, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, types.ErrNameRequired
	}
	if len(trimmed) > 255 {
		return nil, types.ErrNameTooLong
	}

	_, err := s.storage.GetSSHKeyByID(machineID, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Log(logger.Error, fmt.Sprintf("Machine not found for rename: machineID=%s, orgID=%s", machineID, orgID), orgID.String())
			return nil, types.ErrMachineNotFound
		}
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to get machine for rename: machineID=%s, error=%v", machineID, err), orgID.String())
		return nil, fmt.Errorf("failed to get machine: %w", err)
	}

	if err := s.storage.UpdateMachineName(machineID, trimmed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Log(logger.Error, fmt.Sprintf("Machine not found during rename update: machineID=%s", machineID), orgID.String())
			return nil, types.ErrMachineNotFound
		}
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to update machine name: machineID=%s, error=%v", machineID, err), orgID.String())
		return nil, fmt.Errorf("failed to rename machine: %w", err)
	}

	return &types.RenameMachineResponse{
		ID:   machineID.String(),
		Name: trimmed,
	}, nil
}

func (s *RegistrationService) DeleteMachine(orgID uuid.UUID, machineID uuid.UUID) error {
	_, err := s.storage.GetSSHKeyByID(machineID, orgID)
	if err != nil {
		return fmt.Errorf("machine not found: %w", err)
	}

	hasApps, err := s.storage.HasActiveAppServers(machineID)
	if err != nil {
		return fmt.Errorf("failed to check app servers: %w", err)
	}
	if hasApps {
		return types.ErrMachineHasApps
	}

	return s.storage.SoftDeleteSSHKey(machineID)
}

func (s *RegistrationService) GetSSHStatus(orgID uuid.UUID, machineID uuid.UUID) (*types.SSHStatusResponse, error) {
	isActive, lastUsedAt, err := s.storage.GetSSHKeyStatus(machineID, orgID)
	if err != nil {
		return nil, fmt.Errorf("machine not found: %w", err)
	}

	resp := &types.SSHStatusResponse{
		IsActive: isActive,
	}
	if lastUsedAt != nil {
		resp.LastUsedAt = lastUsedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return resp, nil
}

func generateKeyPair() (privateKeyPEM, publicKeyStr, fingerprint string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate RSA key: %w", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicSSHKey, err := cryptossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create SSH public key: %w", err)
	}

	publicKeyBytes := cryptossh.MarshalAuthorizedKey(publicSSHKey)

	hash := sha256.Sum256(publicSSHKey.Marshal())
	fp := "SHA256:" + base64.StdEncoding.EncodeToString(hash[:])

	return string(privatePEM), string(publicKeyBytes), fp, nil
}
