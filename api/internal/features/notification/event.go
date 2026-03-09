package notification

import shared_types "github.com/nixopus/nixopus/api/internal/types"

// preferenceMapping maps event types to the preference category + type
// that controls whether the user wants this notification.
type preferenceMapping struct {
	Category string
	Type     string
}

var eventPreferenceMap = map[shared_types.EventType]preferenceMapping{
	shared_types.EventLoginAlert:          {Category: "security", Type: "login-alerts"},
	shared_types.EventPasswordReset:       {Category: "security", Type: "password-changes"},
	shared_types.EventVerificationEmail:   {Category: "security", Type: "security-alerts"},
	shared_types.EventUserAddedToOrg:      {Category: "activity", Type: "team-updates"},
	shared_types.EventUserRemovedFromOrg:  {Category: "activity", Type: "team-updates"},
	shared_types.EventDeploySuccess:       {Category: "activity", Type: "team-updates"},
	shared_types.EventDeployFailed:        {Category: "activity", Type: "team-updates"},
	shared_types.EventBuildFailed:         {Category: "activity", Type: "team-updates"},
	shared_types.EventHealthCheckCritical: {Category: "activity", Type: "team-updates"},
}

// eventTemplate maps event types to the email template and subject to use.
type eventTemplate struct {
	Subject  string
	Template string
}

var eventTemplateMap = map[shared_types.EventType]eventTemplate{
	shared_types.EventLoginAlert:          {Subject: "Login Notification", Template: "login_notification.html"},
	shared_types.EventPasswordReset:       {Subject: "Password Reset Request", Template: "password_reset.html"},
	shared_types.EventVerificationEmail:   {Subject: "Verification Email", Template: "verification_email.html"},
	shared_types.EventUserAddedToOrg:      {Subject: "New User Added to Organization", Template: "add_user_to_organization.html"},
	shared_types.EventUserRemovedFromOrg:  {Subject: "User Removed from Organization", Template: "remove_user_from_organization.html"},
	shared_types.EventBuildFailed:         {Subject: "Build Failed", Template: "build_failed.html"},
	shared_types.EventHealthCheckCritical: {Subject: "Health Check Critical", Template: "healthcheck_critical.html"},
}

// skipPreferenceCheck lists event types where we always send regardless
// of user preferences (e.g. password reset must always go through).
var skipPreferenceCheck = map[shared_types.EventType]bool{
	shared_types.EventPasswordReset:       true,
	shared_types.EventVerificationEmail:   true,
	shared_types.EventHealthCheckCritical: true,
}

// defaultChannelsForEvent returns the channels to use when the event
// doesn't specify explicit overrides.
func defaultChannelsForEvent(eventType shared_types.EventType) []string {
	switch eventType {
	case shared_types.EventLoginAlert, shared_types.EventPasswordReset, shared_types.EventVerificationEmail:
		return []string{"email"}
	case shared_types.EventUserAddedToOrg, shared_types.EventUserRemovedFromOrg:
		return []string{"email", "slack", "discord"}
	case shared_types.EventDeploySuccess:
		return []string{"slack", "discord"}
	case shared_types.EventDeployFailed:
		return []string{"slack", "discord", "agent"}
	case shared_types.EventBuildFailed, shared_types.EventHealthCheckCritical:
		return []string{"email", "slack", "discord", "agent"}
	default:
		return []string{"email"}
	}
}
