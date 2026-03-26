'use client';

import { useState } from 'react';
import type React from 'react';
import { Mail } from 'lucide-react';
import { SlackIcon } from '@/packages/components/icons/slack-icon';
import { DiscordIcon } from '@/packages/components/icons/discord-icon';
import { useAppSelector } from '@/redux/hooks';
import {
  useGetSMTPConfigurationsQuery,
  useGetWebhookConfigQuery
} from '@/redux/services/settings/notificationApi';
import { useGetMCPServersQuery, useGetMCPCatalogQuery } from '@/redux/services/settings/mcpApi';
import useNotificationSettings from '@/packages/hooks/settings/use-notification-settings';
import type { SMTPConfig, WebhookConfig } from '@/redux/types/notification';
import type { MCPProvider, MCPServer } from '@/redux/types/mcp';
import type { FetchBaseQueryError } from '@reduxjs/toolkit/query/react';

export type IntegrationId = 'smtp' | 'slack' | 'discord';
export type IntegrationCategory = 'email' | 'messaging' | 'tools';

export interface IntegrationDefinition {
  id: IntegrationId;
  nameKey: string;
  name: string;
  category: IntegrationCategory;
  icon: React.ComponentType<{ className?: string }>;
}

export const INTEGRATIONS: IntegrationDefinition[] = [
  { id: 'smtp', name: 'SMTP', nameKey: 'integrations.smtp.name', category: 'email', icon: Mail },
  {
    id: 'slack',
    name: 'Slack',
    nameKey: 'integrations.slack.name',
    category: 'messaging',
    icon: SlackIcon
  },
  {
    id: 'discord',
    name: 'Discord',
    nameKey: 'integrations.discord.name',
    category: 'messaging',
    icon: DiscordIcon
  }
];

const is404 = (err: unknown): boolean =>
  err != null &&
  typeof err === 'object' &&
  'status' in err &&
  (err as FetchBaseQueryError).status === 404;

export function useIntegrations() {
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const skip = !activeOrganization?.id;

  // Search / sort / pagination state — sent to backend
  const [search, setSearch] = useState('');
  const [sortBy, setSortBy] = useState('name');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');
  const [catalogPage, setCatalogPage] = useState(1);
  const [serversPage, setServersPage] = useState(1);
  const PAGE_SIZE = 12;

  const {
    data: smtpConfig,
    isLoading: smtpLoading,
    error: smtpRawError
  } = useGetSMTPConfigurationsQuery(activeOrganization?.id || '', { skip });

  const {
    data: slackConfig,
    isLoading: slackLoading,
    error: slackRawError
  } = useGetWebhookConfigQuery({ type: 'slack' }, { skip });

  const {
    data: discordConfig,
    isLoading: discordLoading,
    error: discordRawError
  } = useGetWebhookConfigQuery({ type: 'discord' }, { skip });

  const { data: mcpServersResult, isLoading: mcpLoading } = useGetMCPServersQuery(
    { q: search, sortBy, sortDir, page: serversPage, limit: PAGE_SIZE },
    { skip }
  );

  const { data: mcpCatalogResult, isLoading: mcpCatalogLoading } = useGetMCPCatalogQuery(
    { q: search, sortBy, sortDir, page: catalogPage, limit: PAGE_SIZE },
    { skip }
  );

  const notificationSettings = useNotificationSettings();

  const [selectedIntegration, setSelectedIntegration] = useState<IntegrationDefinition | null>(
    null
  );

  const openModal = (integration: IntegrationDefinition) => setSelectedIntegration(integration);
  const closeModal = () => setSelectedIntegration(null);

  const smtpError = is404(smtpRawError) ? null : (smtpRawError ?? null);
  const slackError = is404(slackRawError) ? null : (slackRawError ?? null);
  const discordError = is404(discordRawError) ? null : (discordRawError ?? null);

  const isLoading =
    smtpLoading ||
    slackLoading ||
    discordLoading ||
    mcpLoading ||
    mcpCatalogLoading ||
    notificationSettings.isLoading;

  const getConfigForIntegration = (id: IntegrationId): SMTPConfig | WebhookConfig | null => {
    if (id === 'smtp') return smtpConfig ?? null;
    if (id === 'slack') return slackConfig ?? null;
    if (id === 'discord') return discordConfig ?? null;
    return null;
  };

  const mcpServers: MCPServer[] = mcpServersResult?.items ?? [];
  const mcpCatalog: MCPProvider[] = mcpCatalogResult?.items ?? [];
  const catalogTotalCount = mcpCatalogResult?.totalCount ?? 0;
  const serversTotalCount = mcpServersResult?.totalCount ?? 0;

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
    setCatalogPage(1);
    setServersPage(1);
  };

  const handleSortChange = (newSortBy: string, newSortDir: 'asc' | 'desc') => {
    setSortBy(newSortBy);
    setSortDir(newSortDir);
    setCatalogPage(1);
    setServersPage(1);
  };

  return {
    integrations: INTEGRATIONS,
    smtpConfig: smtpConfig ?? null,
    slackConfig: slackConfig ?? null,
    discordConfig: discordConfig ?? null,
    mcpServers,
    mcpCatalog,
    catalogTotalCount,
    serversTotalCount,
    catalogPage,
    setCatalogPage,
    catalogPageSize: PAGE_SIZE,
    isLoading,
    smtpError,
    slackError,
    discordError,
    selectedIntegration,
    openModal,
    closeModal,
    getConfigForIntegration,
    search,
    sortBy,
    sortDir,
    handleSearchChange,
    handleSortChange,
    handleOnSave: notificationSettings.handleOnSave,
    handleCreateWebhookConfig: notificationSettings.handleCreateWebhookConfig,
    handleUpdateWebhookConfig: notificationSettings.handleUpdateWebhookConfig,
    handleDeleteWebhookConfig: notificationSettings.handleDeleteWebhookConfig,
    handleDeleteSMTPConfiguration: notificationSettings.handleDeleteSMTPConfiguration
  };
}
