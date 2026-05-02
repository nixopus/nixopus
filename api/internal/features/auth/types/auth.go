package types

// AdminRegisteredData contains admin registration status
type AdminRegisteredData struct {
	AdminRegistered bool `json:"admin_registered"`
}

// AdminRegisteredResponse is the typed response for admin registration check
type AdminRegisteredResponse struct {
	Status  string              `json:"status"`
	Message string              `json:"message"`
	Data    AdminRegisteredData `json:"data"`
}

// BootstrapUser represents user in bootstrap response
type BootstrapUser struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	IsOnboarded     bool   `json:"isOnboarded"`
	ProvisionStatus string `json:"provisionStatus"`
}

// BootstrapOrg represents org in bootstrap response
type BootstrapOrg struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// BootstrapResponse is the bootstrap API response
type BootstrapResponse struct {
	User                 BootstrapUser  `json:"user"`
	Organizations        []BootstrapOrg `json:"organizations"`
	ActiveOrganizationID *string        `json:"activeOrganizationId"`
	HasServers           bool           `json:"hasServers"`
	ProvisionID          *string        `json:"provisionId,omitempty"`
	ProvisionStep        *string        `json:"provisionStep,omitempty"`
}
