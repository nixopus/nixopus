package caddy

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/docker/docker/api/types/swarm"
	"github.com/google/uuid"
	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/queue"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/vmihailenco/taskq/v3"
)

// ReconcileResult tracks what the reconciler did during a single run.
type ReconcileResult struct {
	Added   []string
	Removed []string
	Updated []string
	Errors  []string
}

// Reconciler compares the DB-defined proxy state (source of truth) with what
// Caddy currently has configured, and corrects any drift.
type Reconciler struct {
	Storage storage.DeployRepository
	Logger  logger.Logger

	orgLocks   map[uuid.UUID]*sync.Mutex
	orgLocksMu sync.Mutex
}

func NewReconciler(store storage.DeployRepository, lgr logger.Logger) *Reconciler {
	return &Reconciler{
		Storage:  store,
		Logger:   lgr,
		orgLocks: make(map[uuid.UUID]*sync.Mutex),
	}
}

func (r *Reconciler) orgLock(orgID uuid.UUID) *sync.Mutex {
	r.orgLocksMu.Lock()
	defer r.orgLocksMu.Unlock()
	mu, exists := r.orgLocks[orgID]
	if !exists {
		mu = &sync.Mutex{}
		r.orgLocks[orgID] = mu
	}
	return mu
}

// ReconcileOrganization ensures all DB-defined domains for an organization
// exist in Caddy with the correct upstream. It is additive-only: it will
// add missing domains and fix mismatched upstreams, but will NEVER remove
// a domain just because the reconciler doesn't recognise it. Removal only
// happens for domains explicitly enqueued in the pending-removal set (via
// EnqueuePendingRemoval) after a real delete operation failed.
func (r *Reconciler) ReconcileOrganization(ctx context.Context, organizationID uuid.UUID) (*ReconcileResult, error) {
	mu := r.orgLock(organizationID)
	mu.Lock()
	defer mu.Unlock()

	result := &ReconcileResult{}

	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, organizationID.String())

	upstreamHost, err := getSSHHostForOrg(orgCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH host for org %s: %w", organizationID, err)
	}

	desired, err := r.buildDesiredState(orgCtx, organizationID, upstreamHost)
	if err != nil {
		return nil, fmt.Errorf("failed to build desired state: %w", err)
	}

	actual, err := GetCurrentDomains(orgCtx, nil, &r.Logger)
	if err != nil {
		r.Logger.Log(logger.Warning, "failed to read caddy state, attempting full rebuild", err.Error())
		return r.fullRebuild(orgCtx, desired)
	}

	actualMap := make(map[string]string)
	for _, route := range actual {
		actualMap[route.Domain] = route.UpstreamDial
	}

	var toAdd []DomainRoute
	var toUpdate []DomainRoute

	for _, route := range desired {
		actualDial, exists := actualMap[route.Domain]
		if !exists {
			toAdd = append(toAdd, route)
		} else if actualDial != route.UpstreamDial {
			toUpdate = append(toUpdate, route)
		}
	}

	// Process pending removals — these are domains that were explicitly
	// deleted from the DB but whose Caddy removal failed at delete time.
	pendingRemovals, pendingErr := GetPendingRemovals(orgCtx, organizationID)
	if pendingErr != nil {
		r.Logger.Log(logger.Warning, "failed to read pending removals", pendingErr.Error())
	}

	needsReload := len(toAdd) > 0 || len(toUpdate) > 0 || len(pendingRemovals) > 0

	if needsReload {
		client, err := GetCaddyClient(orgCtx, nil, &r.Logger)
		if err != nil {
			return nil, fmt.Errorf("failed to get caddy client for reconciliation: %w", err)
		}

		for _, route := range toAdd {
			host, port, parseErr := parseDial(route.UpstreamDial)
			if parseErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("invalid dial for %s: %v", route.Domain, parseErr))
				continue
			}
			if err := client.AddDomainWithAutoTLS(route.Domain, host, port, caddygo.DomainOptions{}); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to add %s: %v", route.Domain, err))
			} else {
				result.Added = append(result.Added, route.Domain)
			}
		}

		for _, route := range toUpdate {
			host, port, parseErr := parseDial(route.UpstreamDial)
			if parseErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("invalid dial for %s: %v", route.Domain, parseErr))
				continue
			}
			if err := client.AddDomainWithAutoTLS(route.Domain, host, port, caddygo.DomainOptions{}); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to update %s: %v", route.Domain, err))
			} else {
				result.Updated = append(result.Updated, route.Domain)
			}
		}

		for _, domain := range pendingRemovals {
			if err := client.DeleteDomain(domain); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("pending removal failed for %s: %v", domain, err))
			} else {
				result.Removed = append(result.Removed, domain)
				if clearErr := ClearPendingRemoval(organizationID, domain); clearErr != nil {
					r.Logger.Log(logger.Warning, "failed to clear pending removal for "+domain, clearErr.Error())
				}
			}
		}

		if err := client.Reload(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("reload failed: %v", err))
		}
	}

	r.Logger.Log(logger.Info,
		fmt.Sprintf("reconciliation complete: added=%d updated=%d removed=%d errors=%d",
			len(result.Added), len(result.Updated), len(result.Removed), len(result.Errors)),
		organizationID.String())

	return result, nil
}

