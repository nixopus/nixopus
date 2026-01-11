import { cn } from '@/lib/utils';

interface StatPillProps {
  value: number;
  label: string;
  color?: 'emerald' | 'zinc';
}

export function StatPill({ value, label, color }: StatPillProps) {
  return (
    <div className="flex items-center gap-2">
      {color && (
        <span
          className={cn(
            'w-2 h-2 rounded-full',
            color === 'emerald' ? 'bg-emerald-500' : 'bg-zinc-500'
          )}
        />
      )}
      <span className="text-xl font-bold">{value}</span>
      <span className="text-sm text-muted-foreground">{label}</span>
    </div>
  );
}
