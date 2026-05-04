package controller

import (
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/telemetry/types"
)

// trackInstallRequestData builds log context without user-supplied error text (paths, secrets).
func trackInstallRequestData(req *types.TrackInstallRequest) string {
	if req == nil {
		return ""
	}
	return fmt.Sprintf("event_type=%s os=%s arch=%s version=%s duration=%d error_len=%d",
		req.EventType, req.OS, req.Arch, req.Version, req.Duration, len(req.Error))
}