func (r *Reconciler) buildDesiredState(ctx context.Context, organizationID uuid.UUID, upstreamHost string) ([]DomainRoute, error) {
	var desired []DomainRoute

	// Source 1: deploy-managed domains from application_domains table.
	appRoutes, err := r.buildDeployDomains(ctx, organizationID, upstreamHost)
	if err != nil {
		return nil, err
	}
	desired = append(desired, appRoutes...)

	// Source 2: extension-managed domains from Redis hash.
	extRoutes, err := GetExtensionDomains(ctx, organizationID)
	if err != nil {
		r.Logger.Log(logger.Warning, "failed to read extension domains from redis", err.Error())
	} else {
		desired = append(desired, extRoutes...)
	}

	return desired, nil
}

func (r *Reconciler) buildDeployDomains(ctx context.Context, organizationID uuid.UUID, upstreamHost string) ([]DomainRoute, error) {
	apps, err := r.Storage.GetDeployedApplications(organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployed apps: %w", err)
	}

	var routes []DomainRoute

	for _, app := range apps {
		if len(app.Domains) == 0 {
			continue
		}

		if app.BuildPack == shared_types.DockerCompose {
			routes = append(routes, r.buildComposeRoutes(app, upstreamHost)...)
		} else {
			routes = append(routes, r.buildSwarmRoutes(ctx, app, upstreamHost)...)
		}
	}

	return routes, nil
}

// buildComposeRoutes resolves ports from the domain's linked ComposeService or
// port override. Orphaned domains (no service, no override) are skipped.
func (r *Reconciler) buildComposeRoutes(app shared_types.Application, upstreamHost string) []DomainRoute {
	var routes []DomainRoute
	for _, d := range app.Domains {
		if d.Domain == "" {
			continue
		}

		port := d.ResolvePort()
		if port == 0 {
			r.Logger.Log(logger.Warning,
				fmt.Sprintf("skipping orphaned compose domain %s (no service linked, no port override)", d.Domain), "")
			continue
		}

		routes = append(routes, DomainRoute{
			Domain:       d.Domain,
			UpstreamDial: fmt.Sprintf("%s:%d", upstreamHost, port),
		})
	}
	return routes
}

// buildSwarmRoutes uses Swarm service discovery to resolve published ports.
func (r *Reconciler) buildSwarmRoutes(ctx context.Context, app shared_types.Application, upstreamHost string) []DomainRoute {
	publishedPort, err := r.getPublishedPort(ctx, app.Name)
	if err != nil {
		r.Logger.Log(logger.Warning,
			fmt.Sprintf("service %s unreachable, skipping %d domain(s)", app.Name, len(app.Domains)),
			err.Error())
		return nil
	}

	dial := fmt.Sprintf("%s:%d", upstreamHost, publishedPort)
	var routes []DomainRoute
	for _, d := range app.Domains {
		if d.Domain == "" {
			continue
		}
		routes = append(routes, DomainRoute{Domain: d.Domain, UpstreamDial: dial})
	}
	return routes
}

