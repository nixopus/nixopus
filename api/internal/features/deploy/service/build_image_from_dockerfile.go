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

// buildImageFromDockerfile builds a Docker image from a Dockerfile in the
// contextPath directory. The Dockerfile is specified by the dockerfile
// parameter, which is relative to the contextPath directory. If force is true,
// the build will be forced, even if the image already exists. The buildArgs
// parameter is a map of build arguments, which are passed to the Dockerfile as
// environment variables. The labels parameter is a map of labels, which are
// applied to the built image. The image_name parameter is the name of the
// built image. The function returns the ID of the built image, or an error if
// the build fails.
func (s *DeployService) buildImageFromDockerfile(applicationID uuid.UUID, contextPath string, dockerfile string, force bool, buildArgs map[string]*string, labels map[string]string, image_name string, statusID uuid.UUID) (string, error) {
	s.addLog(applicationID, "Starting Docker image build from Dockerfile")
	s.updateStatus(applicationID, shared_types.Building, statusID)

	buildContextTar, err := archive.TarWithOptions(contextPath, &archive.TarOptions{})
	if err != nil {
		errMsg := fmt.Sprintf("failed to create tar from build context: %s", err.Error())
		s.addLog(applicationID, errMsg)
		return "", errors.New(errMsg)
	}
	s.addLog(applicationID, "Created build context archive for Docker build")

	dockerfile_path := filepath.Base(dockerfile)
	s.addLog(applicationID, fmt.Sprintf("Using Dockerfile: %s", dockerfile_path))

	buildOptions := docker_types.ImageBuildOptions{
		Dockerfile:  dockerfile_path,
		Remove:      true,
		Tags:        []string{fmt.Sprintf("%s:latest", image_name)},
		NoCache:     force,
		ForceRemove: force,
		BuildArgs:   buildArgs,
		Labels:      labels,
		BuildID:     uuid.New().String(),
	}

	resp, err := s.dockerRepo.BuildImage(buildOptions, buildContextTar)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to start image build: %s", err.Error())
		s.addLog(applicationID, errMsg)
		return "", errors.New(errMsg)
	}
	defer resp.Body.Close()

	logReader := &LogReader{
		Reader:        resp.Body,
		ApplicationID: applicationID,
		DeployService: s,
	}

	termFd, isTerm := term.GetFdInfo(os.Stdout)
	err = jsonmessage.DisplayJSONMessagesStream(logReader, os.Stdout, termFd, isTerm, nil)
	if err != nil {
		errMsg := fmt.Sprintf("Error processing build output: %s", err.Error())
		s.addLog(applicationID, errMsg)
		return "", errors.New(errMsg)
	}

	s.addLog(applicationID, "Docker image build completed successfully")
	return image_name, nil
}

// LogReader is a custom io.Reader that captures logs from Docker operations and adds them to application logs
type LogReader struct {
	Reader        io.Reader
	ApplicationID uuid.UUID
	DeployService *DeployService
	buffer        []byte
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
					r.DeployService.addLog(r.ApplicationID, "Build: "+jsonMsg.Stream)
				} else if jsonMsg.Status != "" {
					status := jsonMsg.Status
					if jsonMsg.Progress != nil {
						status += " " + jsonMsg.Progress.String()
					}
					r.DeployService.addLog(r.ApplicationID, "Build: "+status)
				} else if jsonMsg.Error != nil {
					r.DeployService.addLog(r.ApplicationID, "Build error: "+jsonMsg.Error.Message)
				}
			} else {
				r.DeployService.addLog(r.ApplicationID, "Build: "+string(line))
			}
		}
	}

	return n, err
}
