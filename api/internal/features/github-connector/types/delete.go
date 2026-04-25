package types

type DeleteGithubConnectorRequest struct {
	ID string `json:"id" validate:"required" description:"GitHub connector ID to delete" example:"550e8400-e29b-41d4-a716-446655440000"`
}
