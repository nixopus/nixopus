package caddy

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/queue"
)

const pendingRemovalKeyPrefix = "caddy:pending_removals:"

func pendingRemovalKey(orgID uuid.UUID) string {
	return pendingRemovalKeyPrefix + orgID.String()
}

// EnqueuePendingRemoval marks a domain for deferred removal from Caddy.
// Called when an explicit Caddy delete fails so the reconciler can retry later.
func EnqueuePendingRemoval(orgID uuid.UUID, domains ...string) error {
	rc := queue.RedisClient()
	if rc == nil {
		return fmt.Errorf("redis client not initialized")
	}
	if len(domains) == 0 {
		return nil
	}

	members := make([]interface{}, len(domains))
	for i, d := range domains {
		members[i] = d
	}
	return rc.SAdd(context.Background(), pendingRemovalKey(orgID), members...).Err()
}

// GetPendingRemovals returns all domains queued for removal for an org.
func GetPendingRemovals(ctx context.Context, orgID uuid.UUID) ([]string, error) {
	rc := queue.RedisClient()
	if rc == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	return rc.SMembers(ctx, pendingRemovalKey(orgID)).Result()
}

// ClearPendingRemoval removes successfully deleted domains from the pending set.
func ClearPendingRemoval(orgID uuid.UUID, domains ...string) error {
	rc := queue.RedisClient()
	if rc == nil {
		return fmt.Errorf("redis client not initialized")
	}
	if len(domains) == 0 {
		return nil
	}

	members := make([]interface{}, len(domains))
	for i, d := range domains {
		members[i] = d
	}
	return rc.SRem(context.Background(), pendingRemovalKey(orgID), members...).Err()
}
