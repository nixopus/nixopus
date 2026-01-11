'use client';
import React from 'react';
import { Mail, Slack, MessageSquare } from 'lucide-react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@/components/ui/form';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { PasswordInputField } from '@/components/ui/password-input-field';
import { Label } from '@/components/ui/label';
import {
  ChannelTabProps,
  NotificationPreferenceCardProps,
  NotificationPreferencesTabProps
} from '@/packages/types/settings';
import { useNotificationChannels } from '@/packages/hooks/settings/use-notification-channels';

const ChannelTab: React.FC<ChannelTabProps> = (props) => {
  const { t } = useTranslation();
  const {
    emailForm,
    slackForm,
    discordForm,
    onSubmitEmail,
    onSubmitSlack,
    onSubmitDiscord,
    isLoading
  } = useNotificationChannels(props);

  const { slackConfig, discordConfig } = props;

  const emailFields = [
    { name: 'smtp_host' as const, component: Input },
    { name: 'smtp_port' as const, component: Input },
    { name: 'smtp_username' as const, component: Input },
    { name: 'smtp_password' as const, component: PasswordInputField, type: 'password' as const },
    { name: 'smtp_from_email' as const, component: Input },
    { name: 'smtp_from_name' as const, component: Input }
  ];

  const channels = [
    {
      id: 'email',
      icon: Mail,
      titleKey: 'settings.notifications.channels.email.title' as const,
      descriptionKey: 'settings.notifications.channels.email.description' as const,
      renderChannel: () => (
        <div className="space-y-6">
          <div>
            <TypographySmall className="text-sm font-medium flex items-center gap-2">
              <Mail className="h-5 w-5" />
              {t('settings.notifications.channels.email.title')}
            </TypographySmall>
            <TypographyMuted className="text-xs mt-1">
              {t('settings.notifications.channels.email.description')}
            </TypographyMuted>
          </div>
          <div>
            <Form {...emailForm}>
              <form onSubmit={emailForm.handleSubmit(onSubmitEmail)} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {emailFields.map((field) => {
                    const FieldComponent = field.component;
                    return (
                      <FormField
                        key={field.name}
                        control={emailForm.control}
                        name={field.name}
                        render={({ field: formField }) => (
                          <FormItem>
                            <FormLabel>
                              {t(
                                `settings.notifications.channels.email.fields.${field.name}.label` as any
                              )}
                            </FormLabel>
                            <FormControl>
                              <FieldComponent
                                {...formField}
                                type={field.type}
                                placeholder={t(
                                  `settings.notifications.channels.email.fields.${field.name}.placeholder` as any
                                )}
                              />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    );
                  })}
                </div>
                <div className="flex justify-end">
                  <Button type="submit" disabled={isLoading}>
                    {t('settings.notifications.channels.save')}
                  </Button>
                </div>
              </form>
            </Form>
          </div>
        </div>
      )
    },
    {
      id: 'slack',
      icon: Slack,
      titleKey: 'settings.notifications.channels.slack.title' as const,
      descriptionKey: 'settings.notifications.channels.slack.description' as const,
      renderChannel: () => (
        <div className="space-y-6">
          <div>
            <TypographySmall className="text-sm font-medium flex items-center gap-2">
              <Slack className="h-5 w-5" />
              {t('settings.notifications.channels.slack.title')}
            </TypographySmall>
            <TypographyMuted className="text-xs mt-1">
              {t('settings.notifications.channels.slack.description')}
            </TypographyMuted>
          </div>
          <div>
            <Form {...slackForm}>
              <form onSubmit={slackForm.handleSubmit(onSubmitSlack)} className="space-y-4">
                <FormField
                  control={slackForm.control}
                  name="webhook_url"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {t('settings.notifications.channels.slack.fields.webhook_url.label')}
                      </FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder={t(
                            'settings.notifications.channels.slack.fields.webhook_url.placeholder'
                          )}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                {slackConfig?.webhook_url && (
                  <FormField
                    control={slackForm.control}
                    name="is_active"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">
                            {t('settings.notifications.channels.slack.fields.is_active.label')}
                          </FormLabel>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={(checked) => {
                              field.onChange(checked);
                              slackForm.setValue('is_active', checked);
                            }}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                )}
                <div className="flex justify-end">
                  <Button type="submit" disabled={isLoading}>
                    {t('settings.notifications.channels.save')}
                  </Button>
                </div>
              </form>
            </Form>
          </div>
        </div>
      )
    },
    {
      id: 'discord',
      icon: MessageSquare,
      titleKey: 'settings.notifications.channels.discord.title' as const,
      descriptionKey: 'settings.notifications.channels.discord.description' as const,
      renderChannel: () => (
        <div className="space-y-6">
          <div>
            <TypographySmall className="text-sm font-medium flex items-center gap-2">
              <MessageSquare className="h-5 w-5" />
              {t('settings.notifications.channels.discord.title')}
            </TypographySmall>
            <TypographyMuted className="text-xs mt-1">
              {t('settings.notifications.channels.discord.description')}
            </TypographyMuted>
          </div>
          <div>
            <Form {...discordForm}>
              <form onSubmit={discordForm.handleSubmit(onSubmitDiscord)} className="space-y-4">
                <FormField
                  control={discordForm.control}
                  name="webhook_url"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {t('settings.notifications.channels.discord.fields.webhook_url.label')}
                      </FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder={t(
                            'settings.notifications.channels.discord.fields.webhook_url.placeholder'
                          )}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                {discordConfig?.webhook_url && (
                  <FormField
                    control={discordForm.control}
                    name="is_active"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">
                            {t('settings.notifications.channels.discord.fields.is_active.label')}
                          </FormLabel>
                        </div>
                        <FormControl>
                          <Switch checked={field.value} onCheckedChange={field.onChange} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                )}
                <div className="flex justify-end">
                  <Button type="submit" disabled={isLoading}>
                    {t('settings.notifications.channels.save')}
                  </Button>
                </div>
              </form>
            </Form>
          </div>
        </div>
      )
    }
  ];

  return <div className="space-y-8">{channels.map((channel) => channel.renderChannel())}</div>;
};

export const NotificationChannelsTab = ChannelTab;

const NotificationPreferenceCard: React.FC<NotificationPreferenceCardProps> = ({
  title,
  description,
  preferences,
  onUpdate
}) => {
  return (
    <div className="space-y-4">
      <div>
        <TypographySmall className="text-sm font-medium">{title}</TypographySmall>
        <TypographyMuted className="text-xs mt-1">{description}</TypographyMuted>
      </div>
      <div className="space-y-4">
        {preferences?.map((pref) => (
          <div className="flex items-center justify-between" key={pref.id}>
            <div className="space-y-0.5">
              <Label htmlFor={pref.id} className="text-base">
                {pref.label}
              </Label>
              <TypographyMuted className="text-xs">{pref.description}</TypographyMuted>
            </div>
            <Switch
              id={pref.id}
              defaultChecked={pref.enabled}
              onCheckedChange={(enabled) => onUpdate?.(pref.id, enabled)}
            />
          </div>
        ))}
      </div>
    </div>
  );
};

export const NotificationPreferencesTab: React.FC<NotificationPreferencesTabProps> = ({
  activityPreferences,
  securityPreferences,
  onUpdatePreference
}) => {
  const { t } = useTranslation();

  return (
    <div className="grid gap-6 md:grid-cols-1">
      <NotificationPreferenceCard
        title={t('settings.notifications.preferences.activity.title')}
        description={t('settings.notifications.preferences.activity.description')}
        preferences={activityPreferences}
        onUpdate={onUpdatePreference}
      />

      <NotificationPreferenceCard
        title={t('settings.notifications.preferences.security.title')}
        description={t('settings.notifications.preferences.security.description')}
        preferences={securityPreferences}
        onUpdate={onUpdatePreference}
      />
    </div>
  );
};
