package packer

// DockerfileSpec represents the root JSON structure for a Dockerfile definition.
// This is a simplified, easy-to-use schema for creating Dockerfiles programmatically.
type DockerfileSpec struct {
	// Parser directives (must come first in Dockerfile)
	Directives *ParserDirectives `json:"directives,omitempty"`

	// Global ARG instructions (appear before first FROM)
	GlobalArgs []ArgInstruction `json:"global_args,omitempty"`

	// Build stages (each stage starts with FROM)
	Stages []Stage `json:"stages"`
}

// ParserDirectives represents parser directives at the top of a Dockerfile.
type ParserDirectives struct {
	// Syntax directive: e.g., "docker/dockerfile:1"
	Syntax string `json:"syntax,omitempty"`

	// Escape character: "\\" (backslash) or "`" (backtick)
	Escape string `json:"escape,omitempty"`

	// Check directive configuration
	Check *CheckDirective `json:"check,omitempty"`
}

// CheckDirective configures build checks.
type CheckDirective struct {
	// Skip specific checks or "all"
	Skip []string `json:"skip,omitempty"`

	// If true, treat check failures as errors
	Error bool `json:"error,omitempty"`
}

// Stage represents a build stage (everything from one FROM to the next).
type Stage struct {
	// FROM instruction that starts this stage
	From FromInstruction `json:"from"`

	// Instructions in this stage (in order)
	Instructions []Instruction `json:"instructions,omitempty"`
}

// FromInstruction represents a FROM instruction.
type FromInstruction struct {
	// Base image name (required)
	Image string `json:"image"`

	// Image tag (mutually exclusive with Digest)
	Tag string `json:"tag,omitempty"`

	// Image digest (mutually exclusive with Tag)
	Digest string `json:"digest,omitempty"`

	// Platform: e.g., "linux/amd64", "linux/arm64"
	Platform string `json:"platform,omitempty"`

	// Stage name for multi-stage builds
	As string `json:"as,omitempty"`
}

// Instruction is a discriminated union interface for all Dockerfile instructions.
// Each instruction type implements this interface.
type Instruction interface {
	Type() string
}

// BaseInstruction provides common fields for instructions that support exec/shell forms.
type BaseInstruction struct {
	// Command as string (shell form) or array (exec form)
	// For exec form, use array: ["cmd", "arg1", "arg2"]
	// For shell form, use string: "cmd arg1 arg2"
	Command interface{} `json:"command,omitempty"`
}

// RunInstruction represents a RUN instruction.
type RunInstruction struct {
	BaseInstruction

	// Mount options for RUN
	Mount []MountOption `json:"mount,omitempty"`

	// Network mode: "default", "none", "host"
	Network string `json:"network,omitempty"`

	// Security mode: "sandbox", "insecure"
	Security string `json:"security,omitempty"`

	// Device options (experimental)
	Device []string `json:"device,omitempty"`
}

func (r RunInstruction) Type() string { return "run" }

// MountOption represents a mount configuration for RUN instructions.
type MountOption struct {
	// Mount type: "bind", "cache", "tmpfs", "secret", "ssh"
	Type string `json:"type"`

	// Target path (required for all types)
	Target string `json:"target"`

	// Source path (for bind mounts)
	Source string `json:"source,omitempty"`

	// From stage (for bind mounts)
	From string `json:"from,omitempty"`

	// Cache ID (for cache mounts)
	ID string `json:"id,omitempty"`

	// Mode (for secret/ssh mounts): octal string like "0600"
	Mode string `json:"mode,omitempty"`

	// UID/GID (for cache mounts)
	UID *int `json:"uid,omitempty"`
	GID *int `json:"gid,omitempty"`

	// Sharing mode (for cache mounts): "shared", "private", "locked"
	Sharing string `json:"sharing,omitempty"`

	// Readonly flag
	Readonly bool `json:"readonly,omitempty"`

	// Required flag (for secret/ssh mounts)
	Required bool `json:"required,omitempty"`
}

// CmdInstruction represents a CMD instruction.
type CmdInstruction struct {
	BaseInstruction

	// If true, this CMD provides default args for ENTRYPOINT
	IsDefaultArgs bool `json:"is_default_args,omitempty"`
}

func (c CmdInstruction) Type() string { return "cmd" }

// EntrypointInstruction represents an ENTRYPOINT instruction.
type EntrypointInstruction struct {
	BaseInstruction
}

func (e EntrypointInstruction) Type() string { return "entrypoint" }

