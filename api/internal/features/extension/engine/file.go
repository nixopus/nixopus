package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type fileModule struct{}

func (fileModule) Type() string { return "file" }

type sftpClient interface {
	Close() error
	MkdirAll(path string) error
	Remove(path string) error
	Rename(fromPath string, toPath string) error
	Create(path string) (*sftp.File, error)
}

var (
	runCommandFn = func(c *ssh.SSH, cmd string) (string, error) { return c.RunCommand(cmd) }
	withSFTPFn   = func(c *ssh.SSH, f func(s sftpClient) (string, error)) (string, error) {
		client, err := c.Connect()
		if err != nil {
			return "", err
		}
		s, err := client.NewSftp()
		if err != nil {
			return "", err
		}
		defer s.Close()
		return f(s)
	}
)

type fileAction func(*ssh.SSH, string, string) (string, error)

func handleMove(c *ssh.SSH, src, dest string) (string, error) {
	return withSFTPFn(c, func(s sftpClient) (string, error) {
		if err := s.Rename(src, dest); err != nil {
			return "", fmt.Errorf("failed to move %s to %s: %w", src, dest, err)
		}
		return fmt.Sprintf("moved %s to %s", src, dest), nil
	})
}

func handleCopy(c *ssh.SSH, src, dest string) (string, error) {
	cmd := fmt.Sprintf("cp -r %s %s", src, dest)
	output, err := runCommandFn(c, cmd)
	if err != nil {
		return "", fmt.Errorf("failed to copy %s to %s: %w (output: %s)", src, dest, err, output)
	}
	return fmt.Sprintf("copied %s to %s", src, dest), nil
}

func handleUpload(c *ssh.SSH, src, dest string) (string, error) {
	f, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer f.Close()

	return withSFTPFn(c, func(s sftpClient) (string, error) {
		if err := s.MkdirAll(filepath.Dir(dest)); err != nil {
			return "", fmt.Errorf("failed to create destination directory %s: %w", filepath.Dir(dest), err)
		}
		out, err := s.Create(dest)
		if err != nil {
			return "", fmt.Errorf("failed to create destination file %s: %w", dest, err)
		}
		defer out.Close()
		if _, err := io.Copy(out, f); err != nil {
			return "", fmt.Errorf("failed to copy file content from %s to %s: %w", src, dest, err)
		}
		return fmt.Sprintf("uploaded to %s", dest), nil
	})
}

func handleDelete(c *ssh.SSH, _src, dest string) (string, error) {
	return withSFTPFn(c, func(s sftpClient) (string, error) {
		if err := s.Remove(dest); err != nil {
			return "", fmt.Errorf("failed to delete %s: %w", dest, err)
		}
		return fmt.Sprintf("deleted %s", dest), nil
	})
}

func handleMkdir(c *ssh.SSH, _src, dest string) (string, error) {
	return withSFTPFn(c, func(s sftpClient) (string, error) {
		if err := s.MkdirAll(dest); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", dest, err)
		}
		return fmt.Sprintf("mkdir %s", dest), nil
	})
}

var actionHandlers = map[string]fileAction{
	"move":   handleMove,
	"copy":   handleCopy,
	"upload": handleUpload,
	"delete": handleDelete,
	"mkdir":  handleMkdir,
}

func (fileModule) Execute(sshClient *ssh.SSH, step types.SpecStep, vars map[string]interface{}) (string, func(), error) {
	action, _ := step.Properties["action"].(string)
	src, _ := step.Properties["src"].(string)
	dest, _ := step.Properties["dest"].(string)

	if action == "mkdir" && dest == "" {
		return "", nil, fmt.Errorf("dest is required for mkdir action")
	}

	h, ok := actionHandlers[action]
	if !ok {
		return "", nil, fmt.Errorf("unsupported file action: %s", action)
	}
	out, err := h(sshClient, src, dest)
	if err != nil {
		return "", nil, err
	}

	var compensate func()
	switch action {
	case "move":
		compensate = func() { _, _ = handleMove(sshClient, dest, src) }
	case "copy":
		compensate = func() { _, _ = handleDelete(sshClient, "", dest) }
	case "upload":
		compensate = func() { _, _ = handleDelete(sshClient, "", dest) }
	case "delete":
		// cannot reliably restore deleted content without backup
		compensate = nil
	case "mkdir":
		compensate = func() { _, _ = handleDelete(sshClient, "", dest) }
	}
	return out, compensate, nil
}

func init() {
	RegisterModule(fileModule{})
}
