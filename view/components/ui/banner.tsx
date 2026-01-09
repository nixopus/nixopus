import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { X, Info, AlertCircle, CheckCircle2, AlertTriangle, LucideIcon } from 'lucide-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

const bannerVariants = cva(
  'relative w-full rounded-lg border p-4 flex items-start justify-between gap-4 transition-colors',
  {
    variants: {
      variant: {
        default: 'bg-background text-foreground border-border',
        info: 'bg-primary/5 text-foreground border-primary/20 [&>svg]:text-primary',
        success:
          'bg-green-500/5 text-foreground border-green-500/20 [&>svg]:text-green-600 dark:[&>svg]:text-green-400',
        warning:
          'bg-yellow-500/5 text-foreground border-yellow-500/20 [&>svg]:text-yellow-600 dark:[&>svg]:text-yellow-400',
        destructive:
          'bg-destructive/5 text-destructive-foreground border-destructive/20 [&>svg]:text-destructive'
      },
      size: {
        default: 'p-4',
        sm: 'p-3',
        lg: 'p-6'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'default'
    }
  }
);

const iconVariants = cva('shrink-0', {
  variants: {
    variant: {
      default: 'text-muted-foreground',
      info: 'text-primary',
      success: 'text-green-600 dark:text-green-400',
      warning: 'text-yellow-600 dark:text-yellow-400',
      destructive: 'text-destructive'
    },
    size: {
      default: 'size-5',
      sm: 'size-4',
      lg: 'size-6'
    }
  },
  defaultVariants: {
    variant: 'default',
    size: 'default'
  }
});

export interface BannerProps
  extends React.ComponentProps<'div'>,
    VariantProps<typeof bannerVariants> {
  title?: string;
  description?: React.ReactNode;
  icon?: LucideIcon;
  showIcon?: boolean;
  onDismiss?: () => void;
  dismissible?: boolean;
  action?: React.ReactNode;
  children?: React.ReactNode;
}

const defaultIcons: Record<
  NonNullable<VariantProps<typeof bannerVariants>['variant']>,
  LucideIcon
> = {
  default: Info,
  info: Info,
  success: CheckCircle2,
  warning: AlertTriangle,
  destructive: AlertCircle
};

function Banner({
  className,
  variant = 'default',
  size = 'default',
  title,
  description,
  icon: Icon,
  showIcon = true,
  onDismiss,
  dismissible = false,
  action,
  children,
  ...props
}: BannerProps) {
  const DefaultIcon = variant ? defaultIcons[variant] : Info;
  const DisplayIcon = Icon || DefaultIcon;

  return (
    <div
      data-slot="banner"
      role="alert"
      className={cn(bannerVariants({ variant, size }), className)}
      {...props}
    >
      <div className="flex items-start gap-3 flex-1 min-w-0">
        {showIcon && (
          <div className={cn('mt-0.5', iconVariants({ variant, size }))}>
            <DisplayIcon className={cn(iconVariants({ variant, size }))} />
          </div>
        )}
        <div className="flex-1 min-w-0">
          {title && (
            <div data-slot="banner-title" className="text-sm font-medium text-foreground mb-1">
              {title}
            </div>
          )}
          {description && (
            <div
              data-slot="banner-description"
              className={cn('text-xs text-muted-foreground', title && 'mt-1')}
            >
              {description}
            </div>
          )}
          {children && (
            <div data-slot="banner-content" className={cn(title || description ? 'mt-2' : '')}>
              {children}
            </div>
          )}
        </div>
      </div>
      {(dismissible || action) && (
        <div className="flex items-center gap-2 shrink-0">
          {action}
          {dismissible && onDismiss && (
            <Button
              variant="ghost"
              size="icon"
              onClick={onDismiss}
              className="h-6 w-6 shrink-0"
              aria-label="Dismiss banner"
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      )}
    </div>
  );
}

export { Banner, bannerVariants };
