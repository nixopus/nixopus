'use client';

import { useTranslation } from '@/hooks/use-translation';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import useNotificationSettings from '@/app/settings/hooks/use-notification-settings';
import NotificationChannelsTab from '@/app/settings/notifications/components/channelTab';
import NotificationPreferencesTab from '@/app/settings/notifications/components/preferenceTab';
import { SMTPFormData } from '@/redux/types/notification';

export function NotificationsSettingsContent() {
  const { t } = useTranslation();
  const settings = useNotificationSettings();

  const handleSave = (data: SMTPFormData) => settings.handleOnSave(data);
  const handleSaveSlack = (data: Record<string, string>) => {
    settings.slackConfig
      ? settings.handleUpdateWebhookConfig({
          type: 'slack',
          webhook_url: data.webhook_url,
          is_active: data.is_active === 'true'
        })
      : settings.handleCreateWebhookConfig({ type: 'slack', webhook_url: data.webhook_url });
  };
  const handleSaveDiscord = (data: Record<string, string>) => {
    settings.discordConfig
      ? settings.handleUpdateWebhookConfig({
          type: 'discord',
          webhook_url: data.webhook_url,
          is_active: data.is_active === 'true'
        })
      : settings.handleCreateWebhookConfig({ type: 'discord', webhook_url: data.webhook_url });
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-semibold">{t('settings.notifications.page.title')}</h2>
      <Tabs defaultValue="channels" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
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
              smtpConfigs={settings.smtpConfigs || undefined}
              slackConfig={settings.slackConfig}
              discordConfig={settings.discordConfig}
              isLoading={settings.isLoading}
              handleOnSave={handleSave}
              handleOnSaveSlack={handleSaveSlack}
              handleOnSaveDiscord={handleSaveDiscord}
            />
          </TabsContent>
        </ResourceGuard>
        <TabsContent value="preferences">
          <NotificationPreferencesTab
            activityPreferences={settings.preferences?.activity}
            securityPreferences={settings.preferences?.security}
            onUpdatePreference={settings.handleUpdatePreference}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
