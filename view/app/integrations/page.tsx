'use client';

import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { ResourceGuard } from '@/packages/components/rbac';
import { Skeleton } from '@nixopus/ui';
import { TypographyH2, TypographyMuted } from '@nixopus/ui';
import { IntegrationsPage } from '@/packages/components/integrations';
import { useFeatureFlags } from '@/packages/hooks/shared/features_provider';
import { FeatureNames } from '@/packages/types/feature-flags';

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
  const { isFeatureEnabled } = useFeatureFlags();
  if (!isFeatureEnabled(FeatureNames.FeatureNotifications)) {
    return <AccessDenied />;
  }
  return (
    <ResourceGuard
      resource="notification"
      action="read"
      loadingFallback={<Skeleton className="h-full w-full" />}
      fallback={<AccessDenied />}
    >
      <IntegrationsPage />
    </ResourceGuard>
  );
}
