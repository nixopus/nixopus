'use client';

import React from 'react';
import { ResourceGuard } from '@/packages/components/rbac';
import PageLayout from '@/packages/layouts/page-layout';

function DashboardPage() {
  return (
    <ResourceGuard resource="dashboard" action="read">
      <PageLayout maxWidth="full" padding="md" spacing="lg">
        hello world
      </PageLayout>
    </ResourceGuard>
  );
}

export default DashboardPage;
