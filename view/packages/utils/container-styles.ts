import { cn } from '@/lib/utils';

export const getStatusColors = (status: string) => {
  const normalizedStatus = (status || '').toLowerCase();
  const isRunning = normalizedStatus === 'running';
  const isExited = normalizedStatus === 'exited';

  return {
    bg: isRunning ? 'bg-emerald-500/10' : isExited ? 'bg-red-500/10' : 'bg-zinc-500/10',
    text: isRunning ? 'text-emerald-500' : isExited ? 'text-red-500' : 'text-zinc-500',
    badge: isRunning
      ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
      : isExited
        ? 'bg-red-500/10 text-red-600 dark:text-red-400'
        : 'bg-zinc-500/10 text-zinc-600 dark:text-zinc-400',
    border: isRunning
      ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20'
      : isExited
        ? 'bg-red-500/10 text-red-500 border-red-500/20'
        : 'bg-amber-500/10 text-amber-500 border-amber-500/20',
    dot: isRunning ? 'bg-emerald-500' : isExited ? 'bg-red-500' : 'bg-amber-500',
    dotPulse: isRunning ? 'bg-emerald-400' : isExited ? 'bg-red-400' : 'bg-amber-400'
  };
};

export const getStatusIconClasses = (isRunning: boolean) => ({
  container: cn('p-2 rounded-lg flex-shrink-0', isRunning ? 'bg-emerald-500/10' : 'bg-zinc-500/10'),
  icon: cn('h-4 w-4', isRunning ? 'text-emerald-500' : 'text-zinc-500')
});

export const getPortColors = (hasPublic: boolean) => ({
  pill: hasPublic
    ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
    : 'bg-muted text-muted-foreground',
  flow: hasPublic ? 'bg-emerald-500/5' : 'bg-zinc-500/5',
  text: hasPublic ? 'text-emerald-600 dark:text-emerald-400' : 'text-muted-foreground'
});

export const textColorClasses = {
  emerald: 'text-emerald-500',
  red: 'text-red-500',
  amber: 'text-amber-500',
  blue: 'text-blue-500',
  purple: 'text-purple-500',
  zinc: 'text-zinc-500'
};

export const bgColorClasses = {
  emerald: 'bg-emerald-500/10',
  red: 'bg-red-500/10',
  amber: 'bg-amber-500/10',
  blue: 'bg-blue-500/10',
  purple: 'bg-purple-500/10',
  zinc: 'bg-zinc-500/10'
};

export const resourceGaugeColors = {
  blue: { bg: 'bg-blue-500', track: 'bg-blue-500/20', text: 'text-blue-500' },
  purple: { bg: 'bg-purple-500', track: 'bg-purple-500/20', text: 'text-purple-500' },
  amber: { bg: 'bg-amber-500', track: 'bg-amber-500/20', text: 'text-amber-500' },
  emerald: { bg: 'bg-emerald-500', track: 'bg-emerald-500/20', text: 'text-emerald-500' }
};
