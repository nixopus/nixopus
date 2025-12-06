package types

type DeleteGithubConnectorRequest struct {
	ID string `json:"id" validate:"required"`
}
