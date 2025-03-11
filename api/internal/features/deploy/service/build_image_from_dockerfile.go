package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	docker_types "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/google/uuid"
	"github.com/moby/term"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type BuildImageFromDockerFile struct {
	applicationID     uuid.UUID
	contextPath       string
	dockerfile        string
	force             bool
	buildArgs         map[string]*string
	labels            map[string]string
	image_name        string
	statusID          uuid.UUID
	deployment_config *shared_types.ApplicationDeployment
}

// buildImageFromDockerfile builds a Docker image from a Dockerfile in the
// contextPath directory. The Dockerfile is specified by the dockerfile
// parameter, which is relative to the contextPath directory. If force is true,
// the build will be forced, even if the image already exists. The buildArgs
// parameter is a map of build arguments, which are passed to the Dockerfile as
// environment variables. The labels parameter is a map of labels, which are
// applied to the built image. The image_name parameter is the name of the
// built image. The function returns the ID of the built image, or an error if
// the build fails.
func (s *DeployService) buildImageFromDockerfile(b BuildImageFromDockerFile) (string, error) {
	s.addLog(b.applicationID, "Starting Docker image build from Dockerfile", b.deployment_config.ID)
	s.updateStatus(b.applicationID, shared_types.Building, b.statusID)

	buildContextTar, err := archive.TarWithOptions(b.contextPath, &archive.TarOptions{})
	if err != nil {
		errMsg := fmt.Sprintf("failed to create tar from build context: %s", err.Error())
		s.addLog(b.applicationID, errMsg, b.deployment_config.ID)
		return "", errors.New(errMsg)
	}
	s.addLog(b.applicationID, "Created build context archive for Docker build", b.deployment_config.ID)

	dockerfile_path := filepath.Base(b.dockerfile)
	s.addLog(b.applicationID, fmt.Sprintf("Using Dockerfile: %s", dockerfile_path), b.deployment_config.ID)

	buildOptions := docker_types.ImageBuildOptions{
		Dockerfile:  dockerfile_path,
		Remove:      true,
		Tags:        []string{fmt.Sprintf("%s:latest", b.image_name)},
		NoCache:     b.force,
		ForceRemove: b.force,
		BuildArgs:   b.buildArgs,
		Labels:      b.labels,
		BuildID:     uuid.New().String(),
	}

	resp, err := s.dockerRepo.BuildImage(buildOptions, buildContextTar)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to start image build: %s", err.Error())
		s.addLog(b.applicationID, errMsg, b.deployment_config.ID)
		return "", errors.New(errMsg)
	}
	defer resp.Body.Close()

	logReader := &LogReader{
		Reader:            resp.Body,
		ApplicationID:     b.applicationID,
		DeployService:     s,
		deployment_config: b.deployment_config,
	}

	termFd, isTerm := term.GetFdInfo(os.Stdout)
	err = jsonmessage.DisplayJSONMessagesStream(logReader, os.Stdout, termFd, isTerm, nil)
	if err != nil {
		errMsg := fmt.Sprintf("Error processing build output: %s", err.Error())
		s.addLog(b.applicationID, errMsg, b.deployment_config.ID)
		return "", errors.New(errMsg)
	}

	s.addLog(b.applicationID, "Docker image build completed successfully", b.deployment_config.ID)
	return b.image_name, nil
}

// LogReader is a custom io.Reader that captures logs from Docker operations and adds them to application logs
type LogReader struct {
	Reader            io.Reader
	ApplicationID     uuid.UUID
	DeployService     *DeployService
	buffer            []byte
	deployment_config *shared_types.ApplicationDeployment
}

func (r *LogReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if n > 0 {
		r.buffer = append(r.buffer, p[:n]...)
		for {
			idx := -1
			for i, b := range r.buffer {
				if b == '\n' {
					idx = i
					break
				}
			}

			if idx == -1 {
				break
			}

			line := r.buffer[:idx]
			r.buffer = r.buffer[idx+1:]

			var jsonMsg jsonmessage.JSONMessage
			if err := json.Unmarshal(line, &jsonMsg); err == nil {
				if jsonMsg.Stream != "" {
					r.DeployService.addLog(r.ApplicationID, "Build: "+jsonMsg.Stream, r.deployment_config.ID)
				} else if jsonMsg.Status != "" {
					status := jsonMsg.Status
					if jsonMsg.Progress != nil {
						status += " " + jsonMsg.Progress.String()
					}
					r.DeployService.addLog(r.ApplicationID, "Build: "+status, r.deployment_config.ID)
				} else if jsonMsg.Error != nil {
					r.DeployService.addLog(r.ApplicationID, "Build error: "+jsonMsg.Error.Message, r.deployment_config.ID)
				}
			} else {
				r.DeployService.addLog(r.ApplicationID, "Build: "+string(line), r.deployment_config.ID)
			}
		}
	}

	return n, err
}
