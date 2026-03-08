package service

import (
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/trail/types"
	"github.com/raghavyuva/nixopus-api/internal/queue"
)

func (s *TrailService) UpgradeResources(userID, orgID string, vcpu, memoryMB int) error {
	provision, err := s.storage.GetCompletedProvisionByUserID(userID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to look up completed provision: %v", err), userID)
		return fmt.Errorf("failed to look up provision: %w", err)
	}

	if provision == nil {
		s.logger.Log(logger.Warning, "No completed provision found for resource upgrade", userID)
		return types.ErrProvisionNotFound
	}

	if provision.LXDContainerName == nil || *provision.LXDContainerName == "" {
		s.logger.Log(logger.Error, "Completed provision has no container name", userID)
		return fmt.Errorf("provision missing container name")
	}

	payload := queue.ResourceUpdatePayload{
		VMName:    *provision.LXDContainerName,
		VcpuCount: vcpu,
		MemoryMB:  memoryMB,
		UserID:    userID,
		OrgID:     orgID,
	}

	if err := queue.EnqueueResourceUpdateTask(s.ctx, payload); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to enqueue resource update: %v", err), userID)
		return types.ErrFailedToEnqueueTask
	}

	s.logger.Log(logger.Info, fmt.Sprintf("Resource upgrade enqueued: vm=%s vcpu=%d mem=%d", payload.VMName, vcpu, memoryMB), userID)
	return nil
}
