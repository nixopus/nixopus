package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// UserPreferencesData represents the JSONB data for user preferences
type UserPreferencesData struct {
	DebugMode             bool     `json:"debug_mode"`
	ShowApiErrorDetails   bool     `json:"show_api_error_details"`
	TerminalScrollback    *int     `json:"terminal_scrollback,omitempty"`
	TerminalFontSize      *int     `json:"terminal_font_size,omitempty"`
	TerminalCursorStyle   *string  `json:"terminal_cursor_style,omitempty"`
	TerminalCursorBlink   *bool    `json:"terminal_cursor_blink,omitempty"`
	TerminalLineHeight    *float64 `json:"terminal_line_height,omitempty"`
	TerminalCursorWidth   *int     `json:"terminal_cursor_width,omitempty"`
	TerminalTabStopWidth  *int     `json:"terminal_tab_stop_width,omitempty"`
	TerminalFontFamily    *string  `json:"terminal_font_family,omitempty"`
	TerminalFontWeight    *string  `json:"terminal_font_weight,omitempty"`
	TerminalLetterSpacing *float64 `json:"terminal_letter_spacing,omitempty"`
}

// UserPreferences represents personal user preferences stored in the database
type UserPreferences struct {
	bun.BaseModel `bun:"table:user_preferences,alias:up" swaggerignore:"true"`
	ID            uuid.UUID           `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID        uuid.UUID           `json:"user_id" bun:"user_id,notnull,type:uuid"`
	Preferences   UserPreferencesData `json:"preferences" bun:"preferences,type:jsonb,notnull"`
	CreatedAt     time.Time           `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time           `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`

	User *User `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
}

// OrganizationSettingsData represents the JSONB data for organization settings
type OrganizationSettingsData struct {
	WebsocketReconnectAttempts       int     `json:"websocket_reconnect_attempts"`
	WebsocketReconnectInterval       int     `json:"websocket_reconnect_interval"`
	ApiRetryAttempts                 int     `json:"api_retry_attempts"`
	DisableApiCache                  bool    `json:"disable_api_cache"`
	ContainerLogTailLines            *int    `json:"container_log_tail_lines,omitempty"`
	ContainerDefaultRestartPolicy    *string `json:"container_default_restart_policy,omitempty"`
	ContainerStopTimeout             *int    `json:"container_stop_timeout,omitempty"`
	ContainerAutoPruneDanglingImages *bool   `json:"container_auto_prune_dangling_images,omitempty"`
	ContainerAutoPruneBuildCache     *bool   `json:"container_auto_prune_build_cache,omitempty"`
}

// OrganizationSettings represents organization-level settings stored in the database
type OrganizationSettings struct {
	bun.BaseModel  `bun:"table:organization_settings,alias:os" swaggerignore:"true"`
	ID             uuid.UUID                `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID uuid.UUID                `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	Settings       OrganizationSettingsData `json:"settings" bun:"settings,type:jsonb,notnull"`
	CreatedAt      time.Time                `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time                `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`

	Organization *Organization `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
}

// DefaultUserPreferencesData returns default values for user preferences
func DefaultUserPreferencesData() UserPreferencesData {
	scrollback := 5000
	fontSize := 13
	cursorStyle := "bar"
	cursorBlink := true
	lineHeight := 1.4
	cursorWidth := 2
	tabStopWidth := 4
	fontFamily := "JetBrains Mono"
	fontWeight := "normal"
	letterSpacing := 0.0

	return UserPreferencesData{
		DebugMode:             false,
		ShowApiErrorDetails:   false,
		TerminalScrollback:    &scrollback,
		TerminalFontSize:      &fontSize,
		TerminalCursorStyle:   &cursorStyle,
		TerminalCursorBlink:   &cursorBlink,
		TerminalLineHeight:    &lineHeight,
		TerminalCursorWidth:   &cursorWidth,
		TerminalTabStopWidth:  &tabStopWidth,
		TerminalFontFamily:    &fontFamily,
		TerminalFontWeight:    &fontWeight,
		TerminalLetterSpacing: &letterSpacing,
	}
}

// DefaultOrganizationSettingsData returns default values for organization settings
func DefaultOrganizationSettingsData() OrganizationSettingsData {
	logTailLines := 100
	restartPolicy := "unless-stopped"
	stopTimeout := 10
	autoPruneImages := false
	autoPruneCache := false

	return OrganizationSettingsData{
		WebsocketReconnectAttempts:       5,
		WebsocketReconnectInterval:       3000,
		ApiRetryAttempts:                 1,
		DisableApiCache:                  false,
		ContainerLogTailLines:            &logTailLines,
		ContainerDefaultRestartPolicy:    &restartPolicy,
		ContainerStopTimeout:             &stopTimeout,
		ContainerAutoPruneDanglingImages: &autoPruneImages,
		ContainerAutoPruneBuildCache:     &autoPruneCache,
	}
}
