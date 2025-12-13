'use client';

import React from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { SessionStatus } from './TerminalSession';

type SessionTabProps = {
  session: { id: string; label: string };
  isActive: boolean;
  status: SessionStatus;
  onSelect: () => void;
  onClose: () => void;
  canClose: boolean;
  index: number;
};

export const SessionTab: React.FC<SessionTabProps> = ({
  session,
  isActive,
  status,
  onSelect,
  onClose,
  canClose,
  index
}) => {
  return (
    <div
      className={cn(
        'group relative flex items-center gap-1.5 px-2 py-1 rounded-md cursor-pointer transition-all duration-200 shrink-0',
        'hover:bg-[var(--terminal-tab-hover)]',
        isActive && 'bg-[var(--terminal-tab-active)] terminal-tab-active'
      )}
      onClick={onSelect}
      style={{
        animation: `terminalFadeIn 0.2s ease-out ${index * 0.05}s both`
      }}
    >
      <div className="relative flex items-center justify-center w-3 h-3">
        {status === 'loading' ? (
          <div className="w-2 h-2 rounded-full bg-amber-400 animate-pulse" />
        ) : status === 'active' ? (
          <div
            className={cn(
              'w-2 h-2 rounded-full bg-emerald-400',
              isActive && 'terminal-ready-indicator'
            )}
          />
        ) : (
          <div className="w-2 h-2 rounded-full bg-zinc-500" />
        )}
      </div>

      <span
        className={cn(
          'text-xs font-medium transition-colors duration-200 whitespace-nowrap',
          isActive ? 'text-[var(--terminal-text)]' : 'text-[var(--terminal-text-muted)]'
        )}
      >
        {session.label}
      </span>

      {canClose && (
        <button
          className={cn(
            'ml-1 p-0.5 rounded transition-all duration-200',
            'opacity-0 group-hover:opacity-100',
            'hover:bg-white/10 text-[var(--terminal-text-muted)] hover:text-[var(--terminal-text)]'
          )}
          onClick={(e) => {
            e.stopPropagation();
            onClose();
          }}
        >
          <X className="h-3 w-3" />
        </button>
      )}
    </div>
  );
};
