package types

type TaskPayload struct {
	CorrelationID         string // trace a single logical enqueue
	Application           Application
	ApplicationDeployment ApplicationDeployment
	Status                *ApplicationDeploymentStatus
	UpdateOptions         UpdateOptions
}

type UpdateOptions struct {
	Force             bool
	ForceWithoutCache bool
}
