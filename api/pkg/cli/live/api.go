package live

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/httpclient"
)

// getApplicationDetails fetches application details from the server to get base_path
// Returns base_path (defaults to "/" if empty)
func getApplicationDetails(server, applicationID, accessToken string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/deploy/application?id=%s", server, applicationID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add auth headers (Bearer token + X-Organization-Id)
	httpclient.SetAuthHeaders(req, accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch application: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch application: status %d, body: %s", resp.StatusCode, string(body))
	}

	var appResp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			BasePath string `json:"base_path"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&appResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if appResp.Status != "success" {
		return "", fmt.Errorf("application fetch failed: %s", appResp.Message)
	}

	basePath := appResp.Data.BasePath
	if basePath == "" {
		basePath = "/"
	}

	return basePath, nil
}
