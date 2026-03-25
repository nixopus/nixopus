package caddy

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/queue"
)

// Extension-managed domains are tracked in a Redis hash per org:
//   caddy:ext_domains:<org_id>  →  { domain: "host:port", ... }
// This lets the reconciler re-add extension domains after a Caddy reset
// without requiring a new DB table.

const extDomainsKeyPrefix = "caddy:ext_domains:"

func extDomainsKey(orgID uuid.UUID) string {
	return extDomainsKeyPrefix + orgID.String()
}

// TrackExtensionDomain records an extension-managed domain→upstream mapping
// so the reconciler can restore it if Caddy loses its config.
func TrackExtensionDomain(orgID uuid.UUID, domain, upstreamDial string) error {
	rc := queue.RedisClient()
	if rc == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return rc.HSet(context.Background(), extDomainsKey(orgID), domain, upstreamDial).Err()
}

// UntrackExtensionDomain removes an extension-managed domain from tracking.
func UntrackExtensionDomain(orgID uuid.UUID, domain string) error {
	rc := queue.RedisClient()
	if rc == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return rc.HDel(context.Background(), extDomainsKey(orgID), domain).Err()
}

// GetExtensionDomains returns all extension-managed domain→upstream mappings for an org.
func GetExtensionDomains(ctx context.Context, orgID uuid.UUID) ([]DomainRoute, error) {
	rc := queue.RedisClient()
	if rc == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	entries, err := rc.HGetAll(ctx, extDomainsKey(orgID)).Result()
	if err != nil {
		return nil, err
	}

	routes := make([]DomainRoute, 0, len(entries))
	for domain, dial := range entries {
		routes = append(routes, DomainRoute{Domain: domain, UpstreamDial: dial})
	}
	return routes, nil
}
