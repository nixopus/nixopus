'use client';

import { cn } from '@/lib/utils';
import { Status } from '@/redux/types/applications';

interface StatusIndicatorProps {
  status?: Status;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
}

const statusConfig: Record<
  Status,
  { color: string; bgColor: string; label: string; pulse?: boolean }
> = {
  draft: {
    color: 'text-blue-500',
    bgColor: 'bg-blue-500',
    label: 'Draft',
    pulse: false
  },
  deployed: {
    color: 'text-emerald-500',
    bgColor: 'bg-emerald-500',
    label: 'Deployed',
    pulse: true
  },
  deploying: {
    color: 'text-blue-500',
    bgColor: 'bg-blue-500',
    label: 'Deploying',
    pulse: true
  },
  building: {
    color: 'text-amber-500',
    bgColor: 'bg-amber-500',
    label: 'Building',
    pulse: true
  },
  cloning: {
    color: 'text-purple-500',
    bgColor: 'bg-purple-500',
    label: 'Cloning',
    pulse: true
  },
  failed: {
    color: 'text-red-500',
    bgColor: 'bg-red-500',
    label: 'Failed',
    pulse: false
  }
};

const defaultConfig = {
  color: 'text-zinc-500',
  bgColor: 'bg-zinc-500',
  label: 'Unknown',
  pulse: false
};

export function StatusIndicator({ status, size = 'md', showLabel = true }: StatusIndicatorProps) {
  const sizeClasses = {
    sm: 'h-1.5 w-1.5',
    md: 'h-2 w-2',
    lg: 'h-3 w-3'
  };

  if (!status) {
    return (
      <div className="flex items-center gap-2">
        <span className={cn('relative flex', sizeClasses[size])}>
          <span
            className={cn('relative inline-flex rounded-full bg-zinc-400', sizeClasses[size])}
          />
        </span>
        {showLabel && <span className="text-sm text-muted-foreground">No deployment</span>}
      </div>
    );
  }

  const config = statusConfig[status] || defaultConfig;

  return (
    <div className="flex items-center gap-2">
      <span className={cn('relative flex', sizeClasses[size])}>
        {config.pulse && (
          <span
            className={cn(
              'animate-ping absolute inline-flex h-full w-full rounded-full opacity-75',
              config.bgColor
            )}
          />
        )}
        <span
          className={cn('relative inline-flex rounded-full', sizeClasses[size], config.bgColor)}
        />
      </span>
      {showLabel && (
        <span className={cn('text-sm font-medium capitalize', config.color)}>{config.label}</span>
      )}
    </div>
  );
}
