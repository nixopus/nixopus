'use client';

import React, { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Button,
  PasswordInputField
} from '@nixopus/ui';
import type { SMTPConfig, SMTPFormData } from '@/redux/types/notification';

const smtpSchema = z.object({
  smtp_host: z.string().min(1, 'SMTP host is required'),
  smtp_port: z.string().min(1, 'SMTP port is required'),
  smtp_username: z.string().min(1, 'SMTP username is required'),
  smtp_password: z.string().min(1, 'SMTP password is required'),
  smtp_from_email: z.string().email('Invalid email address'),
  smtp_from_name: z.string().min(1, 'From name is required')
});

interface SmtpConfigFormProps {
  config: SMTPConfig | null;
  onSave: (data: SMTPFormData) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onClose: () => void;
  canDelete: boolean;
  isLoading?: boolean;
}

export function SmtpConfigForm({
  config,
  onSave,
  onDelete,
  onClose,
  canDelete,
  isLoading
}: SmtpConfigFormProps) {
  const { t } = useTranslation();
  const [confirmDelete, setConfirmDelete] = useState(false);

  const form = useForm<SMTPFormData>({
    resolver: zodResolver(smtpSchema),
    defaultValues: {
      smtp_host: config?.host || '',
      smtp_port: config?.port?.toString() || '',
      smtp_username: config?.username || '',
      smtp_password: config?.password || '',
      smtp_from_email: config?.from_email || '',
      smtp_from_name: config?.from_name || ''
    }
  });

  useEffect(() => {
    if (config) {
      form.reset({
        smtp_host: config.host || '',
        smtp_port: config.port?.toString() || '',
        smtp_username: config.username || '',
        smtp_password: config.password || '',
        smtp_from_email: config.from_email || '',
        smtp_from_name: config.from_name || ''
      });
    }
  }, [config, form]);

  const fields: { name: keyof SMTPFormData; isPassword?: boolean }[] = [
    { name: 'smtp_host' },
    { name: 'smtp_port' },
    { name: 'smtp_username' },
    { name: 'smtp_password', isPassword: true },
    { name: 'smtp_from_email' },
    { name: 'smtp_from_name' }
  ];

  if (confirmDelete) {
    return (
      <div className="space-y-4">
        <p className="text-sm text-muted-foreground">
          {t('integrations.modal.deleteConfirm' as any)}
        </p>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setConfirmDelete(false)}>
            {t('integrations.modal.cancel' as any)}
          </Button>
          <Button
            variant="destructive"
            onClick={async () => {
              try {
                await onDelete(config!.id);
                onClose();
              } catch {
                // error already handled
              }
            }}
            disabled={isLoading}
          >
            {t('integrations.modal.deleteConfirmButton' as any)}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(async (data) => {
          try {
            await onSave(data);
            onClose();
          } catch {
            // error already handled
          }
        })}
        className="space-y-4"
      >
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {fields.map(({ name, isPassword }) => (
            <FormField
              key={name}
              control={form.control}
              name={name}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>
                    {t(`settings.notifications.channels.email.fields.${name}.label` as any)}
                  </FormLabel>
                  <FormControl>
                    {isPassword ? <PasswordInputField {...field} /> : <Input {...field} />}
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          ))}
        </div>
        <div className="flex items-center justify-between pt-2">
          <div>
            {config && canDelete && (
              <Button
                type="button"
                variant="destructive"
                size="sm"
                onClick={() => setConfirmDelete(true)}
              >
                {t('integrations.modal.delete' as any)}
              </Button>
            )}
          </div>
          <div className="flex gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              {t('integrations.modal.cancel' as any)}
            </Button>
            <Button type="submit" disabled={isLoading}>
              {t('integrations.modal.save' as any)}
            </Button>
          </div>
        </div>
      </form>
    </Form>
  );
}
