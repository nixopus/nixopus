package caddy

import (
	"context"
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/deploy/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func BuildMultiUpstreamRoutes(
	ctx context.Context,
	store storage.DeployRepository,
	lgr *logger.Logger,
	app shared_types.Application,
	domains []shared_types.ApplicationDomain,
	fallbackUpstream string,
	fallbackPort int,
) []DomainRoute {
	if len(domains) == 0 {
		return nil
	}

	servers, err := store.GetApplicationServers(app.ID)
	if err != nil || len(servers) <= 1 {
		return buildSingleUpstreamRoutes(domains, fallbackUpstream, fallbackPort, app.BuildPack)
	}

	lbPolicy := mapRoutingStrategy(app.RoutingStrategy)

	upstreams, primaryUpstreamIndex, err := resolveServerUpstreams(ctx, servers, fallbackPort, lgr)
	if err != nil || len(upstreams) == 0 {
		return buildSingleUpstreamRoutes(domains, fallbackUpstream, fallbackPort, app.BuildPack)
	}

	if app.RoutingStrategy == shared_types.RoutingStrategyPrimaryFailover {
		upstreams = orderPrimaryFirst(upstreams, primaryUpstreamIndex)
	}

	var routes []DomainRoute
	for _, d := range domains {
		if d.Domain == "" {
			continue
		}
		port := fallbackPort
		if app.BuildPack == shared_types.DockerCompose {
			port = d.ResolvePort()
			if port == 0 {
				continue
			}
		}
		var domainUpstreams []string
		if app.BuildPack == shared_types.DockerCompose {
			for _, u := range upstreams {
				host, _, err := parseDial(u)
				if err != nil {
					lgr.Log(logger.Warning, fmt.Sprintf("skipping malformed upstream %s: %v", u, err), "")
					continue
				}
				domainUpstreams = append(domainUpstreams, FormatDial(host, port))
			}
		} else {
			domainUpstreams = upstreams
		}

		routes = append(routes, DomainRoute{
			Domain:    d.Domain,
			Upstreams: domainUpstreams,
			LBPolicy:  lbPolicy,
		})
	}
	return routes
}

func buildSingleUpstreamRoutes(
	domains []shared_types.ApplicationDomain,
	upstreamHost string,
	port int,
	buildPack shared_types.BuildPack,
) []DomainRoute {
	var routes []DomainRoute
	for _, d := range domains {
		if d.Domain == "" {
			continue
		}
		p := port
		if buildPack == shared_types.DockerCompose {
			p = d.ResolvePort()
			if p == 0 {
				continue
			}
		}
		routes = append(routes, DomainRoute{
			Domain:       d.Domain,
			UpstreamDial: FormatDial(upstreamHost, p),
		})
	}
	return routes
}

func resolveServerUpstreams(ctx context.Context, servers []shared_types.ApplicationServer, port int, lgr *logger.Logger) ([]string, int, error) {
	var upstreams []string
	primaryUpstreamIndex := -1
	for _, srv := range servers {
		if srv.Server == nil || srv.Server.Host == nil {
			continue
		}
		host := ""
		if srv.Server.ProxyHost != nil && *srv.Server.ProxyHost != "" {
			host = *srv.Server.ProxyHost
		} else {
			host = *srv.Server.Host
		}
		if srv.IsPrimary {
			primaryUpstreamIndex = len(upstreams)
		}
		upstreams = append(upstreams, FormatDial(host, port))
	}
	return upstreams, primaryUpstreamIndex, nil
}

func orderPrimaryFirst(upstreams []string, primaryUpstreamIndex int) []string {
	if primaryUpstreamIndex <= 0 || primaryUpstreamIndex >= len(upstreams) {
		return upstreams
	}
	result := make([]string, len(upstreams))
	result[0] = upstreams[primaryUpstreamIndex]
	copy(result[1:], upstreams[:primaryUpstreamIndex])
	copy(result[1+primaryUpstreamIndex:], upstreams[primaryUpstreamIndex+1:])
	return result
}

func mapRoutingStrategy(strategy shared_types.RoutingStrategy) string {
	switch strategy {
	case shared_types.RoutingStrategyRoundRobin:
		return "round_robin"
	case shared_types.RoutingStrategyPrimaryFailover:
		return "first"
	default:
		return ""
	}
}