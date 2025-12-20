'use client';
import React from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { X, Filter } from 'lucide-react';
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
    <div className={cn('flex items-center justify-between gap-4', className)}>
      <LabelBadges labels={availableLabels} selected={selectedLabels} onToggle={onToggle} />
      {hasActiveFilters && <ClearButton onClick={onClear} />}
    </div>
  );
}

interface ClearButtonProps {
  onClick: () => void;
}

function ClearButton({ onClick }: ClearButtonProps) {
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={onClick}
      className="h-8 gap-1.5 text-xs shrink-0 hover:bg-destructive/10 hover:text-destructive"
    >
      <X className="h-3.5 w-3.5" />
      Clear filters
    </Button>
  );
}

interface LabelBadgesProps {
  labels: string[];
  selected: string[];
  onToggle: (label: string) => void;
}

function LabelBadges({ labels, selected, onToggle }: LabelBadgesProps) {
  return (
    <div className="flex items-center gap-2 flex-wrap flex-1">
      <div className="flex flex-wrap gap-2">
        {labels.map((label) => (
          <LabelBadge
            key={label}
            label={label}
            isSelected={selected.includes(label)}
            onClick={() => onToggle(label)}
          />
        ))}
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
      variant={isSelected ? 'default' : 'secondary'}
      className={cn(
        'cursor-pointer transition-all select-none',
        'hover:scale-105',
        'active:scale-95',
        'text-xs px-3 py-1',
        isSelected ? 'bg-primary text-primary-foreground' : 'bg-secondary/50 hover:bg-secondary'
      )}
      onClick={onClick}
    >
      {label}
    </Badge>
  );
}
