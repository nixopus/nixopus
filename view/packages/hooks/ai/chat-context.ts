'use client';

import { Layers, Box, Globe, GitFork, Plug, type LucideIcon } from 'lucide-react';
import { useGetApplicationsQuery } from '@/redux/services/deploy/applicationsApi';
import { useGetContainersQuery } from '@/redux/services/container/containerApi';
import { useGetAllDomainsQuery } from '@/redux/services/settings/domainsApi';
import { useGetAllGithubRepositoriesQuery } from '@/redux/services/connector/githubConnectorApi';
import { useGetMCPServersQuery } from '@/redux/services/settings/mcpApi';

export interface ChatContext {
  type: string;
  id: string;
  label: string;
  meta?: Record<string, string>;
}

export interface ContextProviderConfig {
  type: string;
  icon: LucideIcon;
  labelKey: string;
}

export interface ContextProviderData {
  config: ContextProviderConfig;
  items: ChatContext[];
  isLoading: boolean;
}

export function formatContextsForAgent(contexts: ChatContext[]): string {
  if (contexts.length === 0) return '';

  const parts = contexts.map((ctx) => {
    const metaStr = ctx.meta
      ? Object.entries(ctx.meta)
          .map(([k, v]) => `${k}: ${v}`)
          .join(', ')
      : '';
    return `[Context: ${ctx.type} "${ctx.label}"${metaStr ? ` (${metaStr})` : ''}]`;
  });

  return parts.join('\n') + '\n\n';
}

/** Strip context prefix from message text so only user-visible content is shown. */
export function stripContextFromMessageText(text: string): string {
  return text.replace(/^(\[Context:[^\]]+\]\s*\n?)*\s*/i, '').trimStart();
}

function useAppsContextProvider(): ContextProviderData {
  const { data, isLoading } = useGetApplicationsQuery({ page: 1, limit: 100 });

  const items: ChatContext[] = (data?.applications ?? []).map((app) => ({
    type: 'App',
    id: app.id,
    label: app.name,
    meta: {
      ID: app.id,
      Environment: app.environment
    }
  }));

  return {
    config: { type: 'App', icon: Layers, labelKey: 'ai.context.apps' },
    items,
    isLoading
  };
}

function useContainersContextProvider(): ContextProviderData {
  const { data, isLoading } = useGetContainersQuery({ page: 1, page_size: 100 });

  const items: ChatContext[] = (data?.containers ?? []).map((c) => ({
    type: 'Container',
    id: c.id,
    label: c.name.replace(/^\//, ''),
    meta: {
      ID: c.id,
      Image: c.image,
      Status: c.state
    }
  }));

  return {
    config: { type: 'Container', icon: Box, labelKey: 'ai.context.containers' },
    items,
    isLoading
  };
}

function useDomainsContextProvider(): ContextProviderData {
  const { data, isLoading } = useGetAllDomainsQuery();

  const items: ChatContext[] = (data ?? []).map((d) => ({
    type: 'Domain',
    id: d.id,
    label: d.name,
    meta: { ID: d.id }
  }));

  return {
    config: { type: 'Domain', icon: Globe, labelKey: 'ai.context.domains' },
    items,
    isLoading
  };
}

function useRepositoriesContextProvider(): ContextProviderData {
  const { data, isLoading } = useGetAllGithubRepositoriesQuery({ page: 1, page_size: 100 });

  const items: ChatContext[] = (data?.repositories ?? []).map((repo) => ({
    type: 'Repository',
    id: repo.id.toString(),
    label: repo.full_name,
    meta: {
      ...(repo.language && { Language: repo.language }),
      ...(repo.default_branch && { Branch: repo.default_branch }),
      Visibility: repo.private ? 'private' : 'public'
    }
  }));

  return {
    config: { type: 'Repository', icon: GitFork, labelKey: 'ai.context.repositories' },
    items,
    isLoading
  };
}

function useIntegrationsContextProvider(): ContextProviderData {
  const { data, isLoading } = useGetMCPServersQuery({ page: 1, limit: 100 });

  const items: ChatContext[] = (data?.items ?? [])
    .filter((s) => s.enabled)
    .map((server) => ({
      type: 'Integration',
      id: server.id,
      label: server.name,
      meta: {
        ID: server.id,
        Provider: server.provider_id
      }
    }));

  return {
    config: { type: 'Integration', icon: Plug, labelKey: 'ai.context.integrations' },
    items,
    isLoading
  };
}

export function useChatContextProviders(): ContextProviderData[] {
  const apps = useAppsContextProvider();
  const containers = useContainersContextProvider();
  const domains = useDomainsContextProvider();
  const repositories = useRepositoriesContextProvider();
  const integrations = useIntegrationsContextProvider();

  return [apps, containers, domains, repositories, integrations];
}
