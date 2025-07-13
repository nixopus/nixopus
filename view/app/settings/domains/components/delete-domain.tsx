import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
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

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('settings.domains.delete.title')}</DialogTitle>
          <DialogDescription>{t('settings.domains.delete.description')}</DialogDescription>
        </DialogHeader>
        <DialogFooter className="flex justify-between sm:justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={() => setOpen(false)}>
            {t('settings.domains.delete.cancel')}
          </Button>
          <Button type="button" disabled={isLoading} onClick={handleDelete}>
            {isLoading
              ? t('settings.domains.delete.deleting')
              : t('settings.domains.delete.delete')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default DeleteDomain;
