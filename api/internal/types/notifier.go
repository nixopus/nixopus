package types

// EventType identifies what happened in the system that triggers a notification.
type EventType string

const (
	EventDeploySuccess       EventType = "deploy.success"
	EventDeployFailed        EventType = "deploy.failed"
	EventBuildFailed         EventType = "deploy.build_failed"
	EventContainerCrashed    EventType = "container.crashed"
	EventHealthCheckCritical EventType = "healthcheck.critical"
	EventLoginAlert          EventType = "auth.login"
	EventPasswordReset       EventType = "auth.password_reset"
	EventVerificationEmail   EventType = "auth.verification"
	EventUserAddedToOrg      EventType = "org.user_added"
	EventUserRemovedFromOrg  EventType = "org.user_removed"
	EventTrialExpired        EventType = "trail.trial_expired"
)

// NotificationEvent is the payload any service emits to trigger notifications.
// Channels is optional: when empty the dispatcher uses the default channels
// for the event type. When set, it overrides the defaults.
type NotificationEvent struct {
	Type           EventType              `json:"type"`
	UserID         string                 `json:"user_id"`
	OrganizationID string                 `json:"organization_id"`
	Data           map[string]interface{} `json:"data"`
	Channels       []string               `json:"channels,omitempty"`
}

// Notifier is a lightweight interface that services and controllers use to emit
// notification events. This avoids coupling every feature to the full
// notification package -- callers only need this 1-method interface.
//
// The concrete implementation is notification.Dispatcher.
type Notifier interface {
	Emit(event NotificationEvent) error
}
