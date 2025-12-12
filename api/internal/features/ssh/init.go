package ssh

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/melbahja/goph"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	server_storage "github.com/raghavyuva/nixopus-api/internal/features/servers/storage"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type SSH struct {
	PrivateKey          string `json:"private_key"`
	PublicKey           string `json:"public_key"`
	Host                string `json:"host"`
	User                string `json:"user"`
	Port                uint   `json:"port"`
	Password            string `json:"password"`
	PrivateKeyProtected string `json:"private_key_protected"`
}

func NewSSH() *SSH {
	return &SSH{
		PrivateKey:          config.AppConfig.SSH.PrivateKey,
		Host:                config.AppConfig.SSH.Host,
		User:                config.AppConfig.SSH.User,
		Port:                config.AppConfig.SSH.Port,
		Password:            config.AppConfig.SSH.Password,
		PrivateKeyProtected: config.AppConfig.SSH.PrivateKeyProtected,
	}
}

// NewSSHWithServer creates a new SSH client using server configuration from the database.
// It fetches the active server for the organization and uses its SSH configuration.
func NewSSHWithServer(db *bun.DB, ctx context.Context, organizationID uuid.UUID) *SSH {
	log := logger.NewLogger()

	// Fetch the active server for the organization
	serverStorage := server_storage.ServerStorage{
		DB:  db,
		Ctx: ctx,
	}

	server, err := serverStorage.GetActiveServer(organizationID)
	if err != nil {
		log.Log(logger.Warning, "Failed to fetch active server, falling back to default SSH config", fmt.Sprintf("organizationID: %s, error: %v", organizationID.String(), err))
		return NewSSH()
	}

	if server == nil {
		log.Log(logger.Warning, "No active server found, falling back to default SSH config", fmt.Sprintf("organizationID: %s", organizationID.String()))
		return NewSSH()
	}

	log.Log(logger.Info, "Using SSH with server configuration", fmt.Sprintf("organizationID: %s, server: %s, host: %s, port: %d, user: %s", organizationID.String(), server.Name, server.Host, server.Port, server.Username))

	// Create SSH config from server configuration
	sshConfig := &SSH{
		Host: server.Host,
		User: server.Username,
		Port: uint(server.Port),
	}

	// Set authentication method (password or private key)
	if server.SSHPassword != nil && *server.SSHPassword != "" {
		sshConfig.Password = *server.SSHPassword
		log.Log(logger.Info, "Using password authentication", fmt.Sprintf("server: %s", server.Name))
	} else if server.SSHPrivateKeyPath != nil && *server.SSHPrivateKeyPath != "" {
		// Read private key from file path
		privateKeyBytes, err := os.ReadFile(*server.SSHPrivateKeyPath)
		if err != nil {
			log.Log(logger.Error, "Failed to read SSH private key file, falling back to default SSH config", fmt.Sprintf("server: %s, path: %s, error: %v", server.Name, *server.SSHPrivateKeyPath, err))
			return NewSSH()
		}
		sshConfig.PrivateKey = string(privateKeyBytes)
		log.Log(logger.Info, "Using private key authentication", fmt.Sprintf("server: %s, path: %s", server.Name, *server.SSHPrivateKeyPath))
	} else {
		log.Log(logger.Warning, "No SSH authentication method configured for server, falling back to default SSH config", fmt.Sprintf("server: %s", server.Name))
		return NewSSH()
	}

	return sshConfig
}

func (s *SSH) ConnectWithPassword() (*goph.Client, error) {
	if s.Password == "" {
		return nil, fmt.Errorf("password is required for SSH connection")
	}

	auth := goph.Password(s.Password)

	client, err := goph.NewConn(&goph.Config{
		User:     s.User,
		Addr:     s.Host,
		Port:     uint(s.Port),
		Auth:     auth,
		Callback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection with password: %w", err)
	}

	return client, nil
}

func (s *SSH) Connect() (*goph.Client, error) {
	if s.User == "" || s.Host == "" {
		return nil, fmt.Errorf("user and host are required for SSH connection")
	}

	client, err := s.ConnectWithPrivateKey()
	if err == nil {
		return client, nil
	}

	fmt.Printf("private key connection failed: %v\n", err)

	client, err = s.ConnectWithPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to connect with both private key and password: %w", err)
	}

	return client, nil
}

func parsePort(port string) uint64 {
	if port == "" {
		return 22
	}
	p, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		return 22
	}
	return p
}

func (s *SSH) ConnectWithPrivateKey() (*goph.Client, error) {
	if s.PrivateKey == "" {
		return nil, fmt.Errorf("private key is required for SSH connection")
	}

	auth, err := goph.Key(s.PrivateKey, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH auth from private key: %w", err)
	}

	client, err := goph.NewConn(&goph.Config{
		User:     s.User,
		Addr:     s.Host,
		Port:     uint(s.Port),
		Auth:     auth,
		Callback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection with private key: %w", err)
	}

	return client, nil
}

// func (s *SSH) ConnectWithPrivateKeyProtected() (*goph.Client, error) {
// 	auth, err := goph.Key(s.PrivateKeyProtected, "")

// 	if err != nil {
// 		log.Fatalf("SSH connection failed: %v", err)
// 	}

// 	client, err := goph.NewConn(&goph.Config{
// 		User:     s.User,
// 		Addr:     s.Host,
// 		Port:     uint(s.Port),
// 		Auth:     auth,
// 		Callback: ssh.InsecureIgnoreHostKey(),
// 	})
// 	if err != nil {
// 		log.Fatalf("SSH connection failed: %v", err)
// 	}

// 	defer client.Close()
// 	return client, nil
// }

func (s *SSH) RunCommand(cmd string) (string, error) {
	client, err := s.Connect()
	if err != nil {
		return "", err
	}
	defer client.Close()

	output, err := client.Run(cmd)
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

func (s *SSH) Terminal() {
	client, err := s.Connect()
	if err != nil {
		fmt.Print("Failed to connect to ssh")
		return
	}
	session, err := client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create session: %s\n", err)
		return
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	fileDescriptor := int(os.Stdin.Fd())
	if terminal.IsTerminal(fileDescriptor) {
		originalState, err := terminal.MakeRaw(fileDescriptor)
		if err != nil {
			panic(err)
		}
		defer terminal.Restore(fileDescriptor, originalState)

		termWidth, termHeight, err := terminal.GetSize(fileDescriptor)
		if err != nil {
			panic(err)
		}

		err = session.RequestPty("xterm-256color", termHeight, termWidth, modes)
		if err != nil {
			panic(err)
		}
	}

	err = session.Shell()
	if err != nil {
		return
	}
	session.Wait()
}
