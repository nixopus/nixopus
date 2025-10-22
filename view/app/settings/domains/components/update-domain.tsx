import React from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useForm } from 'react-hook-form';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import {
  useCreateDomainMutation,
  useUpdateDomainMutation
} from '@/redux/services/settings/domainsApi';
import { Domain } from '@/redux/types/domain';
import { useAppSelector } from '@/redux/hooks';
import { useTranslation } from '@/hooks/use-translation';

interface UpdateDomainDialogProps {
  open: boolean;
  setOpen: React.Dispatch<React.SetStateAction<boolean>>;
  id?: string;
  data?: Domain;
}

function UpdateDomainDialog({ open, setOpen, id, data }: UpdateDomainDialogProps) {
  const { t } = useTranslation();
  const [createDomain, { isLoading }] = useCreateDomainMutation();
  const [updateDomain, { isLoading: isUpdating }] = useUpdateDomainMutation();

  const domainFormSchema = z.object({
    domainName: z
      .string()
      .min(3, { message: t('settings.domains.update.form.validation.minLength') })
      .refine(
        (domain) => {
          // Allow wildcard domains (e.g., *.example.com)
          if (domain.startsWith('*.')) {
            const baseDomain = domain.substring(2);
            return /^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9](?:\.[a-zA-Z]{2,})+$/.test(baseDomain);
          }
          // Regular domain validation
          return /^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9](?:\.[a-zA-Z]{2,})+$/.test(domain);
        },
        { message: t('settings.domains.update.form.validation.invalidFormat') }
      )
  });

  const form = useForm({
    resolver: zodResolver(domainFormSchema),
    defaultValues: {
      domainName: data?.name || ''
    }
  });
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);

  async function onSubmit(data: z.infer<typeof domainFormSchema>) {
    try {
      if (!id) {
        await createDomain({
          name: data.domainName,
          organization_id: activeOrganization?.id || ''
        }).unwrap();
      } else {
        await updateDomain({ name: data.domainName, id: id }).unwrap();
      }
      toast.success(t('settings.domains.update.success'));
      form.reset();
      setOpen(false);
    } catch (error) {
      toast.error(t('settings.domains.update.error'));
    }
  }

  const actions: DialogAction[] = [
    {
      label: t('settings.domains.update.buttons.cancel'),
      onClick: () => {
        form.reset();
        setOpen(false);
      },
      variant: 'outline'
    },
    {
      label: isLoading || isUpdating
        ? t('settings.domains.update.buttons.saving')
        : t('settings.domains.update.buttons.save'),
      onClick: form.handleSubmit(onSubmit),
      disabled: isLoading || isUpdating,
      loading: isLoading || isUpdating,
      variant: 'default'
    }
  ];

  return (
    <DialogWrapper
      open={open}
      onOpenChange={setOpen}
      title={!id ? t('settings.domains.update.addTitle') : t('settings.domains.update.updateTitle')}
      description={t('settings.domains.update.description')}
      actions={actions}
      size="lg"
    >
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <FormField
            control={form.control}
            name="domainName"
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <Input placeholder={t('settings.domains.update.form.placeholder')} {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </form>
      </Form>
    </DialogWrapper>
  );
}

export default UpdateDomainDialog;