func (r *Reconciler) getPublishedPort(ctx context.Context, serviceName string) (int, error) {
	dockerService, err := docker.GetDockerServiceFromContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get docker service: %w", err)
	}

	svc, err := dockerService.GetServiceByName(serviceName)
	if err != nil {
		return 0, fmt.Errorf("failed to get service %s: %w", serviceName, err)
	}
	if svc == nil {
		return 0, fmt.Errorf("service %s not found in swarm", serviceName)
	}

	return extractPublishedPort(*svc)
}

func extractPublishedPort(svc swarm.Service) (int, error) {
	if svc.Endpoint.Ports != nil {
		for _, p := range svc.Endpoint.Ports {
			if p.PublishedPort > 0 {
				return int(p.PublishedPort), nil
			}
		}
	}

	if svc.Spec.EndpointSpec != nil && svc.Spec.EndpointSpec.Ports != nil {
		for _, p := range svc.Spec.EndpointSpec.Ports {
			if p.PublishedPort > 0 {
				return int(p.PublishedPort), nil
			}
		}
	}

	return 0, fmt.Errorf("no published port found for service %s", svc.Spec.Annotations.Name)
}

func (r *Reconciler) fullRebuild(ctx context.Context, desired []DomainRoute) (*ReconcileResult, error) {
	result := &ReconcileResult{}

	if len(desired) == 0 {
		return result, nil
	}

	client, err := GetCaddyClient(ctx, nil, &r.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get caddy client for rebuild: %w", err)
	}

	for _, route := range desired {
		host, port, parseErr := parseDial(route.UpstreamDial)
		if parseErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("invalid dial for %s: %v", route.Domain, parseErr))
			continue
		}
		if err := client.AddDomainWithAutoTLS(route.Domain, host, port, caddygo.DomainOptions{}); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to add %s: %v", route.Domain, err))
		} else {
			result.Added = append(result.Added, route.Domain)
		}
	}

	if err := client.Reload(); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("reload failed: %v", err))
	}

	r.Logger.Log(logger.Info,
		fmt.Sprintf("full rebuild complete: added=%d errors=%d", len(result.Added), len(result.Errors)), "")

	return result, nil
}

func getSSHHostForOrg(ctx context.Context) (string, error) {
	manager, err := ssh.GetSSHManagerFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get SSH manager: %w", err)
	}
	host, err := manager.GetSSHHost()
	if err != nil {
		return "", fmt.Errorf("failed to get SSH host: %w", err)
	}
	return host, nil
}

const (
	QUEUE_CADDY_RECONCILE = "caddy-reconcile"
	TASK_CADDY_RECONCILE  = "task_caddy_reconcile"
	QUEUE_CADDY_HEALTH    = "caddy-health-check"
	TASK_CADDY_HEALTH     = "task_caddy_health_check"
)

var (
	reconcileQueueOnce   sync.Once
	ReconcileQueue       taskq.Queue
	TaskCaddyReconcile   *taskq.Task
	HealthCheckQueue     taskq.Queue
	TaskCaddyHealthCheck *taskq.Task
)

// CaddyTaskPayload is the serializable payload for reconcile/health queue jobs.
type CaddyTaskPayload struct {
	OrganizationID string `json:"organization_id"`
}

// ReconcilerDaemon enqueues reconciliation jobs into Redis for distributed
// processing. Multiple API instances share the same queue; taskq handles
// worker distribution, retries, and deduplication.
type ReconcilerDaemon struct {
	reconciler *Reconciler
	interval   time.Duration
	orgFetcher func(ctx context.Context) ([]uuid.UUID, error)
	stopCh     chan struct{}
	logger     logger.Logger
}

