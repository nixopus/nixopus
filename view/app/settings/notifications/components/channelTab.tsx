'use client';
import React, { useEffect } from 'react';
import { Mail, Slack, MessageSquare } from 'lucide-react';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';
import { SMTPConfig, WebhookConfig, SMTPFormData } from '@/redux/types/notification';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
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
import { toast } from 'sonner';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { PasswordInputField } from '@/components/ui/password-input-field';

interface ChannelTabProps {
  smtpConfigs?: SMTPConfig;
  slackConfig?: WebhookConfig;
  discordConfig?: WebhookConfig;
  isLoading: boolean;
  handleOnSave: (data: SMTPFormData) => void;
  handleOnSaveSlack: (data: Record<string, string>) => void;
  handleOnSaveDiscord: (data: Record<string, string>) => void;
}

const emailFormSchema = z.object({
  smtp_host: z.string().min(1, 'SMTP host is required'),
  smtp_port: z.string().min(1, 'SMTP port is required'),
  smtp_username: z.string().min(1, 'SMTP username is required'),
  smtp_password: z.string().min(1, 'SMTP password is required'),
  smtp_from_email: z.string().email('Invalid email address'),
  smtp_from_name: z.string().min(1, 'From name is required')
});

const slackFormSchema = z.object({
  webhook_url: z.string().url('Invalid webhook URL'),
  is_active: z.boolean().default(true)
});

const discordFormSchema = z.object({
  webhook_url: z.string().url('Invalid webhook URL'),
  is_active: z.boolean().default(true)
});

