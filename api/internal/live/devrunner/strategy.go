package devrunner

// FrameworkStrategy defines the interface for framework-specific dev server configuration.
// Each framework (Next.js, Vite, etc.) implements this interface to provide
// the necessary configuration for running its dev server in a container.
type FrameworkStrategy interface {
	// Name returns the framework identifier (e.g., "nextjs", "vite", "remix")
	Name() string

	// GetBaseImage returns the Docker image to use for the dev container
	GetBaseImage() string

	// GetInstallCommand returns the command to install dependencies
	// Returns nil if no install step is needed
	GetInstallCommand() []string

	// GetDevCommand returns the command to start the dev server
	GetDevCommand() []string

	// GetDefaultPort returns the default port the dev server listens on
	GetDefaultPort() int

	// GetEnvVars returns environment variables needed for dev mode
	GetEnvVars() map[string]string

	// GetHealthCheckPath returns the HTTP path for health checks
	GetHealthCheckPath() string

	// NeedsPolling returns true if file watching requires polling mode
	// Some frameworks need this when running in Docker due to inotify limitations
	NeedsPolling() bool

	// GetWorkdir returns the working directory inside the container
	GetWorkdir() string

	// GetReadyLogPattern returns a regex pattern to detect when the dev server is ready
	// This is used to know when the container is ready to serve requests
	GetReadyLogPattern() string

	// DetectMainFile analyzes the project at the given path and returns
	// the main entry file for the framework. Returns empty string if not applicable
	// or if the framework doesn't use a main file (e.g., uses npm scripts).
	DetectMainFile(projectPath string) string
}
