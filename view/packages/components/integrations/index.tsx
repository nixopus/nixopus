'use client';

import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useRBAC } from '@/packages/utils/rbac';
import { MainPageHeader } from '@nixopus/ui';
import PageLayout from '@/packages/layouts/page-layout';
import { IntegrationCard } from './integration-card';
import { IntegrationConfigModal } from './integration-config-modal';
import { useIntegrations } from '@/packages/hooks/integrations/use-integrations';
import type { IntegrationId } from '@/packages/hooks/integrations/use-integrations';
import type { SMTPFormData } from '@/redux/types/notification';

export function IntegrationsPage() {
  const { t } = useTranslation();
  const { canAccessResource } = useRBAC();
  const {
    integrations,
    isLoading,
    selectedIntegration,
    openModal,
    closeModal,
    getConfigForIntegration,
    handleOnSave,
    handleCreateWebhookConfig,
    handleUpdateWebhookConfig,
    handleDeleteWebhookConfig,
    handleDeleteSMTPConfiguration,
    slackConfig,
    discordConfig
  } = useIntegrations();

  const canConfigure =
    canAccessResource('notification', 'create') || canAccessResource('notification', 'update');
  const canDelete = canAccessResource('notification', 'delete');

  const handleSaveWebhook = async (
    data: { webhook_url: string; is_active: boolean },
    type: 'slack' | 'discord'
  ) => {
    const existing = type === 'slack' ? slackConfig : discordConfig;
    if (existing) {
      await handleUpdateWebhookConfig({
        type,
        webhook_url: data.webhook_url,
        is_active: data.is_active
      });
    } else {
      await handleCreateWebhookConfig({ type, webhook_url: data.webhook_url });
    }
  };

  return (
    <PageLayout maxWidth="full" padding="md" spacing="lg">
      <MainPageHeader label={t('integrations.title' as any)} highlightLabel={false} />
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {integrations.map((integration) => (
          <IntegrationCard
            key={integration.id}
            integration={integration}
            config={getConfigForIntegration(integration.id as IntegrationId)}
            isLoading={isLoading}
            onConfigure={openModal}
            canConfigure={canConfigure}
          />
        ))}
      </div>

      {selectedIntegration && (
        <IntegrationConfigModal
          integration={selectedIntegration}
          config={getConfigForIntegration(selectedIntegration.id as IntegrationId)}
          onClose={closeModal}
          onSaveSMTP={async (data: SMTPFormData) => {
            await handleOnSave(data);
          }}
          onSaveWebhook={handleSaveWebhook}
          onDeleteSMTP={handleDeleteSMTPConfiguration}
          onDeleteWebhook={handleDeleteWebhookConfig}
          canDelete={canDelete}
          isLoading={isLoading}
        />
      )}
    </PageLayout>
  );
}
