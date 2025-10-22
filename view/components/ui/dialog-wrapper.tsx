'use client';

import React from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { LucideIcon } from 'lucide-react';

export interface DialogAction {
  label: string;
  onClick: () => void;
  variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
  disabled?: boolean;
  loading?: boolean;
  icon?: LucideIcon;
  className?: string;
}

export interface DialogWrapperProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  title?: string;
  description?: string;
  children?: React.ReactNode;
  trigger?: React.ReactNode;
  actions?: DialogAction[];
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  className?: string;
  contentClassName?: string;
  headerClassName?: string;
  footerClassName?: string;
  showCloseButton?: boolean;
  closeOnOverlayClick?: boolean;
  closeOnEscape?: boolean;
  loading?: boolean;
  error?: string;
  success?: string;
}

export function DialogWrapper({
  open,
  onOpenChange,
  title,
  description,
  children,
  trigger,
  actions = [],
  size = 'md',
  className,
  contentClassName,
  headerClassName,
  footerClassName,
  showCloseButton = true,
  closeOnOverlayClick = true,
  closeOnEscape = true,
  loading = false,
  error,
  success
}: DialogWrapperProps) {
  const getSizeClasses = () => {
    switch (size) {
      case 'sm':
        return 'sm:max-w-[425px]';
      case 'lg':
        return 'sm:max-w-[600px]';
      case 'xl':
        return 'sm:max-w-[800px]';
      case 'full':
        return 'sm:max-w-[95vw]';
      default:
        return 'sm:max-w-[500px]';
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!closeOnOverlayClick && !closeOnEscape) {
      return;
    }
    onOpenChange?.(newOpen);
  };

  const renderActions = () => {
    if (actions.length === 0) return null;

    return (
      <DialogFooter className={cn('sm:justify-end', footerClassName)}>
        {actions.map((action, index) => {
          const Icon = action.icon;
          return (
            <Button
              key={index}
              variant={action.variant || 'default'}
              onClick={action.onClick}
              disabled={action.disabled || action.loading || loading}
              className={action.className}
            >
              {action.loading && (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-current mr-2" />
              )}
              {Icon && !action.loading && <Icon className="mr-2 h-4 w-4" />}
              {action.label}
            </Button>
          );
        })}
      </DialogFooter>
    );
  };

  const renderStatusMessage = () => {
    if (error) {
      return (
        <div className="p-3 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-md">
          {error}
        </div>
      );
    }
    if (success) {
      return (
        <div className="p-3 text-sm text-green-600 bg-green-50 border border-green-200 rounded-md">
          {success}
        </div>
      );
    }
    return null;
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      {trigger && <DialogTrigger asChild>{trigger}</DialogTrigger>}
      <DialogContent 
        className={cn(
          getSizeClasses(),
          contentClassName
        )}
      >
        {(title || description) && (
          <DialogHeader className={headerClassName}>
            {title && <DialogTitle>{title}</DialogTitle>}
            {description && <DialogDescription>{description}</DialogDescription>}
          </DialogHeader>
        )}
        
        <div className={cn('space-y-4', className)}>
          {renderStatusMessage()}
          {children}
        </div>
        
        {renderActions()}
      </DialogContent>
    </Dialog>
  );
}

export default DialogWrapper;
