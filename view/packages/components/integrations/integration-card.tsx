'use client';

import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Button, Badge, Skeleton, Card, CardContent, CardHeader } from '@nixopus/ui';
import type { IntegrationDefinition } from '@/packages/hooks/integrations/use-integrations';
import type { SMTPConfig, WebhookConfig } from '@/redux/types/notification';

interface IntegrationCardProps {
  integration: IntegrationDefinition;
  config: SMTPConfig | WebhookConfig | null;
  isLoading: boolean;
  onConfigure: (integration: IntegrationDefinition) => void;
  canConfigure: boolean;
}

function getStatusInfo(config: SMTPConfig | WebhookConfig | null) {
  if (!config) return { key: 'integrations.status.notSetUp', variant: 'secondary' as const };
  if (config.is_active)
    return { key: 'integrations.status.connected', variant: 'default' as const };
  return { key: 'integrations.status.configured', variant: 'outline' as const };
}

export function IntegrationCard({
  integration,
  config,
  isLoading,
  onConfigure,
  canConfigure
}: IntegrationCardProps) {
  const { t } = useTranslation();
  const Icon = integration.icon;
  const status = getStatusInfo(config);

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center gap-3 pb-2">
          <Skeleton className="h-8 w-8 rounded-md" />
          <div className="space-y-1 flex-1">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-3 w-16" />
          </div>
        </CardHeader>
        <CardContent className="flex items-center justify-between pt-2">
          <Skeleton className="h-5 w-20 rounded-full" />
          <Skeleton className="h-8 w-24 rounded-md" />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center gap-3 pb-2">
        <div className="flex h-8 w-8 items-center justify-center rounded-md bg-muted">
          <Icon className="h-4 w-4" />
        </div>
        <div>
          <p className="text-sm font-medium leading-none">{t(integration.nameKey as any)}</p>
          <p className="text-xs text-muted-foreground mt-0.5">
            {t(`integrations.categories.${integration.category}` as any)}
          </p>
        </div>
      </CardHeader>
      <CardContent className="flex items-center justify-between pt-2">
        <Badge variant={status.variant}>{t(status.key as any)}</Badge>
        {canConfigure && (
          <Button variant="outline" size="sm" onClick={() => onConfigure(integration)}>
            {t('integrations.configure' as any)}
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
