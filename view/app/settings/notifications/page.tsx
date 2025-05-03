'use client';
import React from 'react';
import DashboardPageHeader from '@/components/layout/dashboard-page-header';
import NotificationPreferencesTab from './components/preferenceTab';
import NotificationChannelsTab from './components/channelTab';
import useNotificationSettings from '../hooks/use-notification-settings';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useAppSelector } from '@/redux/hooks';
import { hasPermission } from '@/lib/permission';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';
import { SMTPFormData } from '@/redux/types/notification';
import { useFeatureFlags } from '@/hooks/features_provider';
import Skeleton from '@/app/file-manager/components/skeleton/Skeleton';
import DisabledFeature from '@/components/features/disabled-feature';
import { FeatureNames } from '@/types/feature-flags';
export type NotificationChannelConfig = {
  [key: string]: string;
};

const Page: React.FC = () => {
  const { t } = useTranslation();
  const user = useAppSelector((state) => state.auth.user);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
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
  const canRead = hasPermission(user, 'notification', 'read', activeOrg?.id);
  const canUpdate = hasPermission(user, 'notification', 'update', activeOrg?.id);
  const canCreate = hasPermission(user, 'notification', 'create', activeOrg?.id);

  if (!canRead) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold">
            {t('settings.notifications.page.accessDenied.title')}
          </h2>
          <p className="text-muted-foreground">
            {t('settings.notifications.page.accessDenied.description')}
          </p>
        </div>
      </div>
    );
  }

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureNotifications)) {
    return <DisabledFeature />;
  }
  
  const handleSave = (data: SMTPFormData) => {
    if (smtpConfigs) {
      if (canUpdate) {
        handleOnSave(data);
      } else {
        toast.error(t('settings.notifications.page.permissions.update'));
      }
    } else {
      if (canCreate) {
        handleOnSave(data);
      } else {
        toast.error(t('settings.notifications.page.permissions.create'));
      }
    }
  };

  const handleSaveSlack = (data: Record<string, string>) => {
    if (slackConfig) {
      if (canUpdate) {
        handleUpdateWebhookConfig({
          type: 'slack',
          webhook_url: data.webhook_url,
          is_active: data.is_active === 'true'
        });
      } else {
        toast.error(t('settings.notifications.page.permissions.update'));
      }
    } else {
      if (canCreate) {
        handleCreateWebhookConfig({
          type: 'slack',
          webhook_url: data.webhook_url
        });
      } else {
        toast.error(t('settings.notifications.page.permissions.create'));
      }
    }
  };

  const handleSaveDiscord = (data: Record<string, string>) => {
    if (discordConfig) {
      if (canUpdate) {
        handleUpdateWebhookConfig({
          type: 'discord',
          webhook_url: data.webhook_url,
          is_active: data.is_active === 'true'
        });
      } else {
        toast.error(t('settings.notifications.page.permissions.update'));
      }
    } else {
      if (canCreate) {
        handleCreateWebhookConfig({
          type: 'discord',
          webhook_url: data.webhook_url
        });
      } else {
        toast.error(t('settings.notifications.page.permissions.create'));
      }
    }
  };

  const handleUpdate = (id: string, enabled: boolean) => {
    handleUpdatePreference(id, enabled);
  };

  const showChannelsTab =
    canCreate ||
    (smtpConfigs && canUpdate) ||
    (slackConfig && canUpdate) ||
    (discordConfig && canUpdate);

  if (!showChannelsTab) {
    return (
      <div className="container mx-auto py-6 space-y-8 max-w-4xl">
        <DashboardPageHeader
          label={t('settings.notifications.page.title')}
          description={t('settings.notifications.page.description')}
        />
        <NotificationPreferencesTab
          activityPreferences={preferences?.activity}
          securityPreferences={preferences?.security}
          updatePreferences={preferences?.update}
          onUpdatePreference={handleUpdate}
        />
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader
        label={t('settings.notifications.page.title')}
        description={t('settings.notifications.page.description')}
      />
      <Tabs defaultValue="channels" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="channels">
            {t('settings.notifications.page.tabs.channels')}
          </TabsTrigger>
          <TabsTrigger value="preferences">
            {t('settings.notifications.page.tabs.preferences')}
          </TabsTrigger>
        </TabsList>
        <TabsContent value="channels">
          <NotificationChannelsTab
            smtpConfigs={smtpConfigs}
            slackConfig={slackConfig}
            discordConfig={discordConfig}
            isLoading={isLoading}
            handleOnSave={handleSave}
            handleOnSaveSlack={handleSaveSlack}
            handleOnSaveDiscord={handleSaveDiscord}
          />
        </TabsContent>
        <TabsContent value="preferences">
          <NotificationPreferencesTab
            activityPreferences={preferences?.activity}
            securityPreferences={preferences?.security}
            updatePreferences={preferences?.update}
            onUpdatePreference={handleUpdate}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default Page;
