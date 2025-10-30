package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *ServersService) verifySSHConnection(req types.CreateServerRequest) error {
	s.logger.Log(logger.Info, "verifying SSH connection", fmt.Sprintf("host=%s, port=%d, user=%s", req.Host, req.Port, req.Username))

	sshClient := &ssh.SSH{
		Host: req.Host,
		User: req.Username,
		Port: uint(req.Port),
	}

	if req.SSHPassword != nil && *req.SSHPassword != "" {
		sshClient.Password = *req.SSHPassword
	}

	if req.SSHPrivateKeyPath != nil && *req.SSHPrivateKeyPath != "" {
		sshClient.PrivateKey = *req.SSHPrivateKeyPath
	}

	client, err := sshClient.Connect()
	if err != nil {
		msg := err.Error()
		s.logger.Log(logger.Error, "SSH connection verification failed", fmt.Sprintf("host=%s, port=%d, user=%s, error=%s", req.Host, req.Port, req.Username, msg))
		return classifySSHError(msg)
	}
	defer client.Close()

	s.logger.Log(logger.Info, "SSH connection verified successfully", fmt.Sprintf("host=%s, port=%d, user=%s", req.Host, req.Port, req.Username))
	return nil
}

// CreateServer creates a new server in the application.
//
// It takes a CreateServerRequest, which contains the server details, and a user ID.
// The user ID is used to associate the server with a user.
//
// It returns a CreateServerResponse containing the server ID, and an error.
// The error is either ErrServerAlreadyExists, or any error that occurred
// while creating the server in the storage layer.
func (s *ServersService) CreateServer(req types.CreateServerRequest, userID string, organizationID string) (types.CreateServerResponse, error) {
	s.logger.Log(logger.Info, "create server request received", fmt.Sprintf("server_name=%s, host=%s, user_id=%s", req.Name, req.Host, userID))

	_, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return types.CreateServerResponse{}, types.ErrInvalidUserID
	}

	if userID == "" {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return types.CreateServerResponse{}, types.ErrInvalidUserID
	}

	validator := validation.NewValidator(s.storage)
	if err := validator.ValidateCreateServerRequest(req); err != nil {
		return types.CreateServerResponse{}, err
	}

	if err := s.verifySSHConnection(req); err != nil {
		s.logger.Log(logger.Error, "SSH connection verification failed", err.Error())
		return types.CreateServerResponse{}, err
	}

	org, err := s.store.Organization.GetOrganization(organizationID)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving organization", err.Error())
		return types.CreateServerResponse{}, fmt.Errorf("organization not found")
	}
	if org == nil || org.ID == uuid.Nil {
		s.logger.Log(logger.Error, "organization not found", organizationID)
		return types.CreateServerResponse{}, fmt.Errorf("organization not found")
	}

	tx, err := s.storage.BeginTx()
	if err != nil {
		s.logger.Log(logger.Error, "failed to start transaction", err.Error())
		return types.CreateServerResponse{}, types.ErrFailedToCreateServer
	}
	defer tx.Rollback()

	txStorage := s.storage.WithTx(tx)

	// Check for existing server by name
	existingServerByName, err := txStorage.GetServerName(req.Name, uuid.MustParse(organizationID))
	if err != nil {
		s.logger.Log(logger.Debug, "error while checking existing server by name", err.Error())
	}

	if existingServerByName != nil {
		s.logger.Log(logger.Error, "server already exists", fmt.Sprintf("server_name=%s", req.Name))
		return types.CreateServerResponse{}, types.ErrServerAlreadyExists
	}

	// Check for existing server by host and port
	existingServerByHost, err := txStorage.GetServerByHost(req.Host, req.Port, uuid.MustParse(organizationID))
	if err != nil {
		s.logger.Log(logger.Debug, "error while checking existing server by host", err.Error())
	}

	if existingServerByHost != nil {
		s.logger.Log(logger.Error, "server with host already exists", fmt.Sprintf("host=%s, port=%d", req.Host, req.Port))
		return types.CreateServerResponse{}, types.ErrServerHostAlreadyExists
	}

	server := &shared_types.Server{
		ID:                uuid.New(),
		Name:              req.Name,
		Description:       req.Description,
		Host:              req.Host,
		Port:              req.Port,
		Username:          req.Username,
		SSHPassword:       req.SSHPassword,
		SSHPrivateKeyPath: req.SSHPrivateKeyPath,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		DeletedAt:         nil,
		UserID:            uuid.MustParse(userID),
		OrganizationID:    uuid.MustParse(organizationID),
	}

	if err := txStorage.CreateServer(server); err != nil {
		s.logger.Log(logger.Error, "error while creating server", err.Error())
		return types.CreateServerResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return types.CreateServerResponse{}, types.ErrFailedToCreateServer
	}

	return types.CreateServerResponse{ID: server.ID.String()}, nil
}

func classifySSHError(msg string) error {
	m := strings.ToLower(msg)
	// DNS/host resolution errors
	if strings.Contains(m, "no such host") || strings.Contains(m, "temporary failure in name resolution") || strings.Contains(m, "name or service not known") {
		return types.ErrInvalidHost
	}
	// Authentication errors
	if strings.Contains(m, "permission denied") || strings.Contains(m, "unable to authenticate") || strings.Contains(m, "auth") {
		return types.ErrSSHAuthenticationFailed
	}
	// Connectivity errors
	if strings.Contains(m, "connection refused") || strings.Contains(m, "no route to host") || strings.Contains(m, "i/o timeout") || strings.Contains(m, "network is unreachable") || strings.Contains(m, "connection timed out") {
		return types.ErrSSHConnectionFailed
	}
	return types.ErrSSHConnectionFailed
}
