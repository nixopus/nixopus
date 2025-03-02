import {
  useCreateSMPTConfigurationMutation,
  useGetNotificationPreferencesQuery,
  useGetSMTPConfigurationsQuery,
  useUpdateNotificationPreferencesMutation,
  useUpdateSMTPConfigurationMutation
} from '@/redux/services/settings/notificationApi';
import {
  CreateSMTPConfigRequest,
  UpdatePreferenceRequest,
  UpdateSMTPConfigRequest
} from '@/redux/types/notification';
import { toast } from 'sonner';

function useNotificationSettings() {
  const { data: smtpConfigs, isLoading, error } = useGetSMTPConfigurationsQuery();
  const [createSMTPConfiguration, { isLoading: isCreating }] = useCreateSMPTConfigurationMutation();
  const [updateSMTPConfiguration, { isLoading: isUpdating }] = useUpdateSMTPConfigurationMutation();

  const handleCreateSMTPConfiguration = async (data: CreateSMTPConfigRequest) => {
    await createSMTPConfiguration(data);
  };

  const handleUpdateSMTPConfiguration = async (data: UpdateSMTPConfigRequest) => {
    await updateSMTPConfiguration(data);
  };

  const { data: preferences, isLoading: isLoadingPreferences } =
    useGetNotificationPreferencesQuery();

  const [updateNotificationPreferences, { isLoading: isUpdatingPreferences }] =
    useUpdateNotificationPreferencesMutation();

  const handleOnSave = async (data: any) => {
    try {
      const smtpConfig = {
        host: data.smtpServer,
        port: parseInt(data.port),
        username: data.username,
        password: data.password,
        from_email: data.fromEmail,
        from_name: data.fromName
      };
      if (smtpConfigs?.id) {
        await handleUpdateSMTPConfiguration({ ...smtpConfig, id: smtpConfigs?.id });
      } else {
        await handleCreateSMTPConfiguration(smtpConfig);
      }
      toast.success('Email configuration saved successfully');
    } catch (error) {
      toast.error('Failed to save email configuration');
    }
  };

  const handleUpdatePreference = async (id: string, enabled: boolean) => {
    try {
      await updateNotificationPreferences({
        category: getPreferenceCategoryFromId(id),
        type: id,
        enabled: enabled
      });
      toast.success('Preferences updated successfully');
    } catch (error) {
      toast.error('Failed to update preferences');
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

  return {
    smtpConfigs,
    handleOnSave,
    preferences,
    isLoading:
      isLoading || isLoadingPreferences || isCreating || isUpdating || isUpdatingPreferences,
    error,
    handleUpdatePreference
  };
}

export default useNotificationSettings;
