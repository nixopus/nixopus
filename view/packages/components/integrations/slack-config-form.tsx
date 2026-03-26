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
  Switch
} from '@nixopus/ui';
import type { WebhookConfig } from '@/redux/types/notification';

const schema = z.object({
  webhook_url: z.string().url('Invalid webhook URL'),
  is_active: z.boolean().default(true)
});

type FormValues = z.infer<typeof schema>;

interface SlackConfigFormProps {
  config: WebhookConfig | null;
  onSave: (data: { webhook_url: string; is_active: boolean }) => Promise<void>;
  onDelete: (type: string) => Promise<void>;
  onClose: () => void;
  canDelete: boolean;
  isLoading?: boolean;
}

export function SlackConfigForm({
  config,
  onSave,
  onDelete,
  onClose,
  canDelete,
  isLoading
}: SlackConfigFormProps) {
  const { t } = useTranslation();
  const [confirmDelete, setConfirmDelete] = useState(false);

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      webhook_url: config?.webhook_url || '',
      is_active: config?.is_active ?? true
    }
  });

  useEffect(() => {
    if (config) {
      form.reset({ webhook_url: config.webhook_url || '', is_active: config.is_active });
    }
  }, [config]); // eslint-disable-line react-hooks/exhaustive-deps

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
              await onDelete('slack');
              onClose();
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
          await onSave(data);
          onClose();
        })}
        className="space-y-4"
      >
        <FormField
          control={form.control}
          name="webhook_url"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Webhook URL</FormLabel>
              <FormControl>
                <Input {...field} placeholder="https://hooks.slack.com/services/..." />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="is_active"
          render={({ field }) => (
            <FormItem className="flex items-center justify-between">
              <FormLabel>Active</FormLabel>
              <FormControl>
                <Switch checked={field.value} onCheckedChange={field.onChange} />
              </FormControl>
            </FormItem>
          )}
        />
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
