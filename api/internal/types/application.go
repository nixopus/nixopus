package types

import (
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Application struct {
	bun.BaseModel        `bun:"table:applications,alias:a" swaggerignore:"true"`
	ID                   uuid.UUID                `json:"id" bun:"id,pk,type:uuid"`
	Name                 string                   `json:"name" bun:"name,notnull"`
	Port                 int                      `json:"port" bun:"port,notnull"`
	Environment          Environment              `json:"environment" bun:"environment,notnull"`
	ProxyServer          ProxyServer              `json:"proxy_server" bun:"proxy_server,notnull,default:caddy"`
	BuildVariables       string                   `json:"build_variables" bun:"build_variables,notnull"`
	EnvironmentVariables string                   `json:"environment_variables" bun:"environment_variables,notnull"`
	BuildPack            BuildPack                `json:"build_pack" bun:"build_pack,notnull"`
	Repository           string                   `json:"repository" bun:"repository,notnull"`
	Branch               string                   `json:"branch" bun:"branch,notnull"`
	PreRunCommand        string                   `json:"pre_run_command" bun:"pre_run_command,notnull"`
	PostRunCommand       string                   `json:"post_run_command" bun:"post_run_command,notnull"`
	DockerfilePath       string                   `json:"dockerfile_path" bun:"dockerfile_path,notnull,default:Dockerfile"`
	BasePath             string                   `json:"base_path" bun:"base_path,notnull,default:/"`
	UserID               uuid.UUID                `json:"user_id" bun:"user_id,notnull,type:uuid"`
	OrganizationID       uuid.UUID                `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	FamilyID             *uuid.UUID               `json:"family_id,omitempty" bun:"family_id,type:uuid"`
	CreatedAt            time.Time                `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt            time.Time                `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	User                 *User                    `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Status               *ApplicationStatus       `json:"status,omitempty" bun:"rel:has-one,join:id=application_id"`
	Logs                 []*ApplicationLogs       `json:"logs,omitempty" bun:"rel:has-many,join:id=application_id"`
	Deployments          []*ApplicationDeployment `json:"deployments,omitempty" bun:"rel:has-many,join:id=application_id"`
	Organization         *Organization            `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
	Labels               []string                 `json:"labels,omitempty" bun:"labels,array"`
	Domains              []*ApplicationDomain     `json:"domains,omitempty" bun:"rel:has-many,join:id=application_id"`
	ComposeServices      []*ComposeService        `json:"compose_services,omitempty" bun:"rel:has-many,join:id=application_id"`
	IsLiveDeployment     bool                     `json:"is_live_deployment" bun:"is_live_deployment,notnull,default:false"`
	Source               Source                   `json:"source" bun:"source,notnull,default:'github'"`
	RoutingStrategy      RoutingStrategy          `json:"routing_strategy" bun:"routing_strategy,notnull,default:'single'"`
	Servers              []*ApplicationServer     `json:"servers,omitempty" bun:"rel:has-many,join:id=application_id"`
}

type ApplicationDeployment struct {
	bun.BaseModel   `bun:"table:application_deployment,alias:ad" swaggerignore:"true"`
	ID              uuid.UUID                    `json:"id" bun:"id,pk,type:uuid"`
	ApplicationID   uuid.UUID                    `json:"application_id" bun:"application_id,notnull,type:uuid"`
	CreatedAt       time.Time                    `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt       time.Time                    `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	CommitHash      string                       `json:"commit_hash" bun:"commit_hash"`
	Application     *Application                 `json:"application,omitempty" bun:"rel:belongs-to,join:application_id=id"`
	Status          *ApplicationDeploymentStatus `json:"status,omitempty" bun:"rel:has-one,join:id=application_deployment_id"`
	Logs            []*ApplicationLogs           `json:"logs,omitempty" bun:"rel:has-many,join:id=application_deployment_id"`
	ContainerID     string                       `json:"container_id" bun:"container_id"`
	ContainerName   string                       `json:"container_name" bun:"container_name"`
	ContainerImage  string                       `json:"container_image" bun:"container_image"`
	ContainerStatus string                       `json:"container_status" bun:"container_status"`
	ImageS3Key      string                       `json:"image_s3_key" bun:"image_s3_key,default:''"`
	ImageSize       int64                        `json:"image_size" bun:"image_size,default:0"`
	ServerID           *uuid.UUID              `json:"server_id,omitempty"            bun:"server_id,type:uuid"`
	ParentDeploymentID *uuid.UUID              `json:"parent_deployment_id,omitempty" bun:"parent_deployment_id,type:uuid"`
	Children           []*ApplicationDeployment `json:"children,omitempty"            bun:"rel:has-many,join:id=parent_deployment_id"`
}

type ApplicationStatus struct {
	bun.BaseModel `bun:"table:application_status,alias:as" swaggerignore:"true"`
	ID            uuid.UUID `json:"id" bun:"id,pk,type:uuid"`
	ApplicationID uuid.UUID `json:"application_id" bun:"application_id,notnull,type:uuid"`
	Status        Status    `json:"status" bun:"status,notnull"`
	CreatedAt     time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`

	Application *Application `json:"application,omitempty" bun:"rel:belongs-to,join:application_id=id"`
}

type ApplicationDeploymentStatus struct {
	bun.BaseModel           `bun:"table:application_deployment_status,alias:ads" swaggerignore:"true"`
	ID                      uuid.UUID `json:"id" bun:"id,pk,type:uuid"`
	ApplicationDeploymentID uuid.UUID `json:"application_deployment_id" bun:"application_deployment_id,notnull,type:uuid"`
	Status                  Status    `json:"status" bun:"status,notnull"`
	CreatedAt               time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt               time.Time `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`

	ApplicationDeployment *ApplicationDeployment `json:"application_deployment,omitempty" bun:"rel:belongs-to,join:application_deployment_id=id"`
}

type ApplicationLogs struct {
	bun.BaseModel           `bun:"table:application_logs,alias:al" swaggerignore:"true"`
	ID                      uuid.UUID `json:"id" bun:"id,pk,type:uuid"`
	ApplicationID           uuid.UUID `json:"application_id" bun:"application_id,notnull,type:uuid"`
	CreatedAt               time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt               time.Time `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	Log                     string    `json:"log" bun:"log,notnull"`
	ApplicationDeploymentID uuid.UUID `json:"application_deployment_id" bun:"application_deployment_id,notnull,type:uuid"`

	ApplicationDeployment *ApplicationDeployment `json:"application_deployment,omitempty" bun:"rel:belongs-to,join:application_deployment_id=id"`
	Application           *Application           `json:"application,omitempty" bun:"rel:belongs-to,join:application_id=id"`
}

type ApplicationServer struct {
	bun.BaseModel `bun:"table:application_servers,alias:as" swaggerignore:"true"`
	ID            uuid.UUID `json:"id"             bun:"id,pk,type:uuid"`
	ApplicationID uuid.UUID `json:"application_id" bun:"application_id,notnull,type:uuid"`
	ServerID      uuid.UUID `json:"server_id"      bun:"server_id,notnull,type:uuid"`
	IsPrimary     bool      `json:"is_primary"     bun:"is_primary,notnull,default:false"`
	CreatedAt     time.Time `json:"created_at"     bun:"created_at,notnull,default:current_timestamp"`
	Server        *SSHKey   `json:"server,omitempty" bun:"rel:belongs-to,join:server_id=id"`
}

type ComposeService struct {
	bun.BaseModel `bun:"table:compose_services,alias:cs" swaggerignore:"true"`
	ID            uuid.UUID `json:"id" bun:"id,pk,type:uuid"`
	ApplicationID uuid.UUID `json:"application_id" bun:"application_id,notnull,type:uuid"`
	ServiceName   string    `json:"service_name" bun:"service_name,notnull"`
	Port          int       `json:"port" bun:"port,notnull"`
	CreatedAt     time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`

	Application *Application         `json:"application,omitempty" bun:"rel:belongs-to,join:application_id=id"`
	Domains     []*ApplicationDomain `json:"domains,omitempty" bun:"rel:has-many,join:id=compose_service_id"`
}

type ApplicationDomain struct {
	bun.BaseModel    `bun:"table:application_domains,alias:ad" swaggerignore:"true"`
	ID               uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	ApplicationID    uuid.UUID  `json:"application_id" bun:"application_id,notnull,type:uuid"`
	Domain           string     `json:"domain" bun:"domain,notnull"`
	ComposeServiceID *uuid.UUID `json:"compose_service_id,omitempty" bun:"compose_service_id,type:uuid"`
	Port             *int       `json:"port,omitempty" bun:"port"`
	CreatedAt        time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`

	Application    *Application    `json:"application,omitempty" bun:"rel:belongs-to,join:application_id=id"`
	ComposeService *ComposeService `json:"compose_service,omitempty" bun:"rel:belongs-to,join:compose_service_id=id"`
}

// ResolvePort returns the upstream port for this domain based on its linked
// ComposeService or explicit port override. Returns 0 when the domain is
// orphaned (service removed, no manual override) and should not be routed.
func (d *ApplicationDomain) ResolvePort() int {
	if d.ComposeService != nil && d.ComposeService.Port > 0 {
		return d.ComposeService.Port
	}
	if d.Port != nil && *d.Port > 0 {
		return *d.Port
	}
	return 0
}

type RoutingStrategy string

const (
	RoutingStrategySingle          RoutingStrategy = "single"
	RoutingStrategyRoundRobin      RoutingStrategy = "round_robin"
	RoutingStrategyPrimaryFailover RoutingStrategy = "primary_failover"
	RoutingStrategyPerServerDomain RoutingStrategy = "per_server_domain"
)

type Status string

const (
	Draft          Status = "draft"
	Started        Status = "started"
	Running        Status = "running"
	Stopped        Status = "stopped"
	Failed         Status = "failed"
	Cloning        Status = "cloning"
	Building       Status = "building"
	Deploying      Status = "deploying"
	Deployed       Status = "deployed"
	Cancelled      Status = "cancelled"
	PartialFailure Status = "partial_failure"
)

type Environment string

var validEnvironmentRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func IsValidEnvironment(env string) bool {
	return len(env) >= 1 && len(env) <= 50 && validEnvironmentRegex.MatchString(env)
}

type ProxyServer string

const (
	Nginx ProxyServer = "nginx"
	Caddy ProxyServer = "caddy"
)

type BuildPack string

const (
	DockerFile    BuildPack = "dockerfile"
	DockerCompose BuildPack = "docker-compose"
	Static        BuildPack = "static"
)

func IsValidBuildPack(bp string) bool {
	switch BuildPack(bp) {
	case DockerFile, DockerCompose, Static:
		return true
	}
	return false
}

type Source string

const (
	SourceGithub  Source = "github"
	SourceS3      Source = "s3"
	SourceZip     Source = "zip"
	SourceStaging Source = "staging"
)

type DeploymentRequestConfig struct {
	Type              DeploymentType `json:"type"`
	Force             bool           `json:"force"`
	ForceWithoutCache bool           `json:"force_without_cache"`
}

type DeploymentType string

const (
	DeploymentTypeCreate   = "create"
	DeploymentTypeUpdate   = "update"
	DeploymentTypeReDeploy = "redeploy"
	DeploymentTypeRollback = "rollback"
	DeploymentTypeRestart  = "restart"
)

type WebhookPayload struct {
	Repository struct {
		ID       uint64 `json:"id"`
		FullName string `json:"full_name"`
	} `json:"repository"`
	Ref    string `json:"ref"`
	Before string `json:"before"`
	After  string `json:"after"`
	Pusher struct {
		Name string `json:"name"`
	} `json:"pusher"`
}
