import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import React from 'react';
import { useDeleteDomainMutation } from '@/redux/services/settings/domainsApi';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';

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
