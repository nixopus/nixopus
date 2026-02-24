package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/google/uuid"
	"github.com/moby/term"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	sshpkg "github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	sftpClient, err := utils.CreateSFTPClientWithRetry(sshManager)
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

	b.TaskContext.AddLog("Starting Docker image build on remote server...")
	buildOutput, err := s.createBuildContextArchiveFromRemote(b.Context, b, buildContextPath, dockerfile_path)
	if err != nil {
		b.TaskContext.LogAndUpdateStatus("Failed to start remote build: "+err.Error(), shared_types.Failed)
		return "", err
	}
	defer buildOutput.Close()
	b.TaskContext.AddLog("Docker build started on remote server")

	logReader := &LogReader{
		Reader:            buildOutput,
		ApplicationID:     b.Application.ID,
		DeployService:     s,
		deployment_config: &b.ApplicationDeployment,
		TaskContext:       b.TaskContext,
		PlainOutput:       true,
	}

	b.TaskContext.AddLog("Processing build output...")
	err = s.processBuildOutput(logReader)
	if err != nil {
		b.TaskContext.LogAndUpdateStatus("Failed to process build output: "+err.Error(), shared_types.Failed)
		return "", err
	}
	b.TaskContext.AddLog("Build output processing completed")

	b.TaskContext.LogAndUpdateStatus("Image built successfully", shared_types.Deploying)

	commitTag := CommitImageTag(b.Application.Name, b.ApplicationDeployment.CommitHash)
	return commitTag, nil
}

// createBuildContextArchiveFromRemote runs docker build on the remote server (build context stays on remote).
// Returns the build output stream for logging. No network transfer of build context.
// The caller must defer Close() on the returned reader to avoid SSH connection leaks.
func (s *TaskService) createBuildContextArchiveFromRemote(ctx context.Context, b BuildConfig, contextPath, dockerfilePath string) (*remoteBuildReader, error) {
	sshManager, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH manager: %w", err)
	}
	clientConn, release, err := sshManager.Borrow("")
	if err != nil {
		return nil, fmt.Errorf("failed to connect via SSH: %w", err)
	}
	session, err := clientConn.NewSession()
	if err != nil {
		release()
		if sshpkg.IsClosedConnectionError(err) {
			sshManager.CloseConnection("")
		}
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	latestTag := fmt.Sprintf("%s:latest", b.Application.Name)
	commitTag := CommitImageTag(b.Application.Name, b.ApplicationDeployment.CommitHash)
	escape := func(x string) string { return "'" + strings.ReplaceAll(x, "'", "'\\''") + "'" }
	quotedPath := escape(contextPath)
	buildCmd := fmt.Sprintf("cd %s && docker build --progress=plain -t %s -t %s -f %s", quotedPath, latestTag, commitTag, dockerfilePath)
	if b.ForceWithoutCache {
		buildCmd += " --no-cache"
	}
	for k, v := range GetMapFromString(b.Application.BuildVariables) {
		buildCmd += " --build-arg " + escape(k+"="+v)
	}
	for k, v := range s.prepareLabels(b) {
		buildCmd += " --label " + escape(k+"="+v)
	}
	buildCmd += " . 2>&1"
	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		release()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	if err := session.Start(buildCmd); err != nil {
		session.Close()
		release()
		return nil, fmt.Errorf("failed to start docker build: %w", err)
	}
	return &remoteBuildReader{
		stdout:  stdout,
		session: session,
		release: release,
	}, nil
}

// remoteBuildReader streams docker build output and returns build failure as Read error when command exits non-zero.
// Always call Close() when done (e.g. via defer) to release the pooled SSH connection borrow.
type remoteBuildReader struct {
	stdout  io.Reader
	session interface {
		Wait() error
		Close() error
	}
	release func()
	closed  bool
	errored bool
}

// Close releases the SSH session and pool borrow. Safe to call multiple times.
func (r *remoteBuildReader) Close() {
	if r.closed {
		return
	}
	r.closed = true
	if r.session != nil {
		r.session.Close()
	}
	if r.release != nil {
		r.release()
	}
}

func (r *remoteBuildReader) Read(p []byte) (n int, err error) {
	if r.closed {
		return 0, io.EOF
	}
	n, err = r.stdout.Read(p)
	if err == io.EOF {
		waitErr := r.checkWait()
		r.Close()
		if waitErr != nil {
			return n, waitErr
		}
		return n, io.EOF
	}
	if err != nil {
		r.Close()
	}
	return n, err
}

func (r *remoteBuildReader) checkWait() error {
	if r.errored {
		return nil
	}
	r.errored = true
	if err := r.session.Wait(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}
	return nil
}

// CommitImageTag returns a Docker image tag based on the app name and commit hash.
// Falls back to "latest" if the commit hash is empty.
func CommitImageTag(appName, commitHash string) string {
	short := shortHash(commitHash)
	if short == "" {
		return appName + ":latest"
	}
	return appName + ":" + short
}

func shortHash(hash string) string {
	if len(hash) >= 8 {
		return hash[:8]
	}
	return hash
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

// processBuildOutput reads the build output stream from the provided LogReader.
// The LogReader's Read method intercepts each line for logging to the database,
// so we just need to drain the reader. Output is discarded rather than written to
// server stdout (which is useless in production and adds I/O overhead).
func (s *TaskService) processBuildOutput(logReader *LogReader) error {
	if logReader.PlainOutput {
		_, err := io.Copy(io.Discard, logReader)
		if err != nil {
			errorMsg := fmt.Sprintf("Build output processing failed: %v", err)
			s.Logger.Log(logger.Error, errorMsg, logReader.deployment_config.ID.String())
			logReader.TaskContext.AddLog(errorMsg)
			return err
		}
		return nil
	}
	termFd, isTerm := term.GetFdInfo(os.Stdout)
	err := jsonmessage.DisplayJSONMessagesStream(logReader, io.Discard, termFd, isTerm, nil)
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
	PlainOutput       bool
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
				level := logger.Error
				if r.PlainOutput {
					level = logger.Info
				}
				msg := "Build: " + string(line)
				r.DeployService.Logger.Log(level, msg, r.deployment_config.ID.String())
				r.TaskContext.AddLog(msg)
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
