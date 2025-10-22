import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
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
  const actions: DialogAction[] = [
    {
      label: cancelText,
      onClick: () => onOpenChange?.(false),
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
      open={open}
      onOpenChange={onOpenChange}
      title={title}
      description={description}
      trigger={trigger}
      actions={actions}
      size="sm"
    />
  );
}
