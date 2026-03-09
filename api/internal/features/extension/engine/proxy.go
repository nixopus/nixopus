package engine

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/caddy"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	"github.com/nixopus/nixopus/api/internal/types"
)

type proxyModule struct{}

func (proxyModule) Type() string { return "proxy" }

func AddDomainToProxy(ctx context.Context, sshClient *ssh.SSH, domain string, port string, upstreamHost string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}
	dial := caddy.FormatDial(upstreamHost, p)
	if err := caddy.AddDomainsWithRetry(ctx, sshClient, nil, []caddy.DomainRoute{
		{Domain: domain, UpstreamDial: dial},
	}); err != nil {
		return err
	}

	if orgID, ok := orgIDFromCtx(ctx); ok {
		if tErr := caddy.TrackExtensionDomain(orgID, domain, dial); tErr != nil {
			log.Printf("failed to track extension domain %s: %v", domain, tErr)
		}
	}
	return nil
}

func UpdateDomainInProxy(ctx context.Context, sshClient *ssh.SSH, domain string, port string, upstreamHost string) error {
	return AddDomainToProxy(ctx, sshClient, domain, port, upstreamHost)
}

func RemoveDomainFromProxy(ctx context.Context, sshClient *ssh.SSH, domain string) error {
	if err := caddy.RemoveDomainsWithRetry(ctx, sshClient, nil, []string{domain}); err != nil {
		if orgID, ok := orgIDFromCtx(ctx); ok {
			if enqErr := caddy.EnqueuePendingRemoval(orgID, domain); enqErr != nil {
				log.Printf("failed to enqueue pending removal for %s: %v", domain, enqErr)
			}
		}
		return err
	}

	if orgID, ok := orgIDFromCtx(ctx); ok {
		if tErr := caddy.UntrackExtensionDomain(orgID, domain); tErr != nil {
			log.Printf("failed to untrack extension domain %s: %v", domain, tErr)
		}
	}
	return nil
}

func orgIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	s, ok := ctx.Value(types.OrganizationIDKey).(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func (proxyModule) Execute(ctx context.Context, sshClient *ssh.SSH, step types.SpecStep, vars map[string]interface{}) (string, func(), error) {
	action, _ := step.Properties["action"].(string)
	domain, _ := step.Properties["domain"].(string)
	port, _ := step.Properties["port"].(string)

	if domain != "" {
		domain = replaceVars(domain, vars)
	}
	if port != "" {
		port = replaceVars(port, vars)
	}

	switch action {
	case "add":
		if domain == "" || port == "" {
			return "proxy step skipped: domain and port are optional", nil, nil
		}
		if sshClient.Host == "" {
			return "", nil, fmt.Errorf("SSH host not configured")
		}
		if err := AddDomainToProxy(ctx, sshClient, domain, port, sshClient.Host); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("proxy added for %s -> %s:%s", domain, sshClient.Host, port), nil, nil
	case "update":
		if domain == "" || port == "" {
			return "proxy step skipped: domain and port are optional", nil, nil
		}
		if sshClient.Host == "" {
			return "", nil, fmt.Errorf("SSH host not configured")
		}
		if err := UpdateDomainInProxy(ctx, sshClient, domain, port, sshClient.Host); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("proxy updated for %s -> %s:%s", domain, sshClient.Host, port), nil, nil
	case "remove":
		if domain == "" {
			return "proxy step skipped: domain is optional", nil, nil
		}
		if err := RemoveDomainFromProxy(ctx, sshClient, domain); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("proxy removed for %s", domain), nil, nil
	default:
		return "", nil, fmt.Errorf("unsupported proxy action: %s", action)
	}
}

func init() {
	RegisterModule(proxyModule{})
}
