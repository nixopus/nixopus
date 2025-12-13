'use client';

import React from 'react';
import { X, Triangle } from 'lucide-react';
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
  onFocus: () => void;
  onClose: () => void;
  closeLabel: string;
};

export const SplitPaneHeader: React.FC<SplitPaneHeaderProps> = ({
  paneIndex,
  isActive,
  canClose,
  onFocus,
  onClose,
  closeLabel
}) => {
  const color = PANE_COLORS[paneIndex % PANE_COLORS.length];
  const triangleColor = isActive ? color.active : color.inactive;

  return (
    <div
      className={cn(
        'flex items-center justify-between h-6 px-2 cursor-pointer transition-all duration-200',
        'bg-transparent hover:bg-white/[0.02]'
      )}
      onClick={onFocus}
    >
      <div className="flex items-center gap-1.5">
        <Triangle
          className={cn(
            'h-3 w-3 transition-all duration-300 rotate-[180deg]',
            isActive && 'drop-shadow-[0_0_4px_currentColor]'
          )}
          style={{ 
            color: triangleColor,
            fill: triangleColor,
            opacity: isActive ? 1 : 0.6
          }}
        />
      </div>
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

