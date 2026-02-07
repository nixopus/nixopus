package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	docker_types "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/google/uuid"
	"github.com/moby/term"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	sshpkg "github.com/raghavyuva/nixopus-api/internal/features/ssh"
	sftputil "github.com/raghavyuva/nixopus-api/internal/live/sftp"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type BuildConfig struct {
	shared_types.TaskPayload
	ContextPath       string
	Force             bool
	ForceWithoutCache bool
	TaskContext       *TaskContext
	Context           context.Context // Context for organization-aware docker service
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

	// Get SSH manager from context to check path and create archive on remote server
	sshManager, err := sshpkg.GetSSHManagerFromContext(b.Context)
	if err != nil {
		b.TaskContext.LogAndUpdateStatus("Failed to get SSH manager: "+err.Error(), shared_types.Failed)
		return "", fmt.Errorf("failed to get SSH manager: %w", err)
	}

	// Check if path exists on remote server using SFTP with retry logic
	b.TaskContext.AddLog("Checking build context path on remote server...")
	sftpClient, err := sftputil.CreateSFTPClientWithRetry(sshManager)
	if err != nil {
		b.TaskContext.LogAndUpdateStatus("Failed to create SFTP client: "+err.Error(), shared_types.Failed)
		return "", fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	info, err := sftpClient.Stat(buildContextPath)
	if err != nil || !info.IsDir() {
		b.TaskContext.LogAndUpdateStatus("Build context path does not exist: "+buildContextPath, shared_types.Failed)
		return "", fmt.Errorf("build context path does not exist: %s", buildContextPath)
	}
	b.TaskContext.AddLog("Build context path verified on remote server")

	dockerfile_path := "Dockerfile"
	if b.Application.DockerfilePath != "" {
		dockerfile_path = strings.TrimPrefix(b.Application.DockerfilePath, "/")
	}

	dockerfileFullPath := filepath.Join(buildContextPath, dockerfile_path)
	b.TaskContext.AddLog("Validating Dockerfile path...")
	_, err = sftpClient.Stat(dockerfileFullPath)
	if err != nil {
		b.TaskContext.LogAndUpdateStatus("Dockerfile not found at path: "+dockerfileFullPath, shared_types.Failed)
		return "", fmt.Errorf("dockerfile not found at path: %s", dockerfileFullPath)
	}
	b.TaskContext.AddLog("Dockerfile validation successful")

	b.TaskContext.AddLog("Creating build context archive on remote server...")
	archive, err := s.createBuildContextArchiveFromRemote(b.Context, buildContextPath)
	if err != nil {
		b.TaskContext.LogAndUpdateStatus("Failed to create build context archive: "+err.Error(), shared_types.Failed)
		return "", err
	}
	b.TaskContext.AddLog("Build context archive created successfully")

	b.TaskContext.AddLog("Starting Docker image build...")

	// Get docker service from context (organization-aware)
	dockerRepo, err := s.getDockerService(b.Context)
	if err != nil {
		b.TaskContext.LogAndUpdateStatus("Failed to get docker service: "+err.Error(), shared_types.Failed)
		return "", fmt.Errorf("failed to get docker service: %w", err)
	}

	buildOptions := s.createBuildOptions(b, dockerfile_path)
	resp, err := dockerRepo.BuildImage(buildOptions, archive)
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

// createBuildContextArchiveFromRemote creates a tar archive of the build context on the remote SSH server.
// It runs tar command on the remote server and streams the output back.
// It returns the archive as an io.Reader and an error if the archive creation fails.
func (s *TaskService) createBuildContextArchiveFromRemote(ctx context.Context, contextPath string) (io.Reader, error) {
	sshManager, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH manager: %w", err)
	}

	// Get connection from pool for streaming (long-running connection)
	// Use retry logic for session creation to handle closed connections
	// Note: This is a special case where we need both session and client for remoteTarReader
	const maxRetries = 2
	var session interface {
		Wait() error
		Close() error
		StdoutPipe() (io.Reader, error)
		Start(string) error
	}
	var client interface {
		Close() error
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		clientConn, connErr := sshManager.Connect()
		if connErr != nil {
			return nil, fmt.Errorf("failed to connect via SSH: %w", connErr)
		}
		client = clientConn

		sess, sessErr := clientConn.NewSession()
		if sessErr != nil {
			if sshpkg.IsClosedConnectionError(sessErr) {
				// Remove bad connection and retry
				sshManager.CloseConnection("")
				if attempt < maxRetries-1 {
					continue
				}
			}
			return nil, fmt.Errorf("failed to create SSH session: %w", sessErr)
		}
		session = sess
		break
	}

	if session == nil {
		return nil, fmt.Errorf("failed to create SSH session after %d attempts", maxRetries)
	}
	// Don't defer session.Close() here - remoteTarReader handles it

	// Create tar command that excludes common directories
	// Use --exclude patterns similar to archive.TarWithOptions
	tarCmd := fmt.Sprintf("cd %s && tar --exclude='.git' --exclude='node_modules' --exclude='vendor' -czf - .", contextPath)

	// Get stdout pipe to stream tar output
	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start the command
	if err := session.Start(tarCmd); err != nil {
		return nil, fmt.Errorf("failed to start tar command: %w", err)
	}

	// Return a reader that will close the session and client when done
	// Note: We don't defer Close() here because we need to keep them alive
	// until the reader is done. The remoteTarReader will handle cleanup.
	return &remoteTarReader{
		reader:  stdout,
		session: session,
		client:  client,
		done:    make(chan struct{}),
	}, nil
}

// remoteTarReader wraps the SSH session stdout and ensures cleanup
type remoteTarReader struct {
	reader  io.Reader
	session interface {
		Wait() error
		Close() error
	}
	client interface {
		Close() error
	}
	closed bool
	done   chan struct{}
}

func (r *remoteTarReader) Read(p []byte) (n int, err error) {
	if r.closed {
		return 0, io.EOF
	}
	n, err = r.reader.Read(p)
	if err == io.EOF {
		// Wait for session to complete, then close both session and client
		go func() {
			r.session.Wait()
			r.session.Close()
			if r.client != nil {
				r.client.Close()
			}
			close(r.done)
		}()
		r.closed = true
	} else if err != nil {
		// If there's an error reading, clean up immediately
		go func() {
			r.session.Close()
			if r.client != nil {
				r.client.Close()
			}
			close(r.done)
		}()
		r.closed = true
	}
	return n, err
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
