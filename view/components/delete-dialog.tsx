import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog';
import { Trash } from 'lucide-react';
import { useState } from 'react';

interface DeleteDialogProps {
  jobName: string;
  onDelete: () => void;
  showButton?: boolean;
  isDeleting?: boolean;
}

export function DeleteDialog({
  jobName,
  onDelete,
  showButton = true,
  isDeleting
}: DeleteDialogProps) {
  const [isOpen, setIsOpen] = useState(false);

  const handleDelete = () => {
    onDelete();
    setIsOpen(false);
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        {showButton ? (
          <Button variant="destructive" className="w-full sm:w-32" disabled={isDeleting}>
            Delete
          </Button>
        ) : (
          <Button variant="destructive" className="mr-2" disabled={isDeleting}>
            <Trash className="h-4 w-4" />
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Are you sure you want to delete {jobName}?</DialogTitle>
          <DialogDescription>
            This action cannot be undone. This will permanently remove {jobName} from your account.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="sm:justify-end">
          <Button variant="outline" onClick={() => setIsOpen(false)}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={handleDelete} disabled={isDeleting}>
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
