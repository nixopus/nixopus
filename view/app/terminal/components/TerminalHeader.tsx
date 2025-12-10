'use client';

import React from 'react';
import { Plus, X, Zap } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { SessionTab } from './SessionTab';
import type { SessionStatus } from './TerminalSession';

type Session = {
  id: string;
  label: string;
};

type TerminalHeaderProps = {
  sessions: Session[];
  activeSessionId: string;
  sessionStatuses: Record<string, SessionStatus>;
  sessionLimit: number;
  onAddSession: () => void;
  onCloseSession: (id: string) => void;
  onSwitchSession: (id: string) => void;
  onToggleTerminal: () => void;
  closeLabel: string;
  newTabLabel: string;
};

export const TerminalHeader: React.FC<TerminalHeaderProps> = ({
  sessions,
  activeSessionId,
  sessionStatuses,
  sessionLimit,
  onAddSession,
  onCloseSession,
  onSwitchSession,
  onToggleTerminal,
  closeLabel,
  newTabLabel
}) => {
  return (
    <div
      className="flex h-10 min-h-10 items-center px-2 gap-2 overflow-hidden shrink-0"
      style={{
        borderBottom: '1px solid var(--terminal-border)',
        background: 'rgba(18, 18, 22, 0.98)',
        width: '100%',
        maxWidth: '100%',
        boxSizing: 'border-box'
      }}
    >
      <div className="flex-1 min-w-0 overflow-hidden">
        <div className="flex items-center gap-1 overflow-x-auto no-scrollbar">
          {sessions.map((session, index) => (
            <SessionTab
              key={session.id}
              session={session}
              isActive={session.id === activeSessionId}
              status={sessionStatuses[session.id] || 'idle'}
              onSelect={() => onSwitchSession(session.id)}
              onClose={() => onCloseSession(session.id)}
              canClose={sessions.length > 1}
              index={index}
            />
          ))}

          {sessions.length < sessionLimit && (
            <button
              className={cn(
                'flex items-center justify-center w-6 h-6 rounded-md transition-all duration-200 shrink-0',
                'hover:bg-[var(--terminal-tab-hover)] text-[var(--terminal-text-muted)]',
                'hover:text-[var(--terminal-accent)]'
              )}
              onClick={onAddSession}
              title={newTabLabel}
            >
              <Plus className="h-3.5 w-3.5" />
            </button>
          )}
        </div>
      </div>

      <div className="flex items-center gap-1 shrink-0">
        <div className="hidden sm:flex items-center gap-1 px-1.5 py-0.5 rounded bg-white/5">
          <Zap className="h-2.5 w-2.5 text-amber-400" />
          <span className="text-[9px] font-medium text-[var(--terminal-text-muted)]">
            {sessions.length}/{sessionLimit}
          </span>
        </div>

        <Button
          variant="ghost"
          size="icon"
          onClick={onToggleTerminal}
          title={closeLabel}
          className={cn(
            'h-6 w-6 rounded-md transition-all duration-200',
            'hover:bg-red-500/10 hover:text-red-400',
            'text-[var(--terminal-text-muted)]'
          )}
        >
          <X className="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  );
};
