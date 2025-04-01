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
import { LucideIcon } from 'lucide-react';
import { ReactNode } from 'react';

interface ConfirmationDialogProps {
  title: string;
  description: string;
  onConfirm: () => void;
  trigger?: ReactNode;
  confirmText?: string;
  cancelText?: string;
  isDeleting?: boolean;
  variant?: 'default' | 'destructive';
  icon?: LucideIcon;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function DeleteDialog({
  title,
  description,
  onConfirm,
  trigger,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  isDeleting,
  variant = 'default',
  icon: Icon,
  open,
  onOpenChange
}: ConfirmationDialogProps) {
  const handleConfirm = () => {
    onConfirm();
    onOpenChange?.(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {trigger && <DialogTrigger asChild>{trigger}</DialogTrigger>}
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogFooter className="sm:justify-end">
          <Button variant="outline" onClick={() => onOpenChange?.(false)}>
            {cancelText}
          </Button>
          <Button
            variant={variant}
            onClick={handleConfirm}
            disabled={isDeleting}
            className={variant === 'destructive' ? 'bg-destructive' : ''}
          >
            {Icon && <Icon className="mr-2 h-4 w-4" />}
            {confirmText}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
