'use client';
import React from 'react';
import ListRepositories from '../components/github-repositories/list-repositories';
import { useAppSelector } from '@/redux/hooks';
import { hasPermission } from '@/lib/permission';
import { useTranslation } from '@/hooks/use-translation';

function page() {
  const user = useAppSelector((state) => state.auth.user);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const canCreate = hasPermission(user, 'deploy', 'create', activeOrg?.id);
  const { t } = useTranslation();

  if (!canCreate) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold">{t('selfHost.create.accessDenied.title')}</h2>
          <p className="text-muted-foreground">{t('selfHost.create.accessDenied.description')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <ListRepositories />
    </div>
  );
}

export default page;
