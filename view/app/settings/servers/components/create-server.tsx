'use client';
import React from 'react';
import {
  Form,
} from '@/components/ui/form';
import { Button } from '@/components/ui/button';
import { Server } from '@/redux/types/server';
import { Plus } from 'lucide-react';
import { useServerForm } from '@/app/settings/servers/hooks/use-server-form';
import { DialogWrapper } from '@/components/ui/dialog-wrapper';
import ServerFormFields from './server-form-fields';

interface CreateServerDialogProps {
  open?: boolean;
  setOpen?: (open: boolean) => void;
  serverId?: string;
  serverData?: Server;
  mode?: 'create' | 'edit';
}

function CreateServerDialog({ open, setOpen, serverId, serverData, mode = 'create' }: CreateServerDialogProps) {
  const { t, form, authType, setAuthType, isEditMode, isLoading, onSubmit } = useServerForm({
    mode,
    serverData,
    serverId,
    onSuccess: () => {
      setOpen?.(false);
    }
  });

  const actions = [
    {
      label: t('servers.create.dialog.buttons.cancel'),
      onClick: () => {
        form.reset();
        setOpen?.(false);
      },
      variant: 'outline' as const
    },
    {
      label: isLoading
        ? (isEditMode ? t('servers.create.dialog.buttons.updating') : t('servers.create.dialog.buttons.creating'))
        : (isEditMode ? t('servers.create.dialog.buttons.update') : t('servers.create.dialog.buttons.create')),
      onClick: () => form.handleSubmit(onSubmit)(),
      disabled: isLoading,
      loading: isLoading
    }
  ];

  return (
    <DialogWrapper
      open={open}
      onOpenChange={setOpen}
      title={!isEditMode ? t('servers.create.dialog.title.add') : t('servers.create.dialog.title.update')}
      description={t('servers.create.dialog.description')}
      trigger={!isEditMode ? (
        <Button variant="outline">
          <Plus className="mr-2 h-4 w-4" />
          {t('servers.create.button')}
        </Button>
      ) : undefined}
      actions={actions}
      size="lg"
      contentClassName="max-h-[90vh] overflow-y-auto"
    >
      <Form {...form}>
        <form className="space-y-4" onSubmit={(e) => { e.preventDefault(); }}>
          <ServerFormFields
            form={form}
            t={t}
            authType={authType}
            setAuthType={setAuthType}
            isEditMode={isEditMode}
          />
        </form>
      </Form>
    </DialogWrapper>
  );
}

export default CreateServerDialog;
