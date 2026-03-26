'use client';

import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { ResourceGuard } from '@/packages/components/rbac';
import { TypographyH2, TypographyMuted } from '@nixopus/ui';
import { IntegrationsPage } from '@/packages/components/integrations';

function AccessDenied() {
  const { t } = useTranslation();
  return (
    <div className="flex h-full items-center justify-center">
      <div className="text-center">
        <TypographyH2>{t('integrations.accessDenied.title' as any)}</TypographyH2>
        <TypographyMuted>{t('integrations.accessDenied.description' as any)}</TypographyMuted>
      </div>
    </div>
  );
}

export default function Page() {
  return (
    <ResourceGuard
      resource="notification"
      action="read"
      loadingFallback={null}
      fallback={<AccessDenied />}
    >
      <IntegrationsPage />
    </ResourceGuard>
  );
}
