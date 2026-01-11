import { useMemo } from 'react';
import { Status } from '@/redux/types/applications';
import { cn } from '@/lib/utils';

interface UseStatusIndicatorProps {
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
  running: {
    color: 'text-emerald-500',
    bgColor: 'bg-emerald-500',
    label: 'Running',
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
  started: {
    color: 'text-amber-500',
    bgColor: 'bg-amber-500',
    label: 'Started',
    pulse: true
  },
  stopped: {
    color: 'text-zinc-500',
    bgColor: 'bg-zinc-500',
    label: 'Stopped',
    pulse: false
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

const sizeClasses = {
  sm: 'h-1.5 w-1.5',
  md: 'h-2 w-2',
  lg: 'h-3 w-3'
};

export function useStatusIndicator({
  status,
  size = 'md',
  showLabel = true
}: UseStatusIndicatorProps) {
  const config = useMemo(() => {
    return status ? statusConfig[status] || defaultConfig : defaultConfig;
  }, [status]);

  const sizeClass = useMemo(() => sizeClasses[size], [size]);

  const indicatorDot = useMemo(
    () => (
      <span className={cn('relative flex', sizeClass)}>
        {config.pulse && (
          <span
            className={cn(
              'animate-ping absolute inline-flex h-full w-full rounded-full opacity-75',
              config.bgColor
            )}
          />
        )}
        <span className={cn('relative inline-flex rounded-full', sizeClass, config.bgColor)} />
      </span>
    ),
    [config, sizeClass]
  );

  const label = useMemo(
    () =>
      showLabel ? (
        <span className={cn('text-sm font-medium capitalize', config.color)}>{config.label}</span>
      ) : null,
    [showLabel, config]
  );

  const noStatusLabel = useMemo(
    () => (showLabel ? <span className="text-sm text-muted-foreground">No deployment</span> : null),
    [showLabel]
  );

  const noStatusIndicatorDot = useMemo(
    () => (
      <span className={cn('relative flex', sizeClass)}>
        <span className={cn('relative inline-flex rounded-full bg-zinc-400', sizeClass)} />
      </span>
    ),
    [sizeClass]
  );

  return {
    config,
    sizeClass,
    indicatorDot,
    label,
    noStatusLabel,
    noStatusIndicatorDot,
    hasStatus: !!status
  };
}
