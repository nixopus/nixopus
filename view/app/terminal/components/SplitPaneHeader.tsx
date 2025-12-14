'use client';

import React from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils';

// Color palette for split panes
const PANE_COLORS = [
  { name: 'blue', active: '#60a5fa', inactive: '#3b82f6' },
  { name: 'emerald', active: '#34d399', inactive: '#10b981' },
  { name: 'amber', active: '#fbbf24', inactive: '#f59e0b' },
  { name: 'purple', active: '#a78bfa', inactive: '#8b5cf6' },
];

type SplitPaneHeaderProps = {
  paneIndex: number;
  isActive: boolean;
  canClose: boolean;
  totalPanes: number;
  onFocus: () => void;
  onClose: () => void;
  closeLabel: string;
};

export const SplitPaneHeader: React.FC<SplitPaneHeaderProps> = ({
  paneIndex,
  isActive,
  canClose,
  totalPanes,
  onFocus,
  onClose,
  closeLabel
}) => {
  const color = PANE_COLORS[paneIndex % PANE_COLORS.length];
  const triangleColor = isActive ? color.active : color.inactive;
  const showTriangle = totalPanes > 1 && isActive;

  return (
    <div
      className={cn(
        'relative flex items-center justify-between h-6 px-2 cursor-pointer transition-all duration-200',
        'bg-transparent hover:bg-white/[0.02]'
      )}
      onClick={onFocus}
    >
      {showTriangle && (
        <div
          className="absolute top-0 left-0 w-0 h-0 z-10"
          style={{
            borderTop: `8px solid ${triangleColor}`,
            borderRight: '8px solid transparent',
          }}
        />
      )}
      <div className="flex-1" />
      <div className="flex items-center">
        {canClose && (
          <button
            className={cn(
              'p-0.5 rounded transition-all duration-200',
              'text-[#666] hover:text-[#fff] hover:bg-white/10'
            )}
            onClick={(e) => {
              e.stopPropagation();
              onClose();
            }}
            title={closeLabel}
          >
            <X className="h-3 w-3" />
          </button>
        )}
      </div>
    </div>
  );
};

