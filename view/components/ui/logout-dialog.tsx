import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@nixopus/ui';
import { Button } from '@nixopus/ui';

interface LogoutDialogProps {
  open: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function LogoutDialog({ open, onConfirm, onCancel }: LogoutDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onCancel}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Confirm Logout</DialogTitle>
        </DialogHeader>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={onCancel}
            className="bg-gray-500 hover:bg-gray-600 text-white"
          >
            Cancel
          </Button>
          <Button onClick={onConfirm} className="bg-red-500 hover:bg-red-600 text-white">
            Logout
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
