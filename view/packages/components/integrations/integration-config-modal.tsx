'use client';

import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@nixopus/ui';
import { SmtpConfigForm } from './smtp-config-form';
import { SlackConfigForm } from './slack-config-form';
import { DiscordConfigForm } from './discord-config-form';
import type { IntegrationDefinition } from '@/packages/hooks/integrations/use-integrations';
import type { SMTPConfig, WebhookConfig, SMTPFormData } from '@/redux/types/notification';

interface IntegrationConfigModalProps {
  integration: IntegrationDefinition;
  config: SMTPConfig | WebhookConfig | null;
  onClose: () => void;
  onSaveSMTP: (data: SMTPFormData) => Promise<void>;
  onSaveWebhook: (
    data: { webhook_url: string; is_active: boolean },
    type: 'slack' | 'discord'
  ) => Promise<void>;
  onDeleteSMTP: (id: string) => Promise<void>;
  onDeleteWebhook: (type: string) => Promise<void>;
  canDelete: boolean;
  isLoading?: boolean;
}

export function IntegrationConfigModal({
  integration,
  config,
  onClose,
  onSaveSMTP,
  onSaveWebhook,
  onDeleteSMTP,
  onDeleteWebhook,
  canDelete,
  isLoading
}: IntegrationConfigModalProps) {
  const { t } = useTranslation();

  const renderForm = () => {
    switch (integration.id) {
      case 'smtp':
        return (
          <SmtpConfigForm
            config={config as SMTPConfig | null}
            onSave={onSaveSMTP}
            onDelete={onDeleteSMTP}
            onClose={onClose}
            canDelete={canDelete}
            isLoading={isLoading}
          />
        );
      case 'slack':
        return (
          <SlackConfigForm
            config={config as WebhookConfig | null}
            onSave={(data) => onSaveWebhook(data, 'slack')}
            onDelete={onDeleteWebhook}
            onClose={onClose}
            canDelete={canDelete}
            isLoading={isLoading}
          />
        );
      case 'discord':
        return (
          <DiscordConfigForm
            config={config as WebhookConfig | null}
            onSave={(data) => onSaveWebhook(data, 'discord')}
            onDelete={onDeleteWebhook}
            onClose={onClose}
            canDelete={canDelete}
            isLoading={isLoading}
          />
        );
    }
  };

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{t(integration.nameKey as any)}</DialogTitle>
        </DialogHeader>
        {renderForm()}
      </DialogContent>
    </Dialog>
  );
}
