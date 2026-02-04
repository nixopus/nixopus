'use client';
import React from 'react';
import { Badge } from '@nixopus/ui';
import { Button } from '@nixopus/ui';
import { X, Tag } from 'lucide-react';
import { cn } from '@/lib/utils';

interface LabelFilterProps {
  availableLabels: string[];
  selectedLabels: string[];
  onToggle: (label: string) => void;
  onClear: () => void;
  className?: string;
}

export function LabelFilter({
  availableLabels,
  selectedLabels,
  onToggle,
  onClear,
  className
}: LabelFilterProps) {
  if (availableLabels.length === 0) return null;

  const hasActiveFilters = selectedLabels.length > 0;

  return (
    <div className={cn('flex items-center gap-3', className)}>
      <div className="flex flex-wrap items-center gap-2 flex-1">
        {availableLabels.map((label) => (
          <LabelBadge
            key={label}
            label={label}
            isSelected={selectedLabels.includes(label)}
            onClick={() => onToggle(label)}
          />
        ))}
        {hasActiveFilters && (
          <Button
            variant="ghost"
            size="sm"
            onClick={onClear}
            className="h-6 px-2 gap-1 text-xs text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
          >
            <X className="h-3 w-3" />
            Clear
          </Button>
        )}
      </div>
    </div>
  );
}

interface LabelBadgeProps {
  label: string;
  isSelected: boolean;
  onClick: () => void;
}

function LabelBadge({ label, isSelected, onClick }: LabelBadgeProps) {
  return (
    <Badge
      variant="outline"
      className={cn(
        'cursor-pointer transition-all select-none text-xs px-2 py-0.5',
        'hover:scale-105 active:scale-95',
        isSelected ? 'border-violet-500 text-violet-500 bg-violet-500/20' : ''
      )}
      onClick={onClick}
    >
      {label}
    </Badge>
  );
}
