'use client';

import { useState } from 'react';
import { Mail, Slack, Hash } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useAppSelector } from '@/redux/hooks';
import {
  useGetSMTPConfigurationsQuery,
  useGetWebhookConfigQuery
} from '@/redux/services/settings/notificationApi';
import useNotificationSettings from '@/packages/hooks/settings/use-notification-settings';
import type { SMTPConfig, WebhookConfig } from '@/redux/types/notification';
import type { FetchBaseQueryError } from '@reduxjs/toolkit/query/react';

export type IntegrationId = 'smtp' | 'slack' | 'discord';
export type IntegrationCategory = 'email' | 'messaging';

export interface IntegrationDefinition {
  id: IntegrationId;
  nameKey: string;
  category: IntegrationCategory;
  icon: LucideIcon;
}

export const INTEGRATIONS: IntegrationDefinition[] = [
  { id: 'smtp', nameKey: 'integrations.smtp.name', category: 'email', icon: Mail },
  { id: 'slack', nameKey: 'integrations.slack.name', category: 'messaging', icon: Slack },
  { id: 'discord', nameKey: 'integrations.discord.name', category: 'messaging', icon: Hash }
];

const is404 = (err: unknown): boolean =>
  err != null &&
  typeof err === 'object' &&
  'status' in err &&
  (err as FetchBaseQueryError).status === 404;

export function useIntegrations() {
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const skip = !activeOrganization?.id;

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

  const notificationSettings = useNotificationSettings();

  const [selectedIntegration, setSelectedIntegration] = useState<IntegrationDefinition | null>(
    null
  );

  const openModal = (integration: IntegrationDefinition) => setSelectedIntegration(integration);
  const closeModal = () => setSelectedIntegration(null);

  const smtpError = is404(smtpRawError) ? null : (smtpRawError ?? null);
  const slackError = is404(slackRawError) ? null : (slackRawError ?? null);
  const discordError = is404(discordRawError) ? null : (discordRawError ?? null);

  const isLoading = smtpLoading || slackLoading || discordLoading;

  const getConfigForIntegration = (id: IntegrationId): SMTPConfig | WebhookConfig | null => {
    if (id === 'smtp') return smtpConfig ?? null;
    if (id === 'slack') return slackConfig ?? null;
    if (id === 'discord') return discordConfig ?? null;
    return null;
  };

  return {
    integrations: INTEGRATIONS,
    smtpConfig: smtpConfig ?? null,
    slackConfig: slackConfig ?? null,
    discordConfig: discordConfig ?? null,
    isLoading,
    smtpError,
    slackError,
    discordError,
    selectedIntegration,
    openModal,
    closeModal,
    getConfigForIntegration,
    handleOnSave: notificationSettings.handleOnSave,
    handleCreateWebhookConfig: notificationSettings.handleCreateWebhookConfig,
    handleUpdateWebhookConfig: notificationSettings.handleUpdateWebhookConfig,
    handleDeleteWebhookConfig: notificationSettings.handleDeleteWebhookConfig,
    handleDeleteSMTPConfiguration: notificationSettings.handleDeleteSMTPConfiguration
  };
}
