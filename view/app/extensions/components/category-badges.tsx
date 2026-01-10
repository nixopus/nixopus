'use client';

import React from 'react';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';

type CategoryBadgesProps = {
  categories: string[];
  selected?: string | null;
  onChange?: (value: string | null) => void;
  className?: string;
  showAll?: boolean;
};

export default function CategoryBadges({
  categories,
  selected = null,
  onChange,
  className,
  showAll = true
}: CategoryBadgesProps) {
  const handleSelect = (value: string | null) => {
    onChange?.(value);
  };

  return (
    <div className={cn('w-full', className)}>
      <div className="flex flex-wrap items-center gap-2">
        {showAll && (
          <button type="button" onClick={() => handleSelect(null)} className="focus:outline-none">
            <Badge variant={selected === null ? 'default' : 'outline'} className="px-3 py-1">
              All
            </Badge>
          </button>
        )}
        {categories.map((cat) => (
          <button
            key={cat}
            type="button"
            onClick={() => handleSelect(cat === selected ? null : cat)}
            className="focus:outline-none"
          >
            <Badge variant={selected === cat ? 'default' : 'outline'} className="px-3 py-1">
              {cat}
            </Badge>
          </button>
        ))}
      </div>
    </div>
  );
}
