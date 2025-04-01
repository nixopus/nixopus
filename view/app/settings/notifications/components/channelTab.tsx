'use client';
import React, { useEffect } from 'react';
import { Mail, Slack, MessageSquare } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';
import { SMTPConfig, WebhookConfig, SMTPFormData } from '@/redux/types/notification';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@/components/ui/form';
import { toast } from 'sonner';

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
  webhook_secret: z.string().min(1, 'Webhook secret is required'),
  channel_id: z.string().min(1, 'Channel ID is required')
});

const discordFormSchema = z.object({
  webhook_url: z.string().url('Invalid webhook URL')
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
      webhook_secret: slackConfig?.webhook_secret || '',
      channel_id: slackConfig?.channel_id || ''
    }
  });

  const discordForm = useForm<z.infer<typeof discordFormSchema>>({
    resolver: zodResolver(discordFormSchema),
    defaultValues: {
      webhook_url: discordConfig?.webhook_url || ''
    }
  });

  useEffect(() => {
    if (slackConfig) {
      slackForm.setValue('webhook_secret', slackConfig.webhook_secret || '');
      slackForm.setValue('channel_id', slackConfig.channel_id || '');
    }
  }, [slackConfig]);

  useEffect(() => {
    if (discordConfig) {
      discordForm.setValue('webhook_url', discordConfig.webhook_url || '');
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
      await handleOnSaveSlack(data);
      toast.success(t('settings.notifications.messages.slack.success'));
    } catch (error) {
      toast.error(t('settings.notifications.messages.slack.error'));
    }
  };

  const onSubmitDiscord = async (data: z.infer<typeof discordFormSchema>) => {
    try {
      await handleOnSaveDiscord(data);
      toast.success(t('settings.notifications.messages.discord.success'));
    } catch (error) {
      toast.error(t('settings.notifications.messages.discord.error'));
    }
  };

  return (
    <div className="grid gap-6 md:grid-cols-1">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Mail className="h-5 w-5" />
            {t('settings.notifications.channels.email.title')}
          </CardTitle>
          <CardDescription>
            {t('settings.notifications.channels.email.description')}
          </CardDescription>
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
                        <Input
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
          <CardTitle className="flex items-center gap-2">
            <Slack className="h-5 w-5" />
            {t('settings.notifications.channels.slack.title')}
          </CardTitle>
          <CardDescription>
            {t('settings.notifications.channels.slack.description')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...slackForm}>
            <form onSubmit={slackForm.handleSubmit(onSubmitSlack)} className="space-y-4">
              <FormField
                control={slackForm.control}
                name="webhook_secret"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {t('settings.notifications.channels.slack.fields.webhook_secret.label')}
                    </FormLabel>
                    <FormControl>
                      <Input
                        type="password"
                        {...field}
                        placeholder={t(
                          'settings.notifications.channels.slack.fields.webhook_secret.placeholder'
                        )}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={slackForm.control}
                name="channel_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {t('settings.notifications.channels.slack.fields.channel_id.label')}
                    </FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder={t(
                          'settings.notifications.channels.slack.fields.channel_id.placeholder'
                        )}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
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
          <CardTitle className="flex items-center gap-2">
            <MessageSquare className="h-5 w-5" />
            {t('settings.notifications.channels.discord.title')}
          </CardTitle>
          <CardDescription>
            {t('settings.notifications.channels.discord.description')}
          </CardDescription>
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
