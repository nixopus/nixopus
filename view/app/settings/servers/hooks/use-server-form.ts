import React, { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useCreateServerMutation, useUpdateServerMutation } from '@/redux/services/settings/serversApi';
import { AuthenticationType, Server } from '@/redux/types/server';
import { useTranslation } from '@/hooks/use-translation';
import { toast } from 'sonner';

type Mode = 'create' | 'edit';

interface UseServerFormParams {
  mode?: Mode;
  serverData?: Server;
  serverId?: string;
  onSuccess?: () => void;
}

export function useServerForm({ mode = 'create', serverData, serverId, onSuccess }: UseServerFormParams) {
  const { t } = useTranslation();
  const [createServer, { isLoading: isCreating }] = useCreateServerMutation();
  const [updateServer, { isLoading: isUpdating }] = useUpdateServerMutation();

  const isEditMode = mode === 'edit';
  const isLoading = isCreating || isUpdating;

  const [authType, setAuthType] = useState<AuthenticationType>(() => {
    if (isEditMode && serverData) {
      return serverData.ssh_password ? AuthenticationType.PASSWORD : AuthenticationType.PRIVATE_KEY;
    }
    return AuthenticationType.PASSWORD;
  });

  const baseServerSchema = z.object({
    name: z
      .string()
      .min(2, { message: t('servers.create.dialog.validation.nameRequired') })
      .max(255, { message: t('servers.create.dialog.validation.nameMaxLength') }),
    description: z
      .string()
      .max(500, { message: t('servers.create.dialog.validation.descriptionMaxLength') })
      .optional(),
    host: z
      .string()
      .min(1, { message: t('servers.create.dialog.validation.hostRequired') })
      .refine(
        (host) => {
          const ipRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
          const hostnameRegex = /^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$/;
          return ipRegex.test(host) || hostnameRegex.test(host);
        },
        { message: t('servers.create.dialog.validation.invalidHost') }
      ),
    port: z
      .number()
      .min(1, { message: t('servers.create.dialog.validation.portRange') })
      .max(65535, { message: t('servers.create.dialog.validation.portRange') }),
    username: z
      .string()
      .min(1, { message: t('servers.create.dialog.validation.usernameRequired') })
      .regex(/^[a-zA-Z0-9\-_]+$/, { message: t('servers.create.dialog.validation.usernameInvalid') }),
    ssh_password: z.string().optional(),
    ssh_private_key_path: z.string().optional()
  });

  const createServerSchema = baseServerSchema.refine((data) => {
    if (authType === AuthenticationType.PASSWORD) {
      return data.ssh_password && data.ssh_password.length > 0;
    } else {
      return data.ssh_private_key_path && data.ssh_private_key_path.length > 0;
    }
  }, {
    message: t('servers.create.dialog.validation.authRequired'),
    path: authType === AuthenticationType.PASSWORD ? ['ssh_password'] : ['ssh_private_key_path']
  });

  const editServerSchema = baseServerSchema.refine((data) => {
    if (authType === AuthenticationType.PASSWORD && data.ssh_password) {
      return data.ssh_password.length > 0;
    }
    if (authType === AuthenticationType.PRIVATE_KEY && data.ssh_private_key_path) {
      return data.ssh_private_key_path.length > 0;
    }
    return true;
  }, {
    message: t('servers.create.dialog.validation.authRequired'),
    path: authType === AuthenticationType.PASSWORD ? ['ssh_password'] : ['ssh_private_key_path']
  });

  const serverFormSchema = isEditMode ? editServerSchema : createServerSchema;

  const form = useForm({
    resolver: zodResolver(serverFormSchema),
    defaultValues: {
      name: serverData?.name || '',
      description: serverData?.description || '',
      host: serverData?.host || '',
      port: serverData?.port || 22,
      username: serverData?.username || '',
      ssh_password: '',
      ssh_private_key_path: serverData?.ssh_private_key_path || ''
    }
  });

  useEffect(() => {
    if (serverData) {
      form.reset({
        name: serverData.name,
        description: serverData.description,
        host: serverData.host,
        port: serverData.port,
        username: serverData.username,
        ssh_password: '',
        ssh_private_key_path: serverData.ssh_private_key_path || ''
      });
      setAuthType(serverData.ssh_password ? AuthenticationType.PASSWORD : AuthenticationType.PRIVATE_KEY);
    } else {
      form.reset({
        name: '',
        description: '',
        host: '',
        port: 22,
        username: '',
        ssh_password: '',
        ssh_private_key_path: ''
      });
      setAuthType(AuthenticationType.PASSWORD);
    }
  }, [serverData, form]);

  async function onSubmit(formData: z.infer<typeof serverFormSchema>) {
    try {
      const baseRequestData = {
        name: formData.name,
        description: formData.description || '',
        host: formData.host,
        port: formData.port,
        username: formData.username
      };

      let requestData: any;

      if (isEditMode) {
        requestData = { ...baseRequestData } as any;
        if (authType === AuthenticationType.PASSWORD && formData.ssh_password) {
          (requestData as any).ssh_password = formData.ssh_password;
        }
        if (authType === AuthenticationType.PRIVATE_KEY && formData.ssh_private_key_path) {
          (requestData as any).ssh_private_key_path = formData.ssh_private_key_path;
        }
      } else {
        requestData = {
          ...baseRequestData,
          ...(authType === AuthenticationType.PASSWORD
            ? { ssh_password: formData.ssh_password }
            : { ssh_private_key_path: formData.ssh_private_key_path })
        };
      }

      if (isEditMode && serverId) {
        await updateServer({ id: serverId, ...requestData }).unwrap();
        toast.success(t('servers.messages.updateSuccess'));
      } else {
        await createServer(requestData).unwrap();
        toast.success(t('servers.messages.createSuccess'));
      }

      form.reset();
      onSuccess?.();
    } catch (e) {
      if (isEditMode) {
        toast.error(t('servers.messages.updateError'));
      } else {
        toast.error(t('servers.messages.createError'));
      }
    }
  }

  return {
    t,
    form,
    authType,
    setAuthType,
    isEditMode,
    isLoading,
    onSubmit
  };
}

export default useServerForm;

