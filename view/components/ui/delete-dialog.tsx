import { DialogWrapper, DialogAction } from '@nixopus/ui';
import { LucideIcon } from 'lucide-react';
import { ReactNode, useState } from 'react';

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
  const isControlled = open !== undefined;
  const [internalOpen, setInternalOpen] = useState(false);
  const currentOpen = isControlled ? open : internalOpen;

  const handleOpenChange = (next: boolean) => {
    if (!isControlled) {
      setInternalOpen(next);
    }
    onOpenChange?.(next);
  };

  const actions: DialogAction[] = [
    {
      label: cancelText,
      onClick: () => handleOpenChange(false),
      variant: 'outline'
    },
    {
      label: confirmText,
      onClick: onConfirm,
      disabled: isDeleting,
      loading: isDeleting,
      variant: variant,
      icon: Icon,
      className: variant === 'destructive' ? 'bg-destructive' : ''
    }
  ];

  return (
    <DialogWrapper
      open={currentOpen}
      onOpenChange={handleOpenChange}
      title={title}
      description={description}
      trigger={trigger}
      actions={actions}
      size="sm"
    />
  );
}
