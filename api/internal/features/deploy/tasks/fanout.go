package tasks

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// addChildDeployment creates and inserts a per-server child ApplicationDeployment.
func (t *TaskService) addChildDeployment(parentDep shared_types.ApplicationDeployment, serverID uuid.UUID) (shared_types.ApplicationDeployment, error) {
	now := time.Now()
	child := shared_types.ApplicationDeployment{
		ApplicationID:   parentDep.ApplicationID,
		CreatedAt:       now,
		UpdatedAt:       now,
		CommitHash:      parentDep.CommitHash,
		ImageS3Key:      parentDep.ImageS3Key,
		ContainerID:     parentDep.ContainerID,
		ContainerName:   parentDep.ContainerName,
		ContainerImage:  parentDep.ContainerImage,
		ContainerStatus: parentDep.ContainerStatus,
		ImageSize:       parentDep.ImageSize,
	}
	child.ID = uuid.New()
	child.ServerID = &serverID
	child.ParentDeploymentID = &parentDep.ID

	if err := t.Storage.AddApplicationDeployment(&child); err != nil {
		return shared_types.ApplicationDeployment{}, err
	}

	initialStatus := &shared_types.ApplicationDeploymentStatus{
		ID:                      uuid.New(),
		ApplicationDeploymentID: child.ID,
		Status:                  shared_types.Deploying,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if err := t.Storage.AddApplicationDeploymentStatus(initialStatus); err != nil {
		return child, err
	}
	child.Status = initialStatus
	return child, nil
}

// updateParentStatus inserts a final status record for the parent deployment based on child outcomes.
func (t *TaskService) updateParentStatus(ctx context.Context, parentDepID uuid.UUID, errs []error) {
	var failCount int
	for _, e := range errs {
		if e != nil {
			failCount++
		}
	}

	var status shared_types.Status
	switch {
	case failCount == 0:
		status = shared_types.Deployed
	case failCount == len(errs):
		status = shared_types.Failed
	default:
		status = shared_types.PartialFailure
	}

	now := time.Now()
	appStatus := &shared_types.ApplicationDeploymentStatus{
		ID:                      uuid.New(),
		ApplicationDeploymentID: parentDepID,
		Status:                  status,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if err := t.Storage.AddApplicationDeploymentStatus(appStatus); err != nil {
		t.Logger.Log(logger.Error, "failed to update parent deployment status", err.Error())
	}
}

// filterServers returns only the servers matching targetIDs. If targetIDs is empty, all servers are returned.
func filterServers(all []shared_types.ApplicationServer, targetIDs []uuid.UUID) []shared_types.ApplicationServer {
	if len(targetIDs) == 0 {
		return all
	}
	set := make(map[uuid.UUID]struct{}, len(targetIDs))
	for _, id := range targetIDs {
		set[id] = struct{}{}
	}
	var out []shared_types.ApplicationServer
	for _, s := range all {
		if _, ok := set[s.ServerID]; ok {
			out = append(out, s)
		}
	}
	return out
}

// fanOut runs fn for each server in parallel. Each goroutine gets a fresh child deployment.
// Returns errors.Join of all goroutine errors (nil if all succeed).
func (t *TaskService) fanOut(
	ctx context.Context,
	d shared_types.TaskPayload,
	servers []shared_types.ApplicationServer,
	fn func(ctx context.Context, d shared_types.TaskPayload) error,
) error {
	errs := make([]error, len(servers))
	var wg sync.WaitGroup
	for i, srv := range servers {
		wg.Add(1)
		go func(idx int, serverID uuid.UUID) {
			defer wg.Done()

			child, err := t.addChildDeployment(d.ApplicationDeployment, serverID)
			if err != nil {
				t.Logger.Log(logger.Error, "failed to create child deployment", err.Error())
				errs[idx] = err
				return
			}

			serverCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, d.Application.OrganizationID.String())
			serverCtx = context.WithValue(serverCtx, shared_types.ServerIDKey, serverID.String())

			serverPayload := d
			serverPayload.ApplicationDeployment = child

			errs[idx] = fn(serverCtx, serverPayload)
		}(i, srv.ServerID)
	}
	wg.Wait()
	t.updateParentStatus(ctx, d.ApplicationDeployment.ID, errs)
	return errors.Join(errs...)
}
