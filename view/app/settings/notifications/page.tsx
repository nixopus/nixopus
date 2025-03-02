'use client';
import React from 'react';
import DashboardPageHeader from '@/components/dashboard-page-header';
import NotificationPreferencesTab from './components/preferenceTab';
import NotificationChannelsTab from './components/channelTab';
import useNotificationSettings from '../hooks/use-notification-settings';

export type NotificationChannelConfig = {
  [key: string]: string;
};

const Page: React.FC = () => {
  const { smtpConfigs, isLoading, handleOnSave, preferences, handleUpdatePreference } =
    useNotificationSettings();
  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader
        label="Notifications"
        description="Manage your notification preferences and channels"
      />
      <NotificationChannelsTab
        smtpConfigs={smtpConfigs}
        isLoading={isLoading}
        handleOnSave={handleOnSave}
      />
      <NotificationPreferencesTab
        activityPreferences={preferences?.activity}
        securityPreferences={preferences?.security}
        updatePreferences={preferences?.update}
        onUpdatePreference={handleUpdatePreference}
      />
    </div>
  );
};

export default Page;
