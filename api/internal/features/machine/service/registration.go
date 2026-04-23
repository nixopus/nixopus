package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
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

type RegistrationService struct {
	storage            *storage.RegistrationStorage
	featureFlagService *ff_service.FeatureFlagService
	billingChecker     MachineBillingChecker
	logger             logger.Logger
	ctx                context.Context
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

	signer, err := cryptossh.ParsePrivateKey([]byte(*sshKey.PrivateKeyEncrypted))
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
	client, err := cryptossh.Dial("tcp", addr, config)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("SSH dial failed for machine %s: %v", machineID, err), orgID.String())
		s.storage.MarkMachineInactive(machineID)
		return &types.VerifyMachineResponse{Status: "failed", IsActive: false}, nil
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("SSH session failed for machine %s: %v", machineID, err), orgID.String())
		s.storage.MarkMachineInactive(machineID)
		return &types.VerifyMachineResponse{Status: "failed", IsActive: false}, nil
	}
	defer session.Close()

	if err := session.Run("echo ok"); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("SSH command failed for machine %s: %v", machineID, err), orgID.String())
		s.storage.MarkMachineInactive(machineID)
		return &types.VerifyMachineResponse{Status: "failed", IsActive: false}, nil
	}

	if err := s.storage.MarkMachineActive(machineID); err != nil {
		return nil, fmt.Errorf("failed to update machine status: %w", err)
	}

	return &types.VerifyMachineResponse{Status: "success", IsActive: true}, nil
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
