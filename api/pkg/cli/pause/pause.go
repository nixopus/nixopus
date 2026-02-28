package pause

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/httpclient"
	"github.com/raghavyuva/nixopus-api/pkg/cli/logincmd"
	"github.com/spf13/cobra"
)

var PauseCmd = &cobra.Command{
	Use:   "pause [app-name]",
	Short: "Pause the live dev service",
	Long: `Pause the live dev service for the current project. This scales the service to 0 replicas.
When you run 'nixopus live' again, the service will resume automatically without a full rebuild.
Optionally specify an app name to pause a specific application in a multi-app family.`,
	RunE: runPause,
}

func runPause(cmd *cobra.Command, args []string) error {
	if err := logincmd.EnsureAuthenticated(); err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config. Run 'nixopus live' to initialize first: %w", err)
	}

	appName := ""
	if len(args) > 0 {
		appName = args[0]
	}

	applicationID, err := cfg.GetApplicationID(appName)
	if err != nil {
		return fmt.Errorf("failed to get application ID: %w", err)
	}

	accessToken, err := config.GetAccessToken()
	if err != nil {
		return fmt.Errorf("not authenticated. Please run 'nixopus login' first: %w", err)
	}

	baseURL := httpclient.BuildURL(cfg.Server, "/api/v1/live/pause")
	reqURL := baseURL + "?application_id=" + url.QueryEscape(applicationID)

	client := httpclient.NewAuthenticatedHTTPClient(accessToken)
	resp, err := client.Post(reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to pause: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := httpclient.ReadResponseBody(resp)
		return fmt.Errorf("pause failed: %w", httpclient.HandleErrorResponse(resp, body, "pause request failed"))
	}

	fmt.Println("Live dev service paused. Run 'nixopus live' to resume.")
	return nil
}
