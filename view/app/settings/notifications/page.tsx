'use client';
import React from 'react';
import DashboardPageHeader from '@/components/layout/dashboard-page-header';
import PageLayout from '@/components/layout/page-layout';
// TODO: Re-enable notifications when feature is working
// import NotificationPreferencesTab from './components/preferenceTab';
// import NotificationChannelsTab from './components/channelTab';
// import useNotificationSettings from '../hooks/use-notification-settings';
// import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
// import { useTranslation } from '@/hooks/use-translation';
// import { SMTPFormData } from '@/redux/types/notification';
// import { useFeatureFlags } from '@/hooks/features_provider';
// import Skeleton from '@/app/file-manager/components/skeleton/Skeleton';
// import DisabledFeature from '@/components/features/disabled-feature';
// import { FeatureNames } from '@/types/feature-flags';
// import { ResourceGuard } from '@/components/rbac/PermissionGuard';

export type NotificationChannelConfig = {
  [key: string]: string;
};

const Page: React.FC = () => {
  // TODO: Re-enable notifications when feature is working
  // Temporarily disabled - notifications feature is not working
  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <DashboardPageHeader
        label="Notifications"
        description="Notifications feature is temporarily disabled"
      />
      <div className="text-center py-8 text-muted-foreground">
        Notifications feature is currently disabled. This will be re-enabled in a future update.
      </div>
    </PageLayout>
  );

  /* COMMENTED OUT - Re-enable when notifications feature is working
  const { t } = useTranslation();
  const {
    smtpConfigs,
    slackConfig,
    discordConfig,
    isLoading,
    handleOnSave,
    handleCreateWebhookConfig,
    handleUpdateWebhookConfig,
    handleDeleteWebhookConfig,
    preferences,
    handleUpdatePreference
  } = useNotificationSettings();
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureNotifications)) {
    return <DisabledFeature />;
  }

  const handleSave = (data: SMTPFormData) => {
    if (smtpConfigs) {
      handleOnSave(data);
    } else {
      handleOnSave(data);
    }
  };

  const handleSaveSlack = (data: Record<string, string>) => {
    if (slackConfig) {
      handleUpdateWebhookConfig({
        type: 'slack',
        webhook_url: data.webhook_url,
        is_active: data.is_active === 'true'
      });
    } else {
      handleCreateWebhookConfig({
        type: 'slack',
        webhook_url: data.webhook_url
      });
    }
  };

  // TODO: Implement proper FeatureFlagRead permission management when this feature is taken up

  const handleSaveDiscord = (data: Record<string, string>) => {
    if (discordConfig) {
      handleUpdateWebhookConfig({
        type: 'discord',
        webhook_url: data.webhook_url,
        is_active: data.is_active === 'true'
      });
    } else {
      handleCreateWebhookConfig({
        type: 'discord',
        webhook_url: data.webhook_url
      });
    }
  };

  const handleUpdate = (id: string, enabled: boolean) => {
    handleUpdatePreference(id, enabled);
  };

  return (
    <ResourceGuard resource="notification" action="read">
      <PageLayout maxWidth="6xl" padding="md" spacing="lg">
        <DashboardPageHeader
          label={t('settings.notifications.page.title')}
          description={t('settings.notifications.page.description')}
        />
        <Tabs defaultValue="channels" className="w-full">
          <TabsList className={`grid w-full grid-cols-2`}>
            <ResourceGuard resource="notification" action="create">
              <TabsTrigger value="channels">
                {t('settings.notifications.page.tabs.channels')}
              </TabsTrigger>
            </ResourceGuard>
            <TabsTrigger value="preferences">
              {t('settings.notifications.page.tabs.preferences')}
            </TabsTrigger>
          </TabsList>
          <ResourceGuard resource="notification" action="create">
            <TabsContent value="channels">
              <NotificationChannelsTab
                smtpConfigs={smtpConfigs || undefined}
                slackConfig={slackConfig}
                discordConfig={discordConfig}
                isLoading={isLoading}
                handleOnSave={handleSave}
                handleOnSaveSlack={handleSaveSlack}
                handleOnSaveDiscord={handleSaveDiscord}
              />
            </TabsContent>
          </ResourceGuard>
          <TabsContent value="preferences">
            <NotificationPreferencesTab
              activityPreferences={preferences?.activity}
              securityPreferences={preferences?.security}
              onUpdatePreference={handleUpdate}
            />
          </TabsContent>
        </Tabs>
      </PageLayout>
    </ResourceGuard>
  );
  */
};

export default Page;
