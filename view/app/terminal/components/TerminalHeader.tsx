'use client';

import React from 'react';
import { Plus, X, SplitSquareVertical, PanelRight, PanelBottom } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { SessionTab } from './SessionTab';
import type { SessionStatus, Session } from '../types';

type TerminalHeaderProps = {
  sessions: Session[];
  activeSessionId: string;
  sessionStatuses: Record<string, SessionStatus>;
  sessionLimit: number;
  maxSplits: number;
  splitPanesCount: number;
  onAddSession: () => void;
  onCloseSession: (id: string) => void;
  onSwitchSession: (id: string) => void;
  onToggleTerminal: () => void;
  onAddSplitPane: () => void;
  terminalPosition: 'bottom' | 'right';
  onTogglePosition: () => void;
  closeLabel: string;
  newTabLabel: string;
};

export const TerminalHeader: React.FC<TerminalHeaderProps> = ({
  sessions,
  activeSessionId,
  sessionStatuses,
  sessionLimit,
  maxSplits,
  splitPanesCount,
  onAddSession,
  onCloseSession,
  onSwitchSession,
  onToggleTerminal,
  onAddSplitPane,
  terminalPosition,
  onTogglePosition,
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
              canClose={true}
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
        {splitPanesCount < maxSplits && (
          <button
            className={cn(
              'flex items-center justify-center w-7 h-7 rounded-md transition-all duration-200 shrink-0',
              'hover:bg-[var(--terminal-tab-hover)] text-[var(--terminal-text-muted)]',
              'hover:text-[var(--terminal-accent)] hover:scale-105'
            )}
            onClick={onAddSplitPane}
            title="Split Terminal"
          >
            <SplitSquareVertical className="h-3.5 w-3.5" />
          </button>
        )}

        <Button
          variant="ghost"
          size="icon"
          onClick={onTogglePosition}
          title={terminalPosition === 'bottom' ? 'Move to Right' : 'Move to Bottom'}
          className={cn(
            'h-6 w-6 rounded-md transition-all duration-200',
            'hover:bg-[var(--terminal-tab-hover)] text-[var(--terminal-text-muted)]',
            'hover:text-[var(--terminal-accent)]'
          )}
        >
          {terminalPosition === 'bottom' ? (
            <PanelRight className="h-3.5 w-3.5" />
          ) : (
            <PanelBottom className="h-3.5 w-3.5" />
          )}
        </Button>

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