const ChannelTab: React.FC<ChannelTabProps> = ({
  smtpConfigs,
  slackConfig,
  discordConfig,
  isLoading,
  handleOnSave,
  handleOnSaveSlack,
  handleOnSaveDiscord
}) => {
  const { t } = useTranslation();

  const emailForm = useForm<z.infer<typeof emailFormSchema>>({
    resolver: zodResolver(emailFormSchema),
    defaultValues: {
      smtp_host: smtpConfigs?.host || '',
      smtp_port: smtpConfigs?.port?.toString() || '',
      smtp_username: smtpConfigs?.username || '',
      smtp_password: smtpConfigs?.password || '',
      smtp_from_email: smtpConfigs?.from_email || '',
      smtp_from_name: smtpConfigs?.from_name || ''
    }
  });

  const slackForm = useForm<z.infer<typeof slackFormSchema>>({
    resolver: zodResolver(slackFormSchema),
    defaultValues: {
      webhook_url: slackConfig?.webhook_url || '',
      is_active: slackConfig?.is_active ?? true
    }
  });

  const discordForm = useForm<z.infer<typeof discordFormSchema>>({
    resolver: zodResolver(discordFormSchema),
    defaultValues: {
      webhook_url: discordConfig?.webhook_url || '',
      is_active: discordConfig?.is_active ?? true
    }
  });

  useEffect(() => {
    if (slackConfig) {
      slackForm.setValue('webhook_url', slackConfig.webhook_url || '');
      slackForm.setValue('is_active', slackConfig.is_active);
    }
  }, [slackConfig]);

  useEffect(() => {
    if (discordConfig) {
      discordForm.setValue('webhook_url', discordConfig.webhook_url || '');
      discordForm.setValue('is_active', discordConfig.is_active);
    }
  }, [discordConfig]);

  useEffect(() => {
    if (smtpConfigs) {
      emailForm.setValue('smtp_host', smtpConfigs.host || '');
      emailForm.setValue('smtp_port', smtpConfigs.port?.toString() || '');
      emailForm.setValue('smtp_username', smtpConfigs.username || '');
      emailForm.setValue('smtp_password', smtpConfigs.password || '');
      emailForm.setValue('smtp_from_email', smtpConfigs.from_email || '');
      emailForm.setValue('smtp_from_name', smtpConfigs.from_name || '');
    }
  }, [smtpConfigs]);

  const onSubmitEmail = async (data: SMTPFormData) => {
    try {
      await handleOnSave(data);
      toast.success(t('settings.notifications.messages.email.success'));
    } catch (error) {
      toast.error(t('settings.notifications.messages.email.error'));
    }
  };

  const onSubmitSlack = async (data: z.infer<typeof slackFormSchema>) => {
    try {
      await handleOnSaveSlack({
        webhook_url: data.webhook_url,
        is_active: data.is_active.toString()
      });
      toast.success(t('settings.notifications.messages.slack.success'));
    } catch (error) {
      toast.error(t('settings.notifications.messages.slack.error'));
    }
  };

  const onSubmitDiscord = async (data: z.infer<typeof discordFormSchema>) => {
    try {
      await handleOnSaveDiscord({
        webhook_url: data.webhook_url,
        is_active: data.is_active.toString()
      });
      toast.success(t('settings.notifications.messages.discord.success'));
    } catch (error) {
      toast.error(t('settings.notifications.messages.discord.error'));
    }
  };

  return (
    <div className="grid gap-6 md:grid-cols-1">
      <Card>
        <CardHeader>
          <TypographySmall className="flex items-center gap-2">
            <Mail className="h-5 w-5" />
            {t('settings.notifications.channels.email.title')}
          </TypographySmall>
          <TypographyMuted>
            {t('settings.notifications.channels.email.description')}
          </TypographyMuted>
        </CardHeader>
        <CardContent>
          <Form {...emailForm}>
            <form onSubmit={emailForm.handleSubmit(onSubmitEmail)} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <FormField
                  control={emailForm.control}
                  name="smtp_host"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {t('settings.notifications.channels.email.fields.smtp_host.label')}
                      </FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder={t(
                            'settings.notifications.channels.email.fields.smtp_host.placeholder'
                          )}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={emailForm.control}
                  name="smtp_port"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {t('settings.notifications.channels.email.fields.smtp_port.label')}
                      </FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder={t(
                            'settings.notifications.channels.email.fields.smtp_port.placeholder'
                          )}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={emailForm.control}
                  name="smtp_username"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {t('settings.notifications.channels.email.fields.smtp_username.label')}
                      </FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder={t(
                            'settings.notifications.channels.email.fields.smtp_username.placeholder'
                          )}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={emailForm.control}
                  name="smtp_password"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {t('settings.notifications.channels.email.fields.smtp_password.label')}
                      </FormLabel>
                      <FormControl>
                        <PasswordInputField
                          type="password"
                          {...field}
                          placeholder={t(
                            'settings.notifications.channels.email.fields.smtp_password.placeholder'
                          )}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={emailForm.control}
                  name="smtp_from_email"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {t('settings.notifications.channels.email.fields.smtp_from_email.label')}
                      </FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder={t(
                            'settings.notifications.channels.email.fields.smtp_from_email.placeholder'
                          )}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={emailForm.control}
                  name="smtp_from_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {t('settings.notifications.channels.email.fields.smtp_from_name.label')}
                      </FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder={t(
                            'settings.notifications.channels.email.fields.smtp_from_name.placeholder'
                          )}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <div className="flex justify-end">
                <Button type="submit" disabled={isLoading}>
                  {t('settings.notifications.channels.save')}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <TypographySmall className="flex items-center gap-2">
            <Slack className="h-5 w-5" />
            {t('settings.notifications.channels.slack.title')}
          </TypographySmall>
          <TypographyMuted>
            {t('settings.notifications.channels.slack.description')}
          </TypographyMuted>
        </CardHeader>
        <CardContent>
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
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <TypographySmall className="flex items-center gap-2">
            <MessageSquare className="h-5 w-5" />
            {t('settings.notifications.channels.discord.title')}
          </TypographySmall>
          <TypographyMuted>
            {t('settings.notifications.channels.discord.description')}
          </TypographyMuted>
        </CardHeader>
        <CardContent>
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
        </CardContent>
      </Card>
    </div>
  );
};

export default ChannelTab;
