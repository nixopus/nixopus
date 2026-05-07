package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// injectUserContext builds the [user-context] block for the first message in a conversation.
// It queries the org's current state directly from the database.
func (s *AgentService) injectUserContext(ctx context.Context, orgID string) string {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		s.logger.Log(logger.Error, "context_injector: invalid org_id", err.Error())
		return ""
	}

	var parts []string

	if line := s.fetchAppsContext(ctx, orgUUID); line != "" {
		parts = append(parts, line)
	}
	if line := s.fetchDomainsContext(ctx, orgUUID); line != "" {
		parts = append(parts, line)
	}
	if line := s.fetchServersContext(ctx, orgUUID); line != "" {
		parts = append(parts, line)
	}
	if line := s.fetchConnectorsContext(ctx, orgUUID); line != "" {
		parts = append(parts, line)
	}
	if line := s.fetchMCPServersContext(ctx, orgUUID); line != "" {
		parts = append(parts, line)
	}

	// Only emit the block when there is actual org data to show.
	// Instance metadata is prepended so the LLM knows the hosting mode.
	if len(parts) == 0 {
		return ""
	}

	all := append([]string{s.fetchInstanceContext()}, parts...)
	return "[user-context]\n" + strings.Join(all, "\n") + "\n[/user-context]"
}

func (s *AgentService) fetchAppsContext(ctx context.Context, orgID uuid.UUID) string {
	var apps []shared_types.Application
	err := s.store.DB.NewSelect().
		Model(&apps).
		Relation("Status").
		Relation("Domains").
		Where("a.organization_id = ?", orgID).
		Limit(25).
		Scan(ctx)
	if err != nil || len(apps) == 0 {
		return ""
	}

	entries := make([]string, 0, len(apps))
	for _, app := range apps {
		status := "unknown"
		if app.Status != nil {
			status = string(app.Status.Status)
		}
		domains := ""
		if len(app.Domains) > 0 {
			domainNames := make([]string, 0, len(app.Domains))
			for _, d := range app.Domains {
				domainNames = append(domainNames, d.Domain)
			}
			domains = ",domains:" + strings.Join(domainNames, ";")
		}
		entries = append(entries, fmt.Sprintf("%s(id:%s,status:%s,port:%d,branch:%s%s)",
			app.Name, app.ID.String(), status, app.Port, app.Branch, domains))
	}
	return "apps: " + strings.Join(entries, " | ")
}

func (s *AgentService) fetchDomainsContext(ctx context.Context, orgID uuid.UUID) string {
	var domains []shared_types.Domain
	err := s.store.DB.NewSelect().
		Model(&domains).
		Where("do.organization_id = ?", orgID).
		Where("do.deleted_at IS NULL").
		Limit(20).
		Scan(ctx)
	if err != nil || len(domains) == 0 {
		return ""
	}

	entries := make([]string, 0, len(domains))
	for _, d := range domains {
		entries = append(entries, fmt.Sprintf("%s(id:%s,type:%s,status:%s)",
			d.Name, d.ID.String(), d.Type, d.Status))
	}
	return "domains: " + strings.Join(entries, " | ")
}

func (s *AgentService) fetchServersContext(ctx context.Context, orgID uuid.UUID) string {
	var servers []shared_types.SSHKey
	err := s.store.DB.NewSelect().
		Model(&servers).
		Where("sk.organization_id = ?", orgID).
		Limit(10).
		Scan(ctx)
	if err != nil || len(servers) == 0 {
		return ""
	}

	entries := make([]string, 0, len(servers))
	for _, srv := range servers {
		host := ""
		if srv.Host != nil {
			host = *srv.Host
		}
		entries = append(entries, fmt.Sprintf("%s(id:%s,host:%s)",
			srv.Name, srv.ID.String(), host))
	}
	return "servers: " + strings.Join(entries, " | ")
}

func (s *AgentService) fetchConnectorsContext(ctx context.Context, orgID uuid.UUID) string {
	// GitHub connectors are associated with users; find users in this org
	var connectors []shared_types.GithubConnector
	err := s.store.DB.NewSelect().
		Model(&connectors).
		Where("gc.user_id IN (SELECT user_id FROM member WHERE organization_id = ?)", orgID).
		Where("gc.deleted_at IS NULL").
		Limit(10).
		Scan(ctx)
	if err != nil || len(connectors) == 0 {
		return ""
	}

	entries := make([]string, 0, len(connectors))
	for _, c := range connectors {
		entries = append(entries, fmt.Sprintf("%s(id:%s)", c.Slug, c.ID.String()))
	}
	return "connectors: " + strings.Join(entries, " | ")
}

func (s *AgentService) fetchMCPServersContext(ctx context.Context, orgID uuid.UUID) string {
	var servers []shared_types.MCPServer
	err := s.store.DB.NewSelect().
		Model(&servers).
		Where("ms.org_id = ?", orgID).
		Where("ms.enabled = true").
		Where("ms.deleted_at IS NULL").
		Limit(15).
		Scan(ctx)
	if err != nil || len(servers) == 0 {
		return ""
	}

	entries := make([]string, 0, len(servers))
	for _, srv := range servers {
		entries = append(entries, fmt.Sprintf("%s(id:%s,provider:%s)",
			srv.Name, srv.ID.String(), srv.ProviderID))
	}
	return "mcp_servers: " + strings.Join(entries, " | ")
}

func (s *AgentService) fetchInstanceContext() string {
	mode := "cloud"
	if config.AppConfig.App.SelfHosted {
		mode = "self-hosted"
	}
	domain := config.AppConfig.App.DeployDomain
	if domain == "" {
		domain = "not configured"
	}
	return fmt.Sprintf("instance: mode=%s, deploy_domain=%s", mode, domain)
}
