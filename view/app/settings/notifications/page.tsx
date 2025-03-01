'use client';
import React from 'react';
import DashboardPageHeader from '@/components/dashboard-page-header';
import NotificationPreferencesTab from './components/preferenceTab';
import NotificationChannelsTab from './components/channelTab';
import { activityPreferences, securityPreferences, updatePreferences } from './utils/preferences';

export type NotificationChannelConfig = {
  [key: string]: string;
};

const Page: React.FC = () => {
  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader
        label="Notifications"
        description="Manage your notification preferences and channels"
      />
      <NotificationChannelsTab />
      <NotificationPreferencesTab
        activityPreferences={activityPreferences}
        securityPreferences={securityPreferences}
        updatePreferences={updatePreferences}
      />
    </div>
  );
};

export default Page;
