import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import React from 'react';
import { useDeleteDomainMutation } from '@/redux/services/settings/domainsApi';
import { toast } from 'sonner';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Domain } from '@/redux/types/domain';
import { ResourceGuard } from '@/packages/components/rbac';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { Button } from '@/components/ui/button';
import { Edit, Trash2 } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { DataTable, TableColumn } from '@/components/ui/data-table';
import { useCreateDomainMutation } from '@/redux/services/settings/domainsApi';
import { useUpdateDomainMutation } from '@/redux/services/settings/domainsApi';
import { useAppSelector } from '@/redux/hooks';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { Input } from '@/components/ui/input';
import { Form, FormControl, FormField, FormItem, FormMessage } from '@/components/ui/form';

interface DeleteDomainProps {
  open: boolean;
  setOpen: (open: boolean) => void;
  id: string;
}

const DeleteDomain = ({ open, setOpen, id }: DeleteDomainProps) => {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = React.useState(false);
  const [deleteDomain] = useDeleteDomainMutation();

  const handleDelete = async () => {
    setIsLoading(true);
    try {
      await deleteDomain({ id });
      toast.success(t('settings.domains.delete.success'));
    } catch (error) {
      toast.error(t('settings.domains.delete.error'));
    } finally {
      setIsLoading(false);
      setOpen(false);
    }
  };

  const actions: DialogAction[] = [
    {
      label: t('settings.domains.delete.cancel'),
      onClick: () => setOpen(false),
      variant: 'outline'
    },
    {
      label: isLoading
        ? t('settings.domains.delete.deleting')
        : t('settings.domains.delete.delete'),
      onClick: handleDelete,
      disabled: isLoading,
      loading: isLoading,
      variant: 'destructive'
    }
  ];

  return (
    <DialogWrapper
      open={open}
      onOpenChange={setOpen}
      title={t('settings.domains.delete.title')}
      description={t('settings.domains.delete.description')}
      actions={actions}
      size="lg"
    />
  );
};

export default DeleteDomain;

interface DomainActionsProps {
  domain: Domain;
}

export function DomainActions({ domain }: DomainActionsProps) {
  const { t } = useTranslation();
  const [isEditModalOpen, setIsEditModalOpen] = React.useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = React.useState(false);

  const handleEdit = () => {
    setIsEditModalOpen(true);
  };

  const handleDelete = () => {
    setIsDeleteModalOpen(true);
  };

  return (
    <div className="flex justify-end gap-2">
      <ResourceGuard resource="domain" action="update">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0 text-primary hover:text-primary/80 hover:bg-primary/10"
              onClick={handleEdit}
            >
              <Edit className="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>{t('settings.domains.actions.edit')}</p>
          </TooltipContent>
        </Tooltip>
      </ResourceGuard>
      <ResourceGuard resource="domain" action="delete">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0 text-red-500 hover:text-red-600 hover:bg-red-50"
              onClick={handleDelete}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>{t('settings.domains.actions.delete')}</p>
          </TooltipContent>
        </Tooltip>
      </ResourceGuard>
      {isEditModalOpen && (
        <UpdateDomainDialog
          open={isEditModalOpen}
          setOpen={setIsEditModalOpen}
          id={domain.id}
          data={domain}
        />
      )}
      {isDeleteModalOpen && (
        <DeleteDomain open={isDeleteModalOpen} setOpen={setIsDeleteModalOpen} id={domain.id} />
      )}
    </div>
  );
}

interface DomainTypeTagProps {
  isWildcard: boolean;
}

export function DomainTypeTag({ isWildcard }: DomainTypeTagProps) {
  const { t } = useTranslation();

  if (isWildcard) {
    return (
      <Badge variant="secondary" className="bg-secondary/20 text-secondary-foreground rounded-full">
        {t('settings.domains.types.wildcard')}
      </Badge>
    );
  }

  return (
    <Badge variant="outline" className="bg-primary/10 text-primary rounded-full">
      {t('settings.domains.types.standard')}
    </Badge>
  );
}

interface DomainsTableProps {
  domains: Domain[];
}

function DomainsTable({ domains }: DomainsTableProps) {
  const { t } = useTranslation();

  const columns: TableColumn<Domain>[] = [
    {
      key: 'name',
      title: t('settings.domains.table.headers.domain'),
      dataIndex: 'name'
    },
    {
      key: 'type',
      title: t('settings.domains.table.headers.type'),
      render: (_, domain) => <DomainTypeTag isWildcard={domain.name.startsWith('*')} />
    },
    {
      key: 'actions',
      title: t('settings.domains.table.headers.actions'),
      render: (_, domain) => <DomainActions domain={domain} />,
      align: 'right'
    }
  ];

  return (
    <DataTable
      data={domains}
      columns={columns}
      containerClassName="divide-y divide-border"
      showBorder={true}
    />
  );
}

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
      label:
        isLoading || isUpdating
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
