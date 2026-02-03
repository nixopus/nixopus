package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// DockerConfig represents Docker configuration for an organization (database model)
type DockerConfig struct {
	bun.BaseModel `bun:"table:docker_configs,alias:dc" swaggerignore:"true"`

	ID             uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID uuid.UUID `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	Name           string    `json:"name" bun:"name,notnull"`
	Description    *string   `json:"description,omitempty" bun:"description"`
	Host           string    `json:"host" bun:"host,notnull"` // e.g., "tcp://docker.example.com:2376"
	Port           *int      `json:"port,omitempty" bun:"port,default:2376"`

	// TLS Configuration
	TLSEnabled          bool    `json:"tls_enabled" bun:"tls_enabled,notnull,default:false"`
	CACertEncrypted     *string `json:"-" bun:"ca_cert_encrypted"`     // Encrypted CA certificate
	ClientCertEncrypted *string `json:"-" bun:"client_cert_encrypted"` // Encrypted client certificate
	ClientKeyEncrypted  *string `json:"-" bun:"client_key_encrypted"`  // Encrypted client private key

	// Metadata
	IsActive   bool       `json:"is_active" bun:"is_active,notnull,default:true"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" bun:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at" bun:"created_at,notnull,default:now()"`
	UpdatedAt  time.Time  `json:"updated_at" bun:"updated_at,notnull,default:now()"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	Organization *Organization `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
}

// DockerConfigRequest represents Docker config creation/update request
type DockerConfigRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	Host        string  `json:"host" validate:"required"`
	Port        *int    `json:"port,omitempty"`
	TLSEnabled  bool    `json:"tls_enabled"`

	// TLS certificates (plaintext in request, will be encrypted before storage)
	CACert     *string `json:"ca_cert,omitempty"`
	ClientCert *string `json:"client_cert,omitempty"`
	ClientKey  *string `json:"client_key,omitempty"`
}

// DockerRuntimeConfig represents runtime Docker configuration (decrypted)
// This is used by DockerManager to create Docker clients
type DockerRuntimeConfig struct {
	Host       string `mapstructure:"host" validate:"required"`
	Port       int    `mapstructure:"port"`
	TLSEnabled bool   `mapstructure:"tls_enabled"`
	CACert     string `mapstructure:"ca_cert"`
	ClientCert string `mapstructure:"client_cert"`
	ClientKey  string `mapstructure:"client_key"`
}
