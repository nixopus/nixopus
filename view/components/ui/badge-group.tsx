import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';

import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';

const badgeGroupVariants = cva('flex flex-wrap items-center', {
  variants: {
    gap: {
      sm: 'gap-1',
      default: 'gap-2',
      lg: 'gap-3'
    },
    orientation: {
      horizontal: 'flex-row',
      vertical: 'flex-col items-start'
    }
  },
  defaultVariants: {
    gap: 'default',
    orientation: 'horizontal'
  }
});

export interface BadgeGroupProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeGroupVariants> {
  /**
   * Whether badges are selectable/clickable
   * @default false
   */
  selectable?: boolean;
}

function BadgeGroup({
  className,
  gap,
  orientation,
  selectable = false,
  ...props
}: BadgeGroupProps) {
  return (
    <div
      data-slot="badge-group"
      className={cn(badgeGroupVariants({ gap, orientation }), className)}
      role={selectable ? 'group' : undefined}
      {...props}
    />
  );
}

export interface BadgeGroupItemProps extends React.ComponentProps<typeof Badge> {
  /**
   * Whether this badge is selected/active
   */
  selected?: boolean;
  /**
   * Whether the badge is clickable
   */
  clickable?: boolean;
  /**
   * Click handler (only used if clickable is true)
   */
  onClick?: () => void;
}

function BadgeGroupItem({
  className,
  selected = false,
  clickable = false,
  onClick,
  variant,
  ...props
}: BadgeGroupItemProps) {
  const displayVariant = selected ? 'default' : variant || 'outline';

  if (clickable) {
    return (
      <button
        type="button"
        onClick={onClick}
        className="focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded-md"
        aria-pressed={selected}
      >
        <Badge
          data-slot="badge-group-item"
          variant={displayVariant}
          className={cn('cursor-pointer transition-colors', className)}
          {...props}
        />
      </button>
    );
  }

  return (
    <Badge data-slot="badge-group-item" variant={displayVariant} className={className} {...props} />
  );
}

export { BadgeGroup, BadgeGroupItem, badgeGroupVariants };
