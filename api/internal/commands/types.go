package commands

// ValidateAPIKeyRequest represents the request body for API key validation
type ValidateAPIKeyRequest struct {
	APIKey string `json:"api_key"`
}

// ValidateAPIKeyResponse represents the response from API key validation endpoint
type ValidateAPIKeyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Valid   bool   `json:"valid"`
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	APIKey               string            `json:"api_key"`
	Name                 string            `json:"name"`
	Repository           string            `json:"repository"`
	Branch               string            `json:"branch,omitempty"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
}

// CreateProjectResponse represents the response from project creation endpoint
type CreateProjectResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	ProjectID string `json:"project_id"`
}

// CreateSessionRequest represents the request body for creating a live session
type CreateSessionRequest struct {
	ApplicationID  string                 `json:"application_id"`
	UserID         string                 `json:"user_id"`
	OrganizationID string                 `json:"organization_id"`
	Config         map[string]interface{} `json:"config"`
}

// CreateSessionResponse represents the response from session creation endpoint
type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
}
