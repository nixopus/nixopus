package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
)

func executeFileStep(sshClient *ssh.SSH, props map[string]interface{}, replacer func(string) string) (string, error) {
	action, _ := props["action"].(string)
	src, _ := props["src"].(string)
	dest, _ := props["dest"].(string)
	src = replacer(src)
	dest = replacer(dest)

	switch action {
	case "move":
		client, err := sshClient.Connect()
		if err != nil {
			return "", err
		}
		s, err := client.NewSftp()
		if err != nil {
			return "", err
		}
		defer s.Close()
		if err := s.Rename(src, dest); err != nil {
			return "", err
		}
		return fmt.Sprintf("moved %s to %s", src, dest), nil
	case "copy":
		// Fallback to remote cp for recursive copy
		return sshClient.RunCommand(fmt.Sprintf("cp -r %s %s", src, dest))
	case "upload":
		client, err := sshClient.Connect()
		if err != nil {
			return "", err
		}
		s, err := client.NewSftp()
		if err != nil {
			return "", err
		}
		defer s.Close()
		f, err := os.Open(src)
		if err != nil {
			return "", err
		}
		defer f.Close()
		if err := s.MkdirAll(filepath.Dir(dest)); err != nil {
			return "", err
		}
		out, err := s.Create(dest)
		if err != nil {
			return "", err
		}
		defer out.Close()
		if _, err := io.Copy(out, f); err != nil {
			return "", err
		}
		return fmt.Sprintf("uploaded to %s", dest), nil
	case "delete":
		client, err := sshClient.Connect()
		if err != nil {
			return "", err
		}
		s, err := client.NewSftp()
		if err != nil {
			return "", err
		}
		defer s.Close()
		if err := s.Remove(dest); err != nil {
			return "", err
		}
		return fmt.Sprintf("deleted %s", dest), nil
	case "mkdir":
		client, err := sshClient.Connect()
		if err != nil {
			return "", err
		}
		s, err := client.NewSftp()
		if err != nil {
			return "", err
		}
		defer s.Close()
		if err := s.MkdirAll(dest); err != nil {
			return "", err
		}
		return fmt.Sprintf("mkdir %s", dest), nil
	default:
		return "unsupported file action", nil
	}
}
