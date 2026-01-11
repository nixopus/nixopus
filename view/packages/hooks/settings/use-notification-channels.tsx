'use client';

import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { toast } from 'sonner';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { SMTPFormData } from '@/redux/types/notification';
import { ChannelTabProps } from '@/packages/types/settings';

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

export function useNotificationChannels({
  smtpConfigs,
  slackConfig,
  discordConfig,
  isLoading,
  handleOnSave,
  handleOnSaveSlack,
  handleOnSaveDiscord
}: ChannelTabProps) {
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [slackConfig]);

  useEffect(() => {
    if (discordConfig) {
      discordForm.setValue('webhook_url', discordConfig.webhook_url || '');
      discordForm.setValue('is_active', discordConfig.is_active);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
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

  return {
    emailForm,
    slackForm,
    discordForm,
    onSubmitEmail,
    onSubmitSlack,
    onSubmitDiscord,
    isLoading
  };
}
