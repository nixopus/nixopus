package types

type ResourceType string

const (
	ResourceTypeUser            ResourceType = "user"
	ResourceTypeOrganization    ResourceType = "organization"
	ResourceTypeRole            ResourceType = "role"
	ResourceTypePermission      ResourceType = "permission"
	ResourceTypeDomain          ResourceType = "domain"
	ResourceTypeGithubConnector ResourceType = "github-connector"
	ResourceTypeNotification    ResourceType = "notification"
	ResourceTypeFileManager     ResourceType = "file-manager"
	ResourceTypeDeploy          ResourceType = "deploy"
	ResourceTypeAudit           ResourceType = "audit"
)
