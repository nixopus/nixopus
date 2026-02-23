package mover

// ConnectionStatus represents connection state (CLI UI only)
type ConnectionStatus int

const (
	ConnectionStatusDisconnected ConnectionStatus = iota
	ConnectionStatusConnecting
	ConnectionStatusConnected
	ConnectionStatusReconnecting
)

// ServiceStatus represents service state (CLI UI only)
type ServiceStatus int

const (
	ServiceStatusUnknown ServiceStatus = iota
	ServiceStatusStarting
	ServiceStatusRunning
	ServiceStatusError
)

// DeploymentInfo holds deployment status information (CLI UI only)
type DeploymentInfo struct {
	Status    string   `json:"status"`     // building, deploying, deployed, failed, etc.
	Message   string   `json:"message"`    // Optional status message
	Error     string   `json:"error"`      // Error message if failed
	Logs      []string `json:"logs"`       // Recent deployment logs
	UpdatedAt string   `json:"updated_at"` // Last update timestamp
}

// StatusInfo holds information for the status box (CLI UI only)
type StatusInfo struct {
	ConnectionStatus ConnectionStatus
	FilesSynced      int
	ChangesDetected  int
	ServiceStatus    ServiceStatus
	URL              string
	EnvPath          string          // Path to environment file if detected
	Deployment       *DeploymentInfo // Current deployment status
}
