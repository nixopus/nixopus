package types

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CliInstallation struct {
	bun.BaseModel `bun:"table:cli_installations,alias:ci" swaggerignore:"true"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	EventType string    `bun:"event_type,notnull" json:"event_type"`
	OS        string    `bun:"os,notnull,default:'unknown'" json:"os"`
	Arch      string    `bun:"arch,notnull,default:'unknown'" json:"arch"`
	Version   string    `bun:"version,notnull" json:"version"`
	Duration  int       `bun:"duration,notnull,default:0" json:"duration"`
	Error     string    `bun:"error" json:"error,omitempty"`
	IPHash    string    `bun:"ip_hash,notnull" json:"-"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

type TrackInstallRequest struct {
	EventType string `json:"event_type"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Version   string `json:"version"`
	Duration  int    `json:"duration"`
	Error     string `json:"error,omitempty"`
}

type TrackInstallResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var (
	ErrInvalidEventType   = errors.New("event_type must be one of: install_started, install_success, install_failure")
	ErrInvalidOS          = errors.New("unrecognized operating system")
	ErrInvalidArch        = errors.New("arch must be amd64 or arm64")
	ErrInvalidVersion     = errors.New("invalid version format")
	ErrInvalidDuration    = errors.New("duration must be between 0 and 7200")
	ErrErrorTooLong       = errors.New("error must be 200 characters or less")
	ErrInvalidRequestType = errors.New("invalid request type")
)

var AllowedEventTypes = map[string]bool{
	"install_started": true,
	"install_success": true,
	"install_failure": true,
}

var AllowedOS = map[string]bool{
	"ubuntu": true, "debian": true, "centos": true, "fedora": true,
	"rhel": true, "alpine": true, "arch": true, "opensuse": true,
	"sles": true, "amzn": true, "ol": true, "rocky": true,
	"almalinux": true, "raspbian": true, "pop": true, "mint": true,
	"manjaro": true, "kali": true, "nixos": true, "gentoo": true,
	"void": true, "slackware": true, "clear-linux-os": true,
	"unknown": true,
}

var AllowedArch = map[string]bool{
	"amd64":   true,
	"arm64":   true,
	"unknown": true,
}
