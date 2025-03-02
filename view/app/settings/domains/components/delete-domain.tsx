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

interface DeleteDomainProps {
  open: boolean;
  setOpen: (open: boolean) => void;
  id: string;
}

const DeleteDomain = ({ open, setOpen, id }: DeleteDomainProps) => {
  const [isLoading, setIsLoading] = React.useState(false);
  const [deleteDomain] = useDeleteDomainMutation();

  const handleDelete = async () => {
    setIsLoading(true);
    try {
      await deleteDomain(id);
      toast.success('Domain deleted successfully');
    } catch (error) {
      toast.error('Failed to delete domain');
    } finally {
      setIsLoading(false);
      setOpen(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Delete Domain</DialogTitle>
          <DialogDescription>Are you sure you want to delete this domain?</DialogDescription>
        </DialogHeader>
        <DialogFooter className="flex justify-between sm:justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button type="button" disabled={isLoading} onClick={handleDelete}>
            {isLoading ? 'Deleting...' : 'Delete'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default DeleteDomain;