// CopyInstruction represents a COPY instruction.
type CopyInstruction struct {
	// Source paths (required)
	Sources []string `json:"sources"`

	// Destination path (required)
	Destination string `json:"destination"`

	// Copy from another stage or image
	From string `json:"from,omitempty"`

	// Change ownership: "user:group" or "uid:gid"
	Chown string `json:"chown,omitempty"`

	// Change permissions: octal string like "0644"
	Chmod string `json:"chmod,omitempty"`

	// Use hard link instead of copy
	Link bool `json:"link,omitempty"`

	// Create parent directories
	Parents bool `json:"parents,omitempty"`

	// Exclude patterns
	Exclude []string `json:"exclude,omitempty"`
}

func (c CopyInstruction) Type() string { return "copy" }

// AddInstruction represents an ADD instruction (extends COPY).
type AddInstruction struct {
	CopyInstruction

	// SHA256 checksum: "sha256:hash"
	Checksum string `json:"checksum,omitempty"`

	// Keep .git directory when adding from git URL
	KeepGitDir bool `json:"keep_git_dir,omitempty"`

	// Unpack archive files
	Unpack bool `json:"unpack,omitempty"`
}

func (a AddInstruction) Type() string { return "add" }

// EnvInstruction represents an ENV instruction.
type EnvInstruction struct {
	// Environment variables as key-value pairs
	Variables map[string]string `json:"variables"`
}

func (e EnvInstruction) Type() string { return "env" }

// ArgInstruction represents an ARG instruction.
type ArgInstruction struct {
	// Argument name (required)
	Name string `json:"name"`

	// Default value
	DefaultValue string `json:"default_value,omitempty"`
}

func (a ArgInstruction) Type() string { return "arg" }

// ExposeInstruction represents an EXPOSE instruction.
type ExposeInstruction struct {
	// Port definitions
	Ports []PortDefinition `json:"ports"`
}

// PortDefinition represents a port in EXPOSE.
type PortDefinition struct {
	// Port number (1-65535)
	Port int `json:"port"`

	// Protocol: "tcp" or "udp" (defaults to "tcp")
	Protocol string `json:"protocol,omitempty"`
}

func (e ExposeInstruction) Type() string { return "expose" }

// VolumeInstruction represents a VOLUME instruction.
type VolumeInstruction struct {
	// Mount point paths
	Paths []string `json:"paths"`
}

func (v VolumeInstruction) Type() string { return "volume" }

// UserInstruction represents a USER instruction.
type UserInstruction struct {
	// Username or UID (required)
	User string `json:"user"`

	// Group name or GID
	Group string `json:"group,omitempty"`
}

func (u UserInstruction) Type() string { return "user" }

// WorkdirInstruction represents a WORKDIR instruction.
type WorkdirInstruction struct {
	// Working directory path (required)
	Path string `json:"path"`
}

func (w WorkdirInstruction) Type() string { return "workdir" }

// LabelInstruction represents a LABEL instruction.
type LabelInstruction struct {
	// Labels as key-value pairs
	Labels map[string]string `json:"labels"`
}

func (l LabelInstruction) Type() string { return "label" }

// StopsignalInstruction represents a STOPSIGNAL instruction.
type StopsignalInstruction struct {
	// Signal name (e.g., "SIGTERM") or number
	Signal string `json:"signal"`
}

func (s StopsignalInstruction) Type() string { return "stopsignal" }

// HealthcheckInstruction represents a HEALTHCHECK instruction.
type HealthcheckInstruction struct {
	// If true, disables healthcheck
	None bool `json:"none,omitempty"`

	// Health check command (string or array)
	Command interface{} `json:"command,omitempty"`

	// Check interval: e.g., "30s", "5m"
	Interval string `json:"interval,omitempty"`

	// Timeout: e.g., "10s"
	Timeout string `json:"timeout,omitempty"`

	// Start period: e.g., "40s"
	StartPeriod string `json:"start_period,omitempty"`

	// Start interval: e.g., "10s"
	StartInterval string `json:"start_interval,omitempty"`

	// Number of retries
	Retries *int `json:"retries,omitempty"`
}

func (h HealthcheckInstruction) Type() string { return "healthcheck" }

type ShellInstruction struct {
	// Shell command array (exec form only)
	Shell []string `json:"shell"`
}

func (s ShellInstruction) Type() string { return "shell" }

type OnbuildInstruction struct {
	// Nested instruction (cannot be FROM, MAINTAINER, or ONBUILD)
	Instruction Instruction `json:"instruction"`
}

func (o OnbuildInstruction) Type() string { return "onbuild" }

// MaintainerInstruction represents a MAINTAINER instruction (deprecated).
type MaintainerInstruction struct {
	// Maintainer name/email
	Name string `json:"name"`
}

func (m MaintainerInstruction) Type() string { return "maintainer" }

// instructionWrapper is used for JSON unmarshaling of Instruction discriminated union.
type instructionWrapper struct {
	Type string `json:"type"`
}
