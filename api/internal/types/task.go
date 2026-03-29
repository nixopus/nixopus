package types

import "github.com/google/uuid"

type TaskPayload struct {
	CorrelationID         string // trace a single logical enqueue
	Application           Application
	ApplicationDeployment ApplicationDeployment
	Status                *ApplicationDeploymentStatus
	UpdateOptions         UpdateOptions
	TargetServerIDs       []uuid.UUID `json:"target_server_ids,omitempty"`
}

type UpdateOptions struct {
	Force             bool
	ForceWithoutCache bool
}
