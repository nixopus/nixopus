import {
    useCreateSMPTConfigurationMutation,
    useGetSMTPConfigurationsQuery,
    useUpdateSMTPConfigurationMutation
} from '@/redux/services/settings/notification';
import { CreateSMTPConfigRequest, UpdateSMTPConfigRequest } from '@/redux/types/notification';
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

    const handleOnSave = async (data: any) => {
        try {
            const smtpConfig = {
                host: data.smtpServer,
                port: parseInt(data.port),
                username: data.username,
                password: data.password,
                from_email: data.fromEmail,
                from_name: data.fromName
            }
            if (smtpConfigs?.id) {
                await handleUpdateSMTPConfiguration({ ...smtpConfig, id: smtpConfigs?.id, });
            } else {
                await handleCreateSMTPConfiguration(smtpConfig);
            }
            toast.success('Email configuration saved successfully');
        } catch (error) {
            toast.error('Failed to save email configuration');
        }
    };

    return {
        smtpConfigs,
        isLoading,
        error,
        isCreating,
        isUpdating,
        handleOnSave
    };
}

export default useNotificationSettings;
