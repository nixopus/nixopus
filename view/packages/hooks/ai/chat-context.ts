'use client';

import { Layers, Box, Globe, type LucideIcon } from 'lucide-react';
import { useGetApplicationsQuery } from '@/redux/services/deploy/applicationsApi';
import { useGetContainersQuery } from '@/redux/services/container/containerApi';
import { useGetAllDomainsQuery } from '@/redux/services/settings/domainsApi';

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

export function useChatContextProviders(): ContextProviderData[] {
  const apps = useAppsContextProvider();
  const containers = useContainersContextProvider();
  const domains = useDomainsContextProvider();

  return [apps, containers, domains];
}
