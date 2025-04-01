'use client';
import React from 'react';
import DashboardPageHeader from '@/components/layout/dashboard-page-header';

function Page() {
  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader
        label="Server Settings"
        description="Manage your servers and their configurations"
      />
    </div>
  );
}

export default Page;
