import { useAppSelector } from '@/redux/hooks';
import {
  useCreateSMPTConfigurationMutation,
  useGetNotificationPreferencesQuery,
  useGetSMTPConfigurationsQuery,
  useUpdateNotificationPreferencesMutation,
  useUpdateSMTPConfigurationMutation,
  useGetWebhookConfigQuery,
  useCreateWebhookConfigMutation,
  useUpdateWebhookConfigMutation,
  useDeleteWebhookConfigMutation
} from '@/redux/services/settings/notificationApi';
import {
  CreateSMTPConfigRequest,
  CreateWebhookConfigRequest,
  SMTPFormData,
  UpdateSMTPConfigRequest,
  UpdateWebhookConfigRequest
} from '@/redux/types/notification';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';

function useNotificationSettings() {
  const { t } = useTranslation();
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const {
    data: smtpConfigs,
    isLoading: isLoadingSMTP,
    error: smtpError
  } = useGetSMTPConfigurationsQuery(activeOrganization?.id || '', {
    skip: !activeOrganization?.id
  });

  const {
    data: slackConfig,
    isLoading: isLoadingSlack,
    error: slackError
  } = useGetWebhookConfigQuery({ type: 'slack' }, { skip: !activeOrganization?.id });

  const {
    data: discordConfig,
    isLoading: isLoadingDiscord,
    error: discordError
  } = useGetWebhookConfigQuery({ type: 'discord' }, { skip: !activeOrganization?.id });

  const [createSMTPConfiguration, { isLoading: isCreatingSMTP }] =
    useCreateSMPTConfigurationMutation();
  const [updateSMTPConfiguration, { isLoading: isUpdatingSMTP }] =
    useUpdateSMTPConfigurationMutation();
  const [createWebhookConfig, { isLoading: isCreatingWebhook }] = useCreateWebhookConfigMutation();
  const [updateWebhookConfig, { isLoading: isUpdatingWebhook }] = useUpdateWebhookConfigMutation();
  const [deleteWebhookConfig] = useDeleteWebhookConfigMutation();

  const handleCreateSMTPConfiguration = async (data: CreateSMTPConfigRequest) => {
    try {
      await createSMTPConfiguration(data);
      toast.success(t('settings.notifications.messages.email.success'));
    } catch (error) {
      toast.error(t('settings.notifications.messages.email.error'));
    }
  };

  const handleUpdateSMTPConfiguration = async (data: UpdateSMTPConfigRequest) => {
    try {
      await updateSMTPConfiguration(data);
      toast.success(t('settings.notifications.messages.email.success'));
    } catch (error) {
      toast.error(t('settings.notifications.messages.email.error'));
    }
  };

  const handleCreateWebhookConfig = async (data: CreateWebhookConfigRequest) => {
    try {
      await createWebhookConfig(data);
      toast.success(t('settings.notifications.messages.webhookConfigSaved'));
    } catch (error) {
      toast.error(t('settings.notifications.messages.webhookConfigFailed'));
    }
  };

  const handleUpdateWebhookConfig = async (data: UpdateWebhookConfigRequest) => {
    try {
      await updateWebhookConfig(data);
      toast.success(t('settings.notifications.messages.webhookConfigUpdated'));
    } catch (error) {
      toast.error(t('settings.notifications.messages.webhookConfigFailed'));
    }
  };

  const handleDeleteWebhookConfig = async (type: string) => {
    try {
      await deleteWebhookConfig({ type, organization_id: activeOrganization?.id || '' });
      toast.success(t('settings.notifications.messages.webhookConfigDeleted'));
    } catch (error) {
      toast.error(t('settings.notifications.messages.webhookConfigFailed'));
    }
  };

  const { data: preferences, isLoading: isLoadingPreferences } =
    useGetNotificationPreferencesQuery();

  const [updateNotificationPreferences, { isLoading: isUpdatingPreferences }] =
    useUpdateNotificationPreferencesMutation();

  const handleOnSave = async (data: SMTPFormData) => {
    try {
      const smtpConfig = {
        host: data.smtp_host,
        port: parseInt(data.smtp_port),
        username: data.smtp_username,
        password: data.smtp_password,
        from_email: data.smtp_from_email,
        from_name: data.smtp_from_name
      };
      if (smtpConfigs?.id) {
        await handleUpdateSMTPConfiguration({
          ...smtpConfig,
          id: smtpConfigs.id,
          organization_id: activeOrganization?.id || ''
        });
      } else {
        await handleCreateSMTPConfiguration({
          ...smtpConfig,
          organization_id: activeOrganization?.id || ''
        });
      }
    } catch (error) {
      toast.error(t('settings.notifications.messages.email.error'));
    }
  };

  const handleUpdatePreference = async (id: string, enabled: boolean) => {
    try {
      const category = getPreferenceCategoryFromId(id);
      const type = getPreferenceTypeFromId(id);
      await updateNotificationPreferences({
        category,
        type,
        enabled
      });
      toast.success(t('settings.notifications.messages.preferencesUpdated'));
    } catch (error) {
      toast.error(t('settings.notifications.messages.preferencesFailed'));
    }
  };

  const getPreferenceCategoryFromId = (id: string): 'activity' | 'security' | 'update' => {
    if (preferences?.activity.some((preference) => preference.id === id)) {
      return 'activity';
    }
    if (preferences?.security.some((preference) => preference.id === id)) {
      return 'security';
    }
    if (preferences?.update.some((preference) => preference.id === id)) {
      return 'update';
    }
    return 'activity';
  };

  const getPreferenceTypeFromId = (
    id: string
  ):
    | 'password-changes'
    | 'security-alerts'
    | 'team-updates'
    | 'login-alerts'
    | 'product-updates'
    | 'newsletter'
    | 'marketing' => {
    switch (id) {
      case 'password-changes':
        return 'password-changes';
      case 'security-alerts':
        return 'security-alerts';
      case 'login-alerts':
        return 'login-alerts';
      case 'product-updates':
        return 'product-updates';
      case 'newsletter':
        return 'newsletter';
      case 'marketing':
        return 'marketing';
      case 'team-updates':
        return 'team-updates';
      default:
        return 'team-updates';
    }
  };

  return {
    smtpConfigs,
    slackConfig,
    discordConfig,
    handleOnSave,
    handleCreateWebhookConfig,
    handleUpdateWebhookConfig,
    handleDeleteWebhookConfig,
    preferences,
    isLoading:
      isLoadingSMTP ||
      isLoadingSlack ||
      isLoadingDiscord ||
      isLoadingPreferences ||
      isCreatingSMTP ||
      isUpdatingSMTP ||
      isCreatingWebhook ||
      isUpdatingWebhook ||
      isUpdatingPreferences,
    error: smtpError || slackError || discordError,
    handleUpdatePreference
  };
}

export default useNotificationSettings;