func NewReconcilerDaemon(
	store storage.DeployRepository,
	lgr logger.Logger,
	interval time.Duration,
	orgFetcher func(ctx context.Context) ([]uuid.UUID, error),
) *ReconcilerDaemon {
	return &ReconcilerDaemon{
		reconciler: NewReconciler(store, lgr),
		interval:   interval,
		orgFetcher: orgFetcher,
		stopCh:     make(chan struct{}),
		logger:     lgr,
	}
}

// SetupQueues registers the reconciliation and health-check queues with the
// shared Redis-backed taskq factory. Must be called after queue.Init().
func (d *ReconcilerDaemon) SetupQueues() {
	reconcileQueueOnce.Do(func() {
		ReconcileQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_CADDY_RECONCILE,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        2,
			MaxNumWorker:        8,
			ReservationSize:     1,
			ReservationTimeout:  5 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          64,
		})

		TaskCaddyReconcile = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_CADDY_RECONCILE,
			RetryLimit: 2,
			Handler: func(ctx context.Context, payload CaddyTaskPayload) error {
				orgID, err := uuid.Parse(payload.OrganizationID)
				if err != nil {
					return fmt.Errorf("invalid org ID in reconcile task: %w", err)
				}
				result, err := d.reconciler.ReconcileOrganization(ctx, orgID)
				if err != nil {
					return err
				}
				if len(result.Errors) > 0 {
					d.logger.Log(logger.Warning,
						fmt.Sprintf("reconciliation for org %s had %d errors", orgID, len(result.Errors)),
						fmt.Sprintf("%v", result.Errors))
				}
				return nil
			},
		})

		HealthCheckQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_CADDY_HEALTH,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        4,
			MaxNumWorker:        16,
			ReservationSize:     1,
			ReservationTimeout:  2 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          128,
		})
	})
}

func (d *ReconcilerDaemon) Start(ctx context.Context) {
	go d.run(ctx)
}

func (d *ReconcilerDaemon) Stop() {
	close(d.stopCh)
}

// Reconciler returns the underlying reconciler for on-demand reconciliation.
func (d *ReconcilerDaemon) Reconciler() *Reconciler {
	return d.reconciler
}

// EnqueueReconcile adds a single org reconciliation job to the Redis queue.
// Safe to call from any goroutine or API instance.
func EnqueueReconcile(orgID uuid.UUID) error {
	if ReconcileQueue == nil || TaskCaddyReconcile == nil {
		return fmt.Errorf("reconcile queue not initialized")
	}

	msg := TaskCaddyReconcile.WithArgs(context.Background(), CaddyTaskPayload{
		OrganizationID: orgID.String(),
	})
	msg.OnceInPeriod(30 * time.Second)

	return ReconcileQueue.Add(msg)
}

func (d *ReconcilerDaemon) run(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopCh:
			d.logger.Log(logger.Info, "reconciler daemon stopped", "")
			return
		case <-ctx.Done():
			d.logger.Log(logger.Info, "reconciler daemon context cancelled", "")
			return
		case <-ticker.C:
			d.enqueueAll(ctx)
		}
	}
}

// enqueueAll fetches all org IDs and enqueues a reconcile job for each.
// taskq's OnceInPeriod prevents duplicate processing.
func (d *ReconcilerDaemon) enqueueAll(ctx context.Context) {
	orgIDs, err := d.orgFetcher(ctx)
	if err != nil {
		d.logger.Log(logger.Error, "failed to fetch organizations for reconciliation", err.Error())
		return
	}

	enqueued := 0
	for _, orgID := range orgIDs {
		if err := EnqueueReconcile(orgID); err != nil {
			d.logger.Log(logger.Warning, "failed to enqueue reconcile for org "+orgID.String(), err.Error())
		} else {
			enqueued++
		}
	}

	if enqueued > 0 {
		d.logger.Log(logger.Info,
			fmt.Sprintf("enqueued %d/%d reconciliation jobs", enqueued, len(orgIDs)), "")
	}
}

// FormatDial creates the dial string used for Caddy upstream configuration.
func FormatDial(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}
