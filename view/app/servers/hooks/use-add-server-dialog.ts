'use client';

import { useState, useMemo } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Key, Lock } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import type { CreateServerRequest } from '@/redux/types/servers';
import type { DialogAction } from '@/components/ui/dialog-wrapper';
import type { FormFieldConfig, AuthTabConfig, AuthMethod, UseAddServerDialogProps } from '../types';

const serverFormSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100),
  host: z.string().min(1, 'Host is required'),
  port: z.coerce.number().min(1).max(65535).default(22),
  ssh_user: z.string().min(1, 'SSH user is required'),
  ssh_private_key: z.string().optional(),
  ssh_password: z.string().optional(),
  docker_socket: z.string().default('/var/run/docker.sock'),
  labels: z.string().optional()
});

export type ServerFormValues = z.infer<typeof serverFormSchema>;

export function useAddServerDialog({ onSubmit, onOpenChange }: UseAddServerDialogProps) {
  const { t } = useTranslation();
  const [authMethod, setAuthMethod] = useState<AuthMethod>('key');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const form = useForm<ServerFormValues>({
    resolver: zodResolver(serverFormSchema),
    defaultValues: {
      name: '',
      host: '',
      port: 22,
      ssh_user: 'root',
      ssh_private_key: '',
      ssh_password: '',
      docker_socket: '/var/run/docker.sock',
      labels: ''
    }
  });

  const basicFields: FormFieldConfig[] = useMemo(
    () => [
      {
        name: 'name',
        label: t('servers.form.name.label'),
        placeholder: t('servers.form.name.placeholder'),
        type: 'text',
        colSpan: 3
      },
      {
        name: 'host',
        label: t('servers.form.host.label'),
        placeholder: t('servers.form.host.placeholder'),
        type: 'text',
        colSpan: 2
      },
      {
        name: 'port',
        label: t('servers.form.port.label'),
        placeholder: t('servers.form.port.placeholder'),
        type: 'number',
        colSpan: 1
      },
      {
        name: 'ssh_user',
        label: t('servers.form.sshUser.label'),
        placeholder: t('servers.form.sshUser.placeholder'),
        type: 'text',
        colSpan: 3
      }
    ],
    [t]
  );

  const advancedFields: FormFieldConfig[] = useMemo(
    () => [
      {
        name: 'docker_socket',
        label: t('servers.form.dockerSocket.label'),
        placeholder: t('servers.form.dockerSocket.placeholder'),
        type: 'text'
      },
      {
        name: 'labels',
        label: t('servers.form.labels.label'),
        placeholder: t('servers.form.labels.placeholder'),
        type: 'text',
        description: t('servers.form.labels.description')
      }
    ],
    [t]
  );

  const authTabs: AuthTabConfig[] = useMemo(
    () => [
      {
        value: 'key',
        label: t('servers.form.authMethod.sshKey'),
        icon: Key,
        field: {
          name: 'ssh_private_key',
          label: t('servers.form.sshKey.label'),
          placeholder: t('servers.form.sshKey.placeholder'),
          type: 'textarea',
          description: t('servers.form.sshKey.description')
        }
      },
      {
        value: 'password',
        label: t('servers.form.authMethod.password'),
        icon: Lock,
        field: {
          name: 'ssh_password',
          label: t('servers.form.sshPassword.label'),
          placeholder: t('servers.form.sshPassword.placeholder'),
          type: 'password'
        }
      }
    ],
    [t]
  );

  const handleSubmit = async (values: ServerFormValues) => {
    setIsSubmitting(true);
    try {
      const data: CreateServerRequest = {
        name: values.name,
        host: values.host,
        port: values.port,
        ssh_user: values.ssh_user,
        docker_socket: values.docker_socket,
        labels: values.labels
          ? values.labels
              .split(',')
              .map((l) => l.trim())
              .filter(Boolean)
          : []
      };

      if (authMethod === 'key' && values.ssh_private_key) {
        data.ssh_private_key = values.ssh_private_key;
      } else if (authMethod === 'password' && values.ssh_password) {
        data.ssh_password = values.ssh_password;
      }

      await onSubmit(data);
      form.reset();
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    form.reset();
    onOpenChange(false);
  };

  const actions: DialogAction[] = [
    {
      label: t('servers.dialog.delete.cancel'),
      onClick: handleCancel,
      variant: 'outline'
    },
    {
      label: t('servers.addServer'),
      onClick: form.handleSubmit(handleSubmit),
      disabled: isSubmitting,
      loading: isSubmitting
    }
  ];

  return {
    form,
    authMethod,
    setAuthMethod,
    isSubmitting,
    actions,
    basicFields,
    advancedFields,
    authTabs,
    authMethodLabel: t('servers.form.authMethod.label'),
    t
  };
}
