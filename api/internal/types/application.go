package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Application struct {
	bun.BaseModel        `bun:"table:applications,alias:a" swaggerignore:"true"`
	ID                   uuid.UUID   `json:"id" bun:"id,pk,type:uuid"`
	Name                 string      `json:"name" bun:"name,notnull"`
	Port                 int         `json:"port" bun:"port,notnull"`
	Environment          Environment `json:"environment" bun:"environment,notnull"`
	BuildVariables       string      `json:"build_variables" bun:"build_variables,notnull"`
	EnvironmentVariables string      `json:"environment_variables" bun:"environment_variables,notnull"`
	BuildPack            BuildPack   `json:"build_pack" bun:"build_pack,notnull"`
	Repository           string      `json:"repository" bun:"repository,notnull"`
	Branch               string      `json:"branch" bun:"branch,notnull"`
	PreRunCommand        string      `json:"pre_run_command" bun:"pre_run_command,notnull"`
	PostRunCommand       string      `json:"post_run_command" bun:"post_run_command,notnull"`
	DomainID             string      `json:"domain_id" bun:"domain_id,notnull"`
	UserID               string      `json:"user_id" bun:"user_id,notnull"`
	CreatedAt            time.Time   `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt            time.Time   `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`

	Domain       *Domain             `json:"domain,omitempty" bun:"rel:belongs-to,join:domain_id=id"`
	User         *User               `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Status       *ApplicationStatus  `json:"status,omitempty" bun:"rel:has-one,join:id=application_id"`
	Logs         []*ApplicationLogs  `json:"logs,omitempty" bun:"rel:has-many,join:id=application_id"`
}


type ApplicationStatus struct {
	bun.BaseModel `bun:"table:application_status,alias:as" swaggerignore:"true"`
	ID            uuid.UUID    `json:"id" bun:"id,pk,type:uuid"`
	ApplicationID uuid.UUID    `json:"application_id" bun:"application_id,notnull,type:uuid"`
	Status        Status       `json:"status" bun:"status,notnull"`
	CreatedAt     time.Time    `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time    `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	
	Application   *Application `json:"-" bun:"rel:belongs-to,join:application_id=id"`
}

type ApplicationLogs struct {
	bun.BaseModel `bun:"table:application_logs,alias:al" swaggerignore:"true"`
	ID            uuid.UUID    `json:"id" bun:"id,pk,type:uuid"`
	ApplicationID uuid.UUID    `json:"application_id" bun:"application_id,notnull,type:uuid"`
	CreatedAt     time.Time    `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time    `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	Log           string      `json:"log" bun:"log,notnull"`
	
	Application   *Application `json:"-" bun:"rel:belongs-to,join:application_id=id"`
}

type Status string

const (
	Started Status = "started"
	Running Status = "running"
	Stopped Status = "stopped"
	Failed  Status = "failed"
)

type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

type BuildPack string

const (
	DockerFile    BuildPack = "dockerfile"
	DockerCompose BuildPack = "docker-compose"
	Static        BuildPack = "static"
)
