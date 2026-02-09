package live

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/httpclient"
	"github.com/raghavyuva/nixopus-api/internal/mover"
)

const (
	pollInterval = 2 * time.Second
	apiTimeout   = 5 * time.Second
)

// Deployment represents a deployment from API
type Deployment struct {
	ID          string  `json:"id"`
	Status      *Status `json:"status,omitempty"`
	Logs        []Log   `json:"logs,omitempty"`
	ContainerID string  `json:"container_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// Status represents deployment status
type Status struct {
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
}

// Log represents a deployment log entry
type Log struct {
	Log       string `json:"log"`
	CreatedAt string `json:"created_at"`
}

// DeploymentPoller polls the API for deployment status updates
type DeploymentPoller struct {
	config        *config.Config
	tracker       *mover.Tracker
	client        *http.Client
	logFetcher    *LogFetcher
	applicationID string
	accessToken   string
	stop          chan struct{}
}

// NewDeploymentPoller creates a new deployment poller
func NewDeploymentPoller(cfg *config.Config, tracker *mover.Tracker, applicationID string) *DeploymentPoller {
	// Get access token from global auth storage
	accessToken, _ := config.GetAccessToken()

	return &DeploymentPoller{
		config:        cfg,
		tracker:       tracker,
		applicationID: applicationID,
		accessToken:   accessToken,
		client: &http.Client{
			Timeout: apiTimeout,
		},
		logFetcher: NewLogFetcher(&Config{
			Server:      cfg.Server,
			AccessToken: accessToken,
			Timeout:     apiTimeout,
		}),
		stop: make(chan struct{}),
	}
}

// Start begins polling for deployment status
func (p *DeploymentPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Poll immediately
	p.poll()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stop:
			return
		case <-ticker.C:
			p.poll()
		}
	}
}

// Stop stops the poller
func (p *DeploymentPoller) Stop() {
	close(p.stop)
}

// poll fetches the current deployment status from the API
func (p *DeploymentPoller) poll() {
	// Use deployments endpoint which includes status and logs
	url := fmt.Sprintf("%s/api/v1/deploy/application/deployments?id=%s&limit=1", p.config.Server, p.applicationID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return // Silently fail, will retry on next poll
	}

	req.Header.Set("Content-Type", "application/json")
	httpclient.SetAuthHeaders(req, p.accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return // Silently fail, will retry on next poll
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return // Silently fail, will retry on next poll
	}

	if resp.StatusCode != http.StatusOK {
		return // Silently fail, will retry on next poll
	}

	var deploymentsResp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Deployments []Deployment `json:"deployments"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &deploymentsResp); err != nil {
		return // Silently fail, will retry on next poll
	}

	// Extract latest deployment status
	if len(deploymentsResp.Data.Deployments) > 0 {
		deploymentInfo := p.extractDeploymentInfoFromDeployment(&deploymentsResp.Data.Deployments[0])
		if deploymentInfo != nil {
			p.tracker.SetDeploymentInfo(deploymentInfo)
		} else {
			// Deployment exists but status is nil
			p.tracker.SetDeploymentInfo(&mover.DeploymentInfo{
				Status:    "unknown",
				Message:   "Deployment found but status unavailable",
				Logs:      []string{},
				UpdatedAt: "",
			})
		}
	} else {
		// No deployments yet
		p.tracker.SetDeploymentInfo(&mover.DeploymentInfo{
			Status:    "pending",
			Message:   "Waiting for deployment to start...",
			Logs:      []string{},
			UpdatedAt: "",
		})
	}
}

// extractDeploymentInfoFromDeployment extracts deployment status from a single deployment
func (p *DeploymentPoller) extractDeploymentInfoFromDeployment(deployment *Deployment) *mover.DeploymentInfo {
	if deployment == nil {
		return nil
	}

	if deployment.Status == nil {
		return nil
	}

	// Convert build logs to BuildLog format
	buildLogs := make([]BuildLog, len(deployment.Logs))
	for i, log := range deployment.Logs {
		buildLogs[i] = BuildLog{
			Log:       log.Log,
			CreatedAt: log.CreatedAt,
		}
	}

	// Fetch combined logs (build + container) in parallel
	maxLogs := 50 // Fetch more logs to have a good mix
	combinedLogs := p.logFetcher.FetchCombinedLogs(buildLogs, deployment.ContainerID, p.applicationID, maxLogs)

	status := deployment.Status.Status
	message := ""
	errorMsg := ""

	// Set appropriate message based on status
	switch status {
	case "building":
		message = "Building your application..."
	case "deploying":
		message = "Deploying your application..."
	case "deployed":
		message = "Application deployed successfully"
	case "failed":
		errorMsg = "Deployment failed"
		// Try to extract error from logs (check latest logs first)
		for i := 0; i < len(combinedLogs) && i < 10; i++ {
			if combinedLogs[i] != "" {
				errorMsg = combinedLogs[i]
				break
			}
		}
	case "cloning":
		message = "Cloning repository..."
	default:
		message = fmt.Sprintf("Status: %s", status)
	}

	return &mover.DeploymentInfo{
		Status:    status,
		Message:   message,
		Error:     errorMsg,
		Logs:      combinedLogs,
		UpdatedAt: deployment.Status.UpdatedAt,
	}
}
