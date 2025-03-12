package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	docker_types "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/google/uuid"
	"github.com/moby/term"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
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

// buildImageFromDockerfile builds a Docker image from the specified Dockerfile and context path.
//
// This function logs the start of the Docker image build process, creates a build context archive
// from the provided context path, and generates Docker build options using the specified parameters.
// It then calls the Docker service to build the image and processes the build output logs. Upon
// successful completion, it logs the successful build and returns the image name.
//
// Parameters:
//
//	b - a BuildImageFromDockerFile struct containing the application ID, context path, Dockerfile path,
//	    build arguments, labels, image name, status ID, and deployment configuration.
//
// Returns:
//
//	string - the name of the built Docker image.
//	error - an error if the build process fails at any step, otherwise nil.
func (s *DeployService) buildImageFromDockerfile(b BuildImageFromDockerFile) (string, error) {
	s.addLog(b.applicationID, types.LogStartingDockerImageBuild, b.deployment_config.ID)
	s.updateStatus(b.applicationID, shared_types.Building, b.statusID)

	archive, err := s.createBuildContextArchive(b.contextPath)
	if err != nil {
		return "", err
	}

	dockerfile_path := filepath.Base(b.dockerfile)
	buildOptions := s.createBuildOptions(b, dockerfile_path)
	resp, err := s.dockerRepo.BuildImage(buildOptions, archive)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	logReader := &LogReader{
		Reader:            resp.Body,
		ApplicationID:     b.applicationID,
		DeployService:     s,
		deployment_config: b.deployment_config,
	}

	err = s.processBuildOutput(logReader)
	if err != nil {
		return "", err
	}

	s.addLog(b.applicationID, types.LogDockerImageBuiltSuccessfully, b.deployment_config.ID)
	return b.image_name, nil
}

// createBuildContextArchive creates a tar archive from the specified build context path.
// It uses the Docker archive package to generate the tar file from the provided directory,
// returning an io.Reader for the tar archive. If an error occurs during the tar creation,
// it returns an error with a descriptive message.
//
// Parameters:
//
//	contextPath - the path to the build context directory to be archived.
//
// Returns:
//
//	io.Reader - a reader for the created tar archive.
//	error - an error if the tar creation fails, otherwise nil.
func (s *DeployService) createBuildContextArchive(contextPath string) (io.Reader, error) {
	buildContextTar, err := archive.TarWithOptions(contextPath, &archive.TarOptions{})
	if err != nil {
		return nil, types.ErrFailedToCreateTarFromContext
	}
	return buildContextTar, nil
}

// createBuildOptions creates a Docker image build options struct from the provided BuildImageFromDockerFile
// and Dockerfile path. It populates the struct with the build context path, sets the Dockerfile path,
// enables removal of intermediate containers, tags the image with the provided name and latest tag,
// enables force removal of build cache when the force flag is set, and sets the build arguments and labels.
//
// Parameters:
//
//	b - the BuildImageFromDockerFile containing the build configuration.
//	dockerfile_path - the path to the Dockerfile to use for the build.
//
// Returns:
//
//	docker_types.ImageBuildOptions - the populated image build options struct.
func (s *DeployService) createBuildOptions(b BuildImageFromDockerFile, dockerfile_path string) docker_types.ImageBuildOptions {
	return docker_types.ImageBuildOptions{
		Dockerfile:  dockerfile_path,
		Remove:      true,
		Tags:        []string{fmt.Sprintf("%s:latest", b.image_name)},
		NoCache:     b.force,
		ForceRemove: b.force,
		BuildArgs:   b.buildArgs,
		Labels:      b.labels,
		BuildID:     uuid.New().String(),
	}
}

// processBuildOutput reads the build output stream from the provided LogReader and
// displays the output stream as JSON messages to the standard output. It captures
// any errors that occur during the build output processing and logs the error
// messages to the application. If an error occurs during the build output
// processing, it returns the error.
func (s *DeployService) processBuildOutput(logReader *LogReader) error {
	termFd, isTerm := term.GetFdInfo(os.Stdout)
	err := jsonmessage.DisplayJSONMessagesStream(logReader, os.Stdout, termFd, isTerm, nil)
	if err != nil {
		s.addLog(logReader.ApplicationID, types.ErrProcessingBuildOutput.Error(), logReader.deployment_config.ID)
		return types.ErrProcessingBuildOutput
	}
	return nil
}

// LogReader is a custom io.Reader that captures logs from Docker operations and adds them to application logs
type LogReader struct {
	Reader            io.Reader
	ApplicationID     uuid.UUID
	DeployService     *DeployService
	buffer            []byte
	deployment_config *shared_types.ApplicationDeployment
}

// Read implements the io.Reader interface for LogReader. It reads from the underlying Reader and
// processes any JSON messages received, logging them to the application. If a message is not a valid
// JSON message, it logs the message verbatim. It buffers the input to handle partial messages.
func (r *LogReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if n > 0 {
		r.buffer = append(r.buffer, p[:n]...)
		for {
			idx := bytes.IndexByte(r.buffer, '\n')
			if idx == -1 {
				break
			}

			line := r.buffer[:idx]
			r.buffer = r.buffer[idx+1:]

			var jsonMsg jsonmessage.JSONMessage
			if err := json.Unmarshal(line, &jsonMsg); err == nil {
				r.processJSONMessage(jsonMsg)
			} else {
				r.DeployService.addLog(r.ApplicationID, "Build: "+string(line), r.deployment_config.ID)
			}
		}
	}
	return n, err
}

// processJSONMessage processes a JSONMessage received during the Docker build process.
// It adds the message to the application's logs based on its content. If the message
// contains a stream, it logs the stream content. If the message has a status, it logs
// the status and any accompanying progress. If the message contains an error, it logs
// the error message. This helps in tracking the build process and diagnosing issues.
func (r *LogReader) processJSONMessage(jsonMsg jsonmessage.JSONMessage) {
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
}
