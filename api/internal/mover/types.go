package mover

import "time"

// MessageType identifies the type of sync message
type MessageType string

const (
	MessageTypeFileChange       MessageType = "file_change"
	MessageTypeFileContent      MessageType = "file_content"
	MessageTypeFileDelete       MessageType = "file_delete"
	MessageTypeEnvVars          MessageType = "env_vars"
	MessageTypeSyncComplete     MessageType = "sync_complete"
	MessageTypeSync             MessageType = "sync"
	MessageTypeAck              MessageType = "ack"
	MessageTypeError            MessageType = "error"
	MessageTypePing             MessageType = "ping"
	MessageTypePong             MessageType = "pong"
	MessageTypeManifest         MessageType = "manifest"
	MessageTypePipelineProgress MessageType = "pipeline_progress"
	MessageTypeBuildStatus      MessageType = "build_status"
	MessageTypeBuildLog         MessageType = "build_log"
	MessageTypeDeploymentStatus MessageType = "deployment_status"
)

// PipelineProgressPayload carries real-time progress from the pipeline agent
// during Dockerfile generation. Sent from server to client over WebSocket.
type PipelineProgressPayload struct {
	StageId string `json:"stage_id"` // e.g. "resolve-repo", "dockerfile-generate"
	Message string `json:"message"`  // human-readable progress message
}

// BuildStatusPayload carries build lifecycle events from server to client.
// Sent at key points during the build process so the CLI can show progress.
type BuildStatusPayload struct {
	Phase   string `json:"phase"`           // "starting", "generating_dockerfile", "dockerfile_ready", "building_container", "error"
	Message string `json:"message"`         // human-readable description
	Error   string `json:"error,omitempty"` // error details when phase is "error"
}

// BuildLogPayload carries a single build log line from the database via
// PostgreSQL NOTIFY -> Gateway -> WebSocket -> CLI.
type BuildLogPayload struct {
	Log       string `json:"log"`
	Timestamp string `json:"timestamp,omitempty"`
}

// DeploymentStatusPayload carries a deployment status change from the database.
type DeploymentStatusPayload struct {
	Status       string `json:"status"`
	DeploymentID string `json:"deployment_id,omitempty"`
}

// EnvVarsPayload carries parsed env key-value pairs from the client (not the raw file)
type EnvVarsPayload struct {
	Vars map[string]string `json:"vars"`
}

// ManifestPayload is sent by server on WebSocket connect with path→checksum for skip logic.
// Phase 3b: RootHash allows client to skip full diff when roots match.
type ManifestPayload struct {
	Paths    map[string]string `json:"paths"`
	RootHash string            `json:"root_hash,omitempty"` // Merkle root of server's file tree
	Version  int               `json:"version,omitempty"`   // For future protocol compatibility
}

// SyncMessage is the envelope for all WebSocket messages
type SyncMessage struct {
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// FileChange represents a single file change event
type FileChange struct {
	Path      string    `json:"path"`
	Operation string    `json:"operation"` // create, modify, delete, rename
	IsDir     bool      `json:"is_dir"`
	Size      int64     `json:"size,omitempty"`
	Checksum  string    `json:"checksum,omitempty"`
	OldPath   string    `json:"old_path,omitempty"` // For renames
	ModTime   time.Time `json:"mod_time,omitempty"`
}

// FileContent represents file data being transferred
type FileContent struct {
	Path        string `json:"path"`
	ChunkIndex  int    `json:"chunk_index"`
	TotalChunks int    `json:"total_chunks"`
	Data        []byte `json:"data"`
	Checksum    string `json:"checksum"`
}

// SyncStatus represents the status of a sync operation
type SyncStatus struct {
	FilesChanged     int   `json:"files_changed"`
	BytesTransferred int64 `json:"bytes_transferred"`
	RebuildRequired  bool  `json:"rebuild_required"`
	BuildDurationMs  int   `json:"build_duration_ms,omitempty"`
}

// ErrorPayload represents an error message
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ConnectionStatus represents connection state
type ConnectionStatus int

const (
	ConnectionStatusDisconnected ConnectionStatus = iota
	ConnectionStatusConnecting
	ConnectionStatusConnected
	ConnectionStatusReconnecting
)

// ServiceStatus represents service state
type ServiceStatus int

const (
	ServiceStatusUnknown ServiceStatus = iota
	ServiceStatusStarting
	ServiceStatusRunning
	ServiceStatusError
)

// DeploymentInfo holds deployment status information
type DeploymentInfo struct {
	Status    string   `json:"status"`     // building, deploying, deployed, failed, etc.
	Message   string   `json:"message"`    // Optional status message
	Error     string   `json:"error"`      // Error message if failed
	Logs      []string `json:"logs"`       // Recent deployment logs
	UpdatedAt string   `json:"updated_at"` // Last update timestamp
}

// StatusInfo holds information for the status box
type StatusInfo struct {
	ConnectionStatus ConnectionStatus
	FilesSynced      int
	ChangesDetected  int
	ServiceStatus    ServiceStatus
	URL              string
	EnvPath          string          // Path to environment file if detected
	Deployment       *DeploymentInfo // Current deployment status
}
