package service

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"strings"
	// "github.com/docker/docker/api/types/image"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// containsSensitiveKeyword checks if a key likely contains sensitive information
func containsSensitiveKeyword(key string) bool {
	sensitiveKeywords := []string{
		"password", "secret", "token", "key", "auth", "credential", "private",
	}

	lowerKey := strings.ToLower(key)
	for _, word := range sensitiveKeywords {
		if strings.Contains(lowerKey, word) {
			return true
		}
	}

	return false
}

// collectContainerLogs collects logs from a running container and adds them to application logs
func (s *DeployService) collectContainerLogs(applicationID uuid.UUID, containerID string) {
	ctx := context.Background()
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	}

	logs, err := s.dockerRepo.ContainerLogs(ctx, containerID, options)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to attach to container logs: %s", err.Error()), "")
		s.addLog(applicationID, fmt.Sprintf("Failed to attach to container logs: %s", err.Error()))
		return
	}
	defer logs.Close()

	scanner := bufio.NewScanner(logs)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 8 {
			logLine := line[8:]
			if options.Timestamps {
				parts := strings.SplitN(logLine, " ", 2)
				if len(parts) == 2 {
					logLine = parts[0] + " " + parts[1]
				}
			}

			s.addLog(applicationID, fmt.Sprintf("Container: %s", logLine))
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Error reading container logs: %s", err.Error()), "")
		s.addLog(applicationID, fmt.Sprintf("Error reading container logs: %s", err.Error()))
	}
}
