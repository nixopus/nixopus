package tasks

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/deploy/caddy"
	s3store "github.com/nixopus/nixopus/api/internal/features/deploy/s3"
	deploy_types "github.com/nixopus/nixopus/api/internal/features/deploy/types"
	sshpkg "github.com/nixopus/nixopus/api/internal/features/ssh"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// RecoverApplications restores applications whose containers have been lost
// (e.g. after a VPS crash) by loading their images from S3 and re-creating
// the Docker Swarm services. If applicationID is non-nil, only that single
// application is recovered; otherwise all deployed applications for the
// organization are processed.
func (s *TaskService) RecoverApplications(ctx context.Context, organizationID uuid.UUID, applicationID *uuid.UUID) (*deploy_types.RecoverResult, error) {
	if !s3store.IsConfigured(config.AppConfig.S3) {
		return nil, deploy_types.ErrS3NotConfigured
	}

	var apps []shared_types.Application
	if applicationID != nil {
		app, err := s.Storage.GetApplicationById(applicationID.String(), organizationID)
		if err != nil {
			return nil, fmt.Errorf("application not found: %w", err)
		}
		apps = []shared_types.Application{app}
	} else {
		var err error
		apps, err = s.Storage.GetDeployedApplications(organizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch deployed applications: %w", err)
		}
	}

	result := &deploy_types.RecoverResult{}
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, organizationID.String())

	for _, app := range apps {
		appResult := deploy_types.RecoverAppResult{
			ApplicationID:   app.ID,
			ApplicationName: app.Name,
		}

		existing, _ := FindServiceByName(orgCtx, app.Name)
		if existing != nil {
			appResult.Reason = "service already running"
			result.Skipped = append(result.Skipped, appResult)
			continue
		}

		deployment, err := s.Storage.GetLatestS3Deployment(app.ID)
		if err != nil {
			appResult.Reason = "failed to query S3 deployment: " + err.Error()
			result.Failed = append(result.Failed, appResult)
			continue
		}
		if deployment == nil {
			appResult.Reason = "no S3 image available"
			result.Skipped = append(result.Skipped, appResult)
			continue
		}

		err = s.recoverSingleApp(orgCtx, &app, deployment)
		if err != nil {
			appResult.Reason = err.Error()
			result.Failed = append(result.Failed, appResult)
			continue
		}

		appResult.Reason = "recovered from deployment " + deployment.ID.String()
		result.Recovered = append(result.Recovered, appResult)
	}

	return result, nil
}

// recoverSingleApp loads an image from S3, tags it, creates a new deployment
// record, starts the Swarm service, and configures proxy domains.
func (s *TaskService) recoverSingleApp(ctx context.Context, app *shared_types.Application, srcDeployment *shared_types.ApplicationDeployment) error {
	newDeployment := shared_types.ApplicationDeployment{
		ID:            uuid.New(),
		ApplicationID: app.ID,
		CommitHash:    srcDeployment.CommitHash,
		ImageS3Key:    srcDeployment.ImageS3Key,
		ImageSize:     srcDeployment.ImageSize,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := s.Storage.AddApplicationDeployment(&newDeployment); err != nil {
		return fmt.Errorf("failed to create deployment record: %w", err)
	}

	newStatus := shared_types.ApplicationDeploymentStatus{
		ID:                      uuid.New(),
		ApplicationDeploymentID: newDeployment.ID,
		Status:                  shared_types.Deploying,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}
	if err := s.Storage.AddApplicationDeploymentStatus(&newStatus); err != nil {
		return fmt.Errorf("failed to create deployment status: %w", err)
	}

	payload := shared_types.TaskPayload{
		CorrelationID:         uuid.NewString(),
		Application:           *app,
		ApplicationDeployment: newDeployment,
		Status:                &newStatus,
	}
	taskCtx := s.NewTaskContext(payload)

	taskCtx.LogAndUpdateStatus("Starting recovery from S3", shared_types.Deploying)

	if err := s.LoadImageFromS3(ctx, srcDeployment.ImageS3Key, taskCtx); err != nil {
		taskCtx.LogAndUpdateStatus("Failed to load image from S3: "+err.Error(), shared_types.Failed)
		return fmt.Errorf("failed to load image from S3: %w", err)
	}

	commitTag := CommitImageTag(app.Name, srcDeployment.CommitHash)
	latestTag := fmt.Sprintf("%s:latest", app.Name)

	sshManager, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get SSH manager: "+err.Error(), shared_types.Failed)
		return fmt.Errorf("failed to get SSH manager: %w", err)
	}

	clientConn, release, err := sshManager.Borrow("")
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to connect via SSH: "+err.Error(), shared_types.Failed)
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer release()

	session, err := clientConn.NewSession()
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to create SSH session: "+err.Error(), shared_types.Failed)
		return fmt.Errorf("SSH session failed: %w", err)
	}

	tagCmd := fmt.Sprintf("docker tag %s %s", commitTag, latestTag)
	output, err := session.CombinedOutput(tagCmd)
	session.Close()

	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to tag image: "+err.Error(), shared_types.Failed)
		return fmt.Errorf("docker tag failed: %w (output: %s)", err, string(output))
	}
	taskCtx.AddLog("Image tagged as " + latestTag)

	containerResult, err := s.AtomicUpdateContainer(ctx, payload, taskCtx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to create service: "+err.Error(), shared_types.Failed)
		return fmt.Errorf("failed to create service: %w", err)
	}

	taskCtx.AddLog("Service created with container id " + containerResult.ContainerID)

	swarmPort, _ := strconv.Atoi(containerResult.AvailablePort)
	s.configureRecoveryDomains(ctx, app, swarmPort, taskCtx)

	taskCtx.LogAndUpdateStatus("Recovery completed successfully", shared_types.Deployed)
	return nil
}

func (s *TaskService) configureRecoveryDomains(ctx context.Context, app *shared_types.Application, swarmPort int, taskCtx *TaskContext) {
	domains, err := s.Storage.GetApplicationDomains(app.ID)
	if err != nil {
		taskCtx.AddLog("Warning: failed to load domains for proxy setup: " + err.Error())
		return
	}
	if len(domains) == 0 {
		return
	}

	upstreamHost, err := GetSSHHostForOrganization(ctx, app.OrganizationID)
	if err != nil {
		taskCtx.AddLog("Warning: failed to get SSH host for domain setup: " + err.Error())
		return
	}

	isCompose := app.BuildPack == shared_types.DockerCompose
	var routes []caddy.DomainRoute
	for i := range domains {
		d := &domains[i]
		if d.Domain == "" {
			continue
		}

		port := swarmPort
		if isCompose {
			port = d.ResolvePort()
			if port == 0 {
				taskCtx.AddLog(fmt.Sprintf("Skipping orphaned compose domain %s during recovery", d.Domain))
				continue
			}
		}

		routes = append(routes, caddy.DomainRoute{
			Domain:       d.Domain,
			UpstreamDial: caddy.FormatDial(upstreamHost, port),
		})
	}
	if len(routes) == 0 {
		return
	}

	if err := caddy.AddDomainsWithRetry(ctx, nil, &s.Logger, routes); err != nil {
		taskCtx.AddLog("Warning: failed to configure proxy: " + err.Error())
	}
}
