package caddy

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	"github.com/raghavyuva/caddygo"
)

// DomainRoute represents a domain-to-upstream mapping extracted from Caddy config.
type DomainRoute struct {
	Domain       string
	UpstreamDial string // host:port format
}

// AddDomainsWithRetry adds multiple domains to Caddy with retry and tunnel
// recovery. All domains are added, then a single Reload is issued. On failure
// the stale tunnel is invalidated before retry.
func AddDomainsWithRetry(ctx context.Context, sshClient *ssh.SSH, lgr *logger.Logger, domains []DomainRoute) error {
	if len(domains) == 0 {
		return nil
	}

	l := resolveLogger(lgr)

	return WithRetry(func() error {
		client, err := GetCaddyClient(ctx, sshClient, lgr)
		if err != nil {
			return fmt.Errorf("failed to get caddy client: %w", err)
		}

		for _, d := range domains {
			host, port, parseErr := parseDial(d.UpstreamDial)
			if parseErr != nil {
				return fmt.Errorf("invalid upstream %s: %w", d.UpstreamDial, parseErr)
			}
			if err := client.AddDomainWithAutoTLS(d.Domain, host, port, caddygo.DomainOptions{}); err != nil {
				return fmt.Errorf("failed to add domain %s: %w", d.Domain, err)
			}
		}

		if err := client.Reload(); err != nil {
			return fmt.Errorf("failed to reload caddy: %w", err)
		}
		return nil
	}, DefaultMaxRetries, l, func() {
		invalidateTunnelFromCtx(ctx, sshClient)
	})
}

// RemoveDomainsWithRetry removes multiple domains from Caddy with retry.
func RemoveDomainsWithRetry(ctx context.Context, sshClient *ssh.SSH, lgr *logger.Logger, domains []string) error {
	if len(domains) == 0 {
		return nil
	}

	l := resolveLogger(lgr)

	return WithRetry(func() error {
		client, err := GetCaddyClient(ctx, sshClient, lgr)
		if err != nil {
			return fmt.Errorf("failed to get caddy client: %w", err)
		}

		for _, domain := range domains {
			if err := client.DeleteDomain(domain); err != nil {
				return fmt.Errorf("failed to remove domain %s: %w", domain, err)
			}
		}

		if err := client.Reload(); err != nil {
			return fmt.Errorf("failed to reload caddy: %w", err)
		}
		return nil
	}, DefaultMaxRetries, l, func() {
		invalidateTunnelFromCtx(ctx, sshClient)
	})
}

// PingCaddy checks if the Caddy admin API is reachable through the SSH tunnel.
func PingCaddy(ctx context.Context, sshClient *ssh.SSH, lgr *logger.Logger) error {
	client, err := GetCaddyClient(ctx, sshClient, lgr)
	if err != nil {
		return fmt.Errorf("failed to get caddy client: %w", err)
	}

	if err := client.Reload(); err != nil {
		return fmt.Errorf("caddy admin API unreachable: %w", err)
	}
	return nil
}

// GetCurrentDomains reads the current Caddy config and returns all configured
// domain-to-upstream mappings. This is used by the reconciler to diff against
// the database state.
func GetCurrentDomains(ctx context.Context, sshClient *ssh.SSH, lgr *logger.Logger) ([]DomainRoute, error) {
	client, err := GetCaddyClient(ctx, sshClient, lgr)
	if err != nil {
		return nil, fmt.Errorf("failed to get caddy client: %w", err)
	}

	resp, err := client.HTTPClient.Get(client.BaseURL + "/config/")
	if err != nil {
		return nil, fmt.Errorf("failed to get caddy config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("caddy config request failed with status %d", resp.StatusCode)
	}

	var config caddy.Config
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode caddy config: %w", err)
	}

	return extractDomainRoutes(&config)
}

// GetCaddyConfig fetches the raw Caddy config for snapshot/restore operations.
func GetCaddyConfig(ctx context.Context, sshClient *ssh.SSH, lgr *logger.Logger) (*caddy.Config, error) {
	client, err := GetCaddyClient(ctx, sshClient, lgr)
	if err != nil {
		return nil, fmt.Errorf("failed to get caddy client: %w", err)
	}

	resp, err := client.HTTPClient.Get(client.BaseURL + "/config/")
	if err != nil {
		return nil, fmt.Errorf("failed to get caddy config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("caddy config request failed with status %d", resp.StatusCode)
	}

	var config caddy.Config
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode caddy config: %w", err)
	}

	return &config, nil
}

