package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	docker_types "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/google/uuid"
	"github.com/moby/term"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type BuildConfig struct {
	shared_types.TaskPayload
	ContextPath       string
	Force             bool
	ForceWithoutCache bool
	TaskContext       *TaskContext
}

// buildImageFromDockerfile builds a Docker image from a Dockerfile using the provided DeployerConfig. It logs
// the deployment status and image build output to the database, and returns the name of the built image.
func (s *TaskService) BuildImage(b BuildConfig) (string, error) {
	b.TaskContext.LogAndUpdateStatus("Starting image build", shared_types.Building)
	// For monorepo setups, we need to consider the base path
	buildContextPath := b.ContextPath
	if b.Application.BasePath != "" && b.Application.BasePath != "/" {
		buildContextPath = filepath.Join(b.ContextPath, b.Application.BasePath)
	}

	if _, err := os.Stat(buildContextPath); os.IsNotExist(err) {
		b.TaskContext.LogAndUpdateStatus("Build context path does not exist: "+buildContextPath, shared_types.Failed)
		return "", fmt.Errorf("build context path does not exist: %s", buildContextPath)
	}

	b.TaskContext.AddLog("Creating build context archive...")
	archive, err := s.createBuildContextArchive(buildContextPath)
	if err != nil {
		b.TaskContext.LogAndUpdateStatus("Failed to create build context archive: "+err.Error(), shared_types.Failed)
		return "", err
	}
	b.TaskContext.AddLog("Build context archive created successfully")

	dockerfile_path := "Dockerfile"
	if b.Application.DockerfilePath != "" {
		dockerfile_path = strings.TrimPrefix(b.Application.DockerfilePath, "/")
	}

	dockerfileFullPath := filepath.Join(buildContextPath, dockerfile_path)
	b.TaskContext.AddLog("Validating Dockerfile path...")
	if _, err := os.Stat(dockerfileFullPath); os.IsNotExist(err) {
		b.TaskContext.LogAndUpdateStatus("Dockerfile not found at path: "+dockerfileFullPath, shared_types.Failed)
		return "", fmt.Errorf("dockerfile not found at path: %v", err)
	}
	b.TaskContext.AddLog("Dockerfile validation successful")

	b.TaskContext.AddLog("Starting Docker image build...")
	buildOptions := s.createBuildOptions(b, dockerfile_path)
	resp, err := s.DockerRepo.BuildImage(buildOptions, archive)
	if err != nil {
		b.TaskContext.LogAndUpdateStatus("Failed to build image: "+err.Error(), shared_types.Failed)
		return "", err
	}
	b.TaskContext.AddLog("Docker build started successfully")
	defer resp.Body.Close()

	logReader := &LogReader{
		Reader:            resp.Body,
		ApplicationID:     b.Application.ID,
		DeployService:     s,
		deployment_config: &b.ApplicationDeployment,
		TaskContext:       b.TaskContext,
	}

	b.TaskContext.AddLog("Processing build output...")
	err = s.processBuildOutput(logReader)
	if err != nil {
		b.TaskContext.LogAndUpdateStatus("Failed to process build output: "+err.Error(), shared_types.Failed)
		return "", err
	}
	b.TaskContext.AddLog("Build output processing completed")

	b.TaskContext.LogAndUpdateStatus("Image built successfully", shared_types.Deploying)

	return b.Application.Name, nil
}

// createBuildContextArchive creates a tar archive of the build context at the provided path.
// It returns the archive as an io.Reader and an error if the archive creation fails.
func (s *TaskService) createBuildContextArchive(contextPath string) (io.Reader, error) {
	buildContextTar, err := archive.TarWithOptions(contextPath, &archive.TarOptions{
		ExcludePatterns: []string{".git", "node_modules", "vendor"},
	})
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
func (s *TaskService) createBuildOptions(b BuildConfig, dockerfile_path string) docker_types.ImageBuildOptions {
	return docker_types.ImageBuildOptions{
		Dockerfile:  dockerfile_path,
		Remove:      true,
		Tags:        []string{fmt.Sprintf("%s:latest", b.Application.Name)},
		NoCache:     b.ForceWithoutCache,
		ForceRemove: b.Force,
		BuildArgs:   s.prepareBuildArgs(b),
		Labels:      s.prepareLabels(b),
		BuildID:     b.ApplicationDeployment.ID.String(), // build id is the deployment id
	}
}

// prepareBuildArgs takes a DeployerConfig and returns a map of build arguments extracted from the build variables
// specified in the deployment request. The returned map has the same keys as the build variables, and the values
// are pointers to the same strings as the build variables. This is because the docker build options requires
// the build arguments to be pointers to strings.
func (s *TaskService) prepareBuildArgs(d BuildConfig) map[string]*string {
	buildArgs := make(map[string]*string)
	for k, v := range GetMapFromString(d.Application.BuildVariables) {
		value := v
		buildArgs[k] = &value
	}
	return buildArgs
}

func (s *TaskService) prepareLabels(d BuildConfig) map[string]string {
	labels := make(map[string]string)
	labels["com.application.id"] = d.Application.ID.String()
	labels["com.application.name"] = d.Application.Name
	labels["com.deployment.id"] = d.ApplicationDeployment.ID.String()
	labels["com.commit_hash"] = d.ApplicationDeployment.CommitHash
	labels["com.user_id"] = string(d.Application.UserID.String())
	return labels
}

// processBuildOutput reads the build output stream from the provided LogReader and
// displays the output stream as JSON messages to the standard output. It captures
// any errors that occur during the build output processing and logs the error
// messages to the application. If an error occurs during the build output
// processing, it returns the error.
func (s *TaskService) processBuildOutput(logReader *LogReader) error {
	termFd, isTerm := term.GetFdInfo(os.Stdout)
	err := jsonmessage.DisplayJSONMessagesStream(logReader, os.Stdout, termFd, isTerm, nil)
	if err != nil {
		errorMsg := fmt.Sprintf("Build output processing failed: %v", err)
		s.Logger.Log(logger.Error, errorMsg, logReader.deployment_config.ID.String())
		logReader.TaskContext.AddLog(errorMsg)
		return err
	}
	return nil
}

// LogReader is a custom io.Reader that captures logs from Docker operations and adds them to application logs
type LogReader struct {
	Reader            io.Reader
	ApplicationID     uuid.UUID
	DeployService     *TaskService
	buffer            []byte
	deployment_config *shared_types.ApplicationDeployment
	TaskContext       *TaskContext
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
				r.DeployService.Logger.Log(logger.Error, "Build: "+string(line), r.deployment_config.ID.String())
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
		r.DeployService.Logger.Log(logger.Info, "Build: "+jsonMsg.Stream, r.deployment_config.ID.String())
		r.TaskContext.AddLog("Build: " + jsonMsg.Stream)
	} else if jsonMsg.Status != "" {
		status := jsonMsg.Status
		if jsonMsg.Progress != nil {
			status += " " + jsonMsg.Progress.String()
		}
		r.DeployService.Logger.Log(logger.Info, "Build: "+status, r.deployment_config.ID.String())
		r.TaskContext.AddLog("Build: " + status)
	} else if jsonMsg.Error != nil {
		r.DeployService.Logger.Log(logger.Error, "Build error: "+jsonMsg.Error.Message, r.deployment_config.ID.String())
		r.TaskContext.AddLog("Build error: " + jsonMsg.Error.Message)
	}
}
