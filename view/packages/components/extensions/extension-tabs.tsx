'use client';

import { Skeleton } from '@nixopus/ui';
import { BadgeGroup, BadgeGroupItem } from '@nixopus/ui';
import { DataTable } from '@nixopus/ui';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import React from 'react';
import { OverviewTabProps } from '@/packages/types/extension';

export function OverviewTab({ extension, isLoading, variableColumns }: OverviewTabProps) {
  const { t } = useTranslation();

  if (isLoading) {
    return <Skeleton className="h-40 w-full" />;
  }

  return (
    <div className="space-y-6">
      <BadgeGroup>
        {extension?.category && (
          <BadgeGroupItem variant="secondary">{extension.category}</BadgeGroupItem>
        )}
        {extension?.extension_type && (
          <BadgeGroupItem variant="outline">{extension.extension_type}</BadgeGroupItem>
        )}
        {extension?.version && <BadgeGroupItem>v{extension.version}</BadgeGroupItem>}
        {extension?.is_verified && <BadgeGroupItem>Verified</BadgeGroupItem>}
      </BadgeGroup>

      {extension?.variables && extension.variables.length > 0 && variableColumns && (
        <DataTable
          data={extension.variables}
          columns={variableColumns}
          showBorder={true}
          striped={false}
        />
      )}
    </div>
  );
}
