package types

type TaskPayload struct {
	Application           Application
	ApplicationDeployment ApplicationDeployment
	Status                *ApplicationDeploymentStatus
	UpdateOptions         UpdateOptions
}

type UpdateOptions struct {
	Force             bool
	ForceWithoutCache bool
}