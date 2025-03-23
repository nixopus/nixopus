package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	docker_types "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/google/uuid"
	"github.com/moby/term"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// buildImageFromDockerfile builds a Docker image from a Dockerfile using the provided DeployerConfig. It logs
// the deployment status and image build output to the database, and returns the name of the built image.
func (s *DeployService) buildImageFromDockerfile(b DeployerConfig) (string, error) {
	s.addLog(b.application.ID, types.LogStartingDockerImageBuild, b.deployment_config.ID)
	s.updateStatus(b.deployment_config.ID, shared_types.Building, b.appStatus.ID)
	archive, err := s.createBuildContextArchive(b.contextPath)
	if err != nil {
		return "", err
	}

	s.logger.Log("considering dockerfile base path as ", b.application.DockerfilePath, "")
	dockerfile_path := "Dockerfile"
	if b.application.DockerfilePath != "" {
		dockerfile_path = b.application.DockerfilePath
	}
	buildOptions := s.createBuildOptions(b, dockerfile_path)
	resp, err := s.dockerRepo.BuildImage(buildOptions, archive)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	logReader := &LogReader{
		Reader:            resp.Body,
		ApplicationID:     b.application.ID,
		DeployService:     s,
		deployment_config: b.deployment_config,
	}

	err = s.processBuildOutput(logReader)
	if err != nil {
		return "", err
	}

	s.addLog(b.application.ID, types.LogDockerImageBuiltSuccessfully, b.deployment_config.ID)
	return b.application.Name, nil
}

// createBuildContextArchive creates a tar archive of the build context at the provided path.
// It returns the archive as an io.Reader and an error if the archive creation fails.
func (s *DeployService) createBuildContextArchive(contextPath string) (io.Reader, error) {
	buildContextTar, err := archive.TarWithOptions(contextPath, &archive.TarOptions{})
	if err != nil {
		return nil, types.ErrFailedToCreateTarFromContext
	}
	return buildContextTar, nil
}

// createBuildOptions creates a docker_types.ImageBuildOptions struct based on the provided DeployerConfig.
// The returned ImageBuildOptions include the following settings:
// - Dockerfile: the path to the Dockerfile
// - Remove: true, to remove intermediate containers
// - Tags: two tags, one for the latest version of the image and one for the deployment-specific version
// - NoCache: true if the force flag is set in the deployment config, false otherwise
// - ForceRemove: true if the force flag is set in the deployment config, false otherwise
// - BuildArgs: a map of build variables extracted from the deployment request
// - Labels: a map of labels extracted from the deployment request
// - BuildID: a unique identifier for the build
func (s *DeployService) createBuildOptions(b DeployerConfig, dockerfile_path string) docker_types.ImageBuildOptions {
	return docker_types.ImageBuildOptions{
		Dockerfile:  dockerfile_path,
		Remove:      true,
		Tags:        []string{fmt.Sprintf("%s:latest", b.application.Name)},
		NoCache:     b.deployment.Force,
		ForceRemove: b.deployment.Force,
		BuildArgs:   s.prepareBuildArgs(b),
		Labels:      s.prepareLabels(b),
		BuildID:     b.deployment_config.ID.String(), // build id is the deployment id
	}
}

// prepareBuildArgs takes a DeployerConfig and returns a map of build arguments extracted from the build variables
// specified in the deployment request. The returned map has the same keys as the build variables, and the values
// are pointers to the same strings as the build variables. This is because the docker build options requires
// the build arguments to be pointers to strings.
func (s *DeployService) prepareBuildArgs(d DeployerConfig) map[string]*string {
	buildArgs := make(map[string]*string)
	for k, v := range GetMapFromString(d.application.BuildVariables) {
		value := v
		buildArgs[k] = &value
	}
	return buildArgs
}

func (s *DeployService) prepareLabels(d DeployerConfig) map[string]string {
	labels := make(map[string]string)
	labels["com.application.id"] = d.application.ID.String()
	labels["com.application.name"] = d.application.Name
	labels["com.deployment.id"] = d.deployment_config.ID.String()
	labels["com.commit_hash"] = d.deployment_config.CommitHash
	labels["com.user_id"] = string(d.application.UserID.String())
	return labels
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
