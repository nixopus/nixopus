'use client';

import { cn } from '@/lib/utils';

interface StatBlockProps {
  value: string | number;
  label: string;
  sublabel?: string;
  color?: 'emerald' | 'red' | 'amber' | 'blue' | 'purple';
  pulse?: boolean;
}

export function StatBlock({ value, label, sublabel, color, pulse }: StatBlockProps) {
  const colorClasses = {
    emerald: 'text-emerald-500',
    red: 'text-red-500',
    amber: 'text-amber-500',
    blue: 'text-blue-500',
    purple: 'text-purple-500'
  };

  return (
    <div className="relative">
      <div className="space-y-1">
        <div className="flex items-center gap-2">
          {pulse && (
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
            </span>
          )}
          <span
            className={cn(
              'text-2xl font-bold tracking-tight capitalize',
              color && colorClasses[color]
            )}
          >
            {value}
          </span>
        </div>
        <p className="text-sm text-muted-foreground">{label}</p>
        {sublabel && <p className="text-xs text-muted-foreground/60">{sublabel}</p>}
      </div>
    </div>
  );
}
