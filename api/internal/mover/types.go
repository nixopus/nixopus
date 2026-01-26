package mover

import "time"

// MessageType identifies the type of sync message
type MessageType string

const (
	MessageTypeFileChange  MessageType = "file_change"
	MessageTypeFileContent MessageType = "file_content"
	MessageTypeFileDelete  MessageType = "file_delete"
	MessageTypeSync        MessageType = "sync"
	MessageTypeAck         MessageType = "ack"
	MessageTypeError       MessageType = "error"
	MessageTypePing        MessageType = "ping"
	MessageTypePong        MessageType = "pong"
)

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
