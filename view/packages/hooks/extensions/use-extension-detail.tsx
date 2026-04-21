import { useMemo, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useGetExtensionQuery } from '@/redux/services/extensions/extensionsApi';
import { useCreateTemplateDeploymentMutation } from '@/redux/services/deploy/applicationsApi';
import { DialogAction } from '@nixopus/ui';
import { useExtensionInput } from '@/packages/hooks/extensions/use-extension-input';
import { TableColumn } from '@nixopus/ui';
import { Extension } from '@/redux/types/extension';
import { toast } from 'sonner';

function useExtensionDetails() {
  const { t } = useTranslation();
  const params = useParams();
  const router = useRouter();
  const id = (params?.id as string) || '';
  const { data: extension, isLoading } = useGetExtensionQuery({ id });
  const [runModalOpen, setRunModalOpen] = useState(false);
  const [createTemplateDeployment, { isLoading: isInstalling }] =
    useCreateTemplateDeploymentMutation();

  const handleRunExtension = async (values: Record<string, unknown>) => {
    if (!extension) return;
    try {
      const app = await createTemplateDeployment({
        template_id: extension.extension_id,
        name: extension.name,
        variables: values
      }).unwrap();
      setRunModalOpen(false);
      router.push(`/apps/application/${app.id}`);
      toast.success(t('extensions.installSuccess') || 'App created successfully');
    } catch (e) {
      toast.error(t('extensions.installFailed') || 'Failed to create app');
    }
  };

  const { values, errors, handleChange, handleSubmit, requiredFields } = useExtensionInput({
    extension,
    open: runModalOpen,
    onSubmit: handleRunExtension,
    onClose: () => setRunModalOpen(false)
  });

  const actions: DialogAction[] = [
    {
      label: t('common.cancel'),
      onClick: () => setRunModalOpen(false),
      variant: 'ghost'
    },
    {
      label: t('extensions.install'),
      onClick: handleSubmit,
      variant: 'default',
      disabled: isInstalling,
      loading: isInstalling
    }
  ];

  const isOnlyProxyDomain =
    requiredFields.length === 1 &&
    (requiredFields[0].variable_name.toLowerCase() === 'proxy_domain' ||
      requiredFields[0].variable_name.toLowerCase() === 'domain');
  const noFieldsToShow = requiredFields.length === 0;

  const variableColumns: TableColumn<NonNullable<Extension['variables']>[0]>[] = useMemo(
    () => [
      {
        key: 'name',
        title: 'Name',
        dataIndex: 'variable_name',
        width: '25%'
      },
      {
        key: 'type',
        title: 'Type',
        dataIndex: 'variable_type',
        width: '17%'
      },
      {
        key: 'required',
        title: 'Required',
        render: (_, record) => (record.is_required ? 'Yes' : 'No'),
        width: '17%'
      },
      {
        key: 'default',
        title: 'Default',
        render: (_, record) => String(record.default_value ?? ''),
        width: '17%',
        className: 'truncate'
      },
      {
        key: 'description',
        title: 'Description',
        dataIndex: 'description',
        width: '24%'
      }
    ],
    []
  );

  const buttonText = t('extensions.install') || 'Install';

  return {
    runModalOpen,
    isLoading,
    extension,
    router,
    setRunModalOpen,
    t,
    variableColumns,
    handleRunExtension,
    handleChange,
    handleSubmit,
    requiredFields,
    values,
    errors,
    buttonText,
    isOnlyProxyDomain,
    noFieldsToShow,
    actions
  };
}

export default useExtensionDetails;