// RestoreCaddyConfig restores a previously captured Caddy config snapshot.
func RestoreCaddyConfig(ctx context.Context, sshClient *ssh.SSH, lgr *logger.Logger, config *caddy.Config) error {
	client, err := GetCaddyClient(ctx, sshClient, lgr)
	if err != nil {
		return fmt.Errorf("failed to get caddy client: %w", err)
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config snapshot: %w", err)
	}

	resp, err := client.HTTPClient.Post(client.BaseURL+"/load", "application/json", jsonReader(configJSON))
	if err != nil {
		return fmt.Errorf("failed to restore caddy config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("caddy restore failed with status %d", resp.StatusCode)
	}
	return nil
}

// AddDomainsAtomic snapshots the Caddy config, applies domain additions,
// and rolls back on failure.
func AddDomainsAtomic(ctx context.Context, sshClient *ssh.SSH, lgr *logger.Logger, domains []DomainRoute) error {
	if len(domains) == 0 {
		return nil
	}

	snapshot, err := GetCaddyConfig(ctx, sshClient, lgr)
	if err != nil {
		return fmt.Errorf("failed to snapshot caddy config: %w", err)
	}

	if addErr := AddDomainsWithRetry(ctx, sshClient, lgr, domains); addErr != nil {
		l := resolveLogger(lgr)
		l.Log(logger.Warning, "domain add failed, rolling back caddy config", addErr.Error())

		if restoreErr := RestoreCaddyConfig(ctx, sshClient, lgr, snapshot); restoreErr != nil {
			return fmt.Errorf("add failed AND rollback failed: %w (original: %v)", restoreErr, addErr)
		}
		return fmt.Errorf("domain add rolled back: %w", addErr)
	}
	return nil
}

// RemoveDomainsAtomic snapshots the Caddy config, applies domain removals,
// and rolls back on failure.
func RemoveDomainsAtomic(ctx context.Context, sshClient *ssh.SSH, lgr *logger.Logger, domains []string) error {
	if len(domains) == 0 {
		return nil
	}

	snapshot, err := GetCaddyConfig(ctx, sshClient, lgr)
	if err != nil {
		return fmt.Errorf("failed to snapshot caddy config: %w", err)
	}

	if removeErr := RemoveDomainsWithRetry(ctx, sshClient, lgr, domains); removeErr != nil {
		l := resolveLogger(lgr)
		l.Log(logger.Warning, "domain remove failed, rolling back caddy config", removeErr.Error())

		if restoreErr := RestoreCaddyConfig(ctx, sshClient, lgr, snapshot); restoreErr != nil {
			return fmt.Errorf("remove failed AND rollback failed: %w (original: %v)", restoreErr, removeErr)
		}
		return fmt.Errorf("domain remove rolled back: %w", removeErr)
	}
	return nil
}

// --- internal helpers ---

func resolveLogger(lgr *logger.Logger) logger.Logger {
	if lgr != nil {
		return *lgr
	}
	return logger.NewLogger()
}

func invalidateTunnelFromCtx(ctx context.Context, sshClient *ssh.SSH) {
	if sshClient != nil && sshClient.Host != "" {
		InvalidateTunnel(sshClient.Host)
	}
}

// extractDomainRoutes parses the Caddy config to find all domain→upstream mappings
// from the "nixopus" server block.
func extractDomainRoutes(config *caddy.Config) ([]DomainRoute, error) {
	if config.AppsRaw == nil {
		return nil, nil
	}

	httpAppRaw, exists := config.AppsRaw["http"]
	if !exists {
		return nil, nil
	}

	var httpApp caddyhttp.App
	if err := json.Unmarshal(httpAppRaw, &httpApp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal http app: %w", err)
	}

	server := httpApp.Servers["nixopus"]
	if server == nil {
		return nil, nil
	}

	var routes []DomainRoute
	for _, route := range server.Routes {
		domain := extractDomainFromRoute(route)
		if domain == "" {
			continue
		}

		upstream := extractUpstreamFromRoute(route)
		if upstream == "" {
			continue
		}

		routes = append(routes, DomainRoute{
			Domain:       domain,
			UpstreamDial: upstream,
		})
	}

	return routes, nil
}

func extractDomainFromRoute(route caddyhttp.Route) string {
	for _, matcherSet := range route.MatcherSetsRaw {
		hostRaw, exists := matcherSet["host"]
		if !exists {
			continue
		}
		var hosts caddyhttp.MatchHost
		if err := json.Unmarshal(hostRaw, &hosts); err == nil && len(hosts) > 0 {
			return string(hosts[0])
		}
	}
	return ""
}

func extractUpstreamFromRoute(route caddyhttp.Route) string {
	for _, handlerRaw := range route.HandlersRaw {
		var handlerMap map[string]json.RawMessage
		if err := json.Unmarshal(handlerRaw, &handlerMap); err != nil {
			continue
		}

		handlerNameRaw, exists := handlerMap["handler"]
		if !exists {
			continue
		}
		var handlerName string
		if err := json.Unmarshal(handlerNameRaw, &handlerName); err != nil || handlerName != "reverse_proxy" {
			continue
		}

		upstreamsRaw, exists := handlerMap["upstreams"]
		if !exists {
			continue
		}
		var upstreams []map[string]interface{}
		if err := json.Unmarshal(upstreamsRaw, &upstreams); err != nil || len(upstreams) == 0 {
			continue
		}

		if dial, ok := upstreams[0]["dial"].(string); ok {
			return dial
		}
	}
	return ""
}

func parseDial(dial string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(dial)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse dial string %q: %w", dial, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("failed to parse dial string %q: invalid port %q", dial, portStr)
	}
	return host, port, nil
}

func jsonReader(data []byte) *jsonBody {
	return &jsonBody{data: data, pos: 0}
}

type jsonBody struct {
	data []byte
	pos  int
}

func (j *jsonBody) Read(p []byte) (int, error) {
	if j.pos >= len(j.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, j.data[j.pos:])
	j.pos += n
	return n, nil
}
