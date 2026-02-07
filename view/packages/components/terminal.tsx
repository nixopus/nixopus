'use client';

import React, { useCallback, useRef } from 'react';
import { cn } from '@/lib/utils';
import { PanelBottom, PanelRight, Plus, SplitSquareVertical, X } from 'lucide-react';
import { useSplitPaneHeader } from '@/packages/hooks/terminal/use-split-pane-header';
import { Button } from '@nixopus/ui';
import {
  SplitPaneHeaderProps,
  TerminalHeaderProps,
  SessionTabProps,
  TerminalPaneProps
} from '../types/terminal';
import useTerminalPane from '../hooks/terminal/use-terminal-pane';
import '@xterm/xterm/css/xterm.css';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useFeatureFlags } from '@/packages/hooks/shared/features_provider';
import DisabledFeature from '@/packages/components/rbac';
import { FeatureNames } from '@/packages/types/feature-flags';
import { AnyPermissionGuard } from '@/packages/components/rbac';
import { useRBAC } from '@/packages/utils/rbac';
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from '@nixopus/ui';
import { useTerminalSessions } from '@/packages/hooks/terminal/use-terminal-sessions';
import { useTerminalStyles } from '@/packages/hooks/terminal/use-terminal-styles';
import { useTerminalKeyboardShortcuts } from '@/packages/hooks/terminal/use-terminal-keyboard-shortcuts';

export const TerminalPane: React.FC<TerminalPaneProps> = ({
  isActive,
  isTerminalOpen,
  canCreate,
  canUpdate,
  setFitAddonRef,
  terminalId,
  onFocus,
  onStatusChange,
  exitHandler
}) => {
  const { paneRef, terminalRef } = useTerminalPane({
    isActive,
    isTerminalOpen,
    canCreate,
    canUpdate,
    setFitAddonRef,
    terminalId,
    onFocus,
    onStatusChange,
    exitHandler
  });

  return (
    <div
      ref={paneRef}
      className="flex-1 relative"
      style={{
        minHeight: '200px',
        minWidth: '200px',
        padding: '4px',
        overflow: 'hidden',
        backgroundColor: 'var(--terminal-bg)',
        scrollbarWidth: 'none',
        msOverflowStyle: 'none',
        height: '100%',
        width: '100%',
        position: 'relative'
      }}
      onClick={onFocus}
      onFocus={onFocus}
      tabIndex={0}
    >
      <div
        ref={terminalRef}
        className={cn(
          'absolute inset-0 transition-opacity duration-200',
          isActive ? 'opacity-100' : 'opacity-70'
        )}
        style={{
          padding: '4px',
          height: '100%',
          width: '100%'
        }}
      />
    </div>
  );
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
          <div
            className="w-2 h-2 rounded-full animate-pulse"
            style={{ backgroundColor: 'var(--terminal-status-loading)' }}
          />
        ) : status === 'active' ? (
          <div
            className="w-2 h-2 rounded-full terminal-ready-indicator"
            style={{ backgroundColor: 'var(--terminal-status-active)' }}
          />
        ) : (
          <div
            className="w-2 h-2 rounded-full"
            style={{ backgroundColor: 'var(--terminal-status-idle)' }}
          />
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
            'hover:bg-[var(--terminal-tab-hover)] text-[var(--terminal-text-muted)] hover:text-[var(--terminal-text)]'
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

export const SplitPaneHeader: React.FC<SplitPaneHeaderProps> = ({
  paneIndex,
  isActive,
  canClose,
  totalPanes,
  onFocus,
  onClose,
  closeLabel
}) => {
  const { triangleColor, showTriangle } = useSplitPaneHeader({
    paneIndex,
    isActive,
    totalPanes
  });

  return (
    <div
      className={cn(
        'relative flex items-center justify-between h-6 px-2 cursor-pointer transition-all duration-200',
        'bg-transparent hover:bg-[var(--terminal-tab-hover)]'
      )}
      onClick={onFocus}
    >
      {showTriangle && (
        <div
          className="absolute top-0 left-0 w-0 h-0 z-10"
          style={{
            borderTop: `8px solid ${triangleColor}`,
            borderRight: '8px solid transparent'
          }}
        />
      )}
      <div className="flex-1" />
      <div className="flex items-center">
        {canClose && (
          <button
            className={cn(
              'p-0.5 rounded transition-all duration-200',
              'text-[var(--terminal-text-muted)] hover:text-[var(--terminal-text)] hover:bg-[var(--terminal-tab-hover)]'
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
    <div className="flex h-10 min-h-10 items-center px-2 gap-2 overflow-hidden shrink-0 w-full max-w-full box-border border-b border-[var(--terminal-border)] bg-[var(--terminal-header-bg)]">
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
            'hover:bg-[var(--terminal-close-hover-bg)] hover:text-[var(--terminal-close-hover-text)]',
            'text-[var(--terminal-text-muted)]'
          )}
        >
          <X className="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  );
};

type TerminalProps = {
  isOpen: boolean;
  toggleTerminal: () => void;
  isTerminalOpen: boolean;
  setFitAddonRef: React.Dispatch<React.SetStateAction<any | null>>;
  terminalPosition: 'bottom' | 'right';
  onTogglePosition: () => void;
};

export const Terminal: React.FC<TerminalProps> = ({
  isOpen,
  toggleTerminal,
  isTerminalOpen,
  setFitAddonRef,
  terminalPosition,
  onTogglePosition
}) => {
  const { t } = useTranslation();
  const containerRef = useRef<HTMLDivElement>(null);
  const { canAccessResource } = useRBAC();
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();

  const canCreate = canAccessResource('terminal', 'create');
  const canUpdate = canAccessResource('terminal', 'update');

  const {
    sessions,
    activeSessionId,
    activePaneId,
    sessionStatuses,
    sessionLimit,
    maxSplits,
    splitPanes,
    addSession,
    closeSession,
    switchSession,
    addSplitPane,
    closeSplitPane,
    focusPane,
    getStatusChangeHandler
  } = useTerminalSessions();

  useTerminalStyles();

  // handle closing last session: close session + terminal panel
  const handleCloseSession = useCallback(
    (sessionId: string) => {
      const isLastSession = sessions.length === 1;
      closeSession(sessionId, isLastSession);

      // close last session + close the terminal panel
      if (isLastSession) {
        toggleTerminal();
      }
    },
    [sessions.length, closeSession, toggleTerminal]
  );

  useTerminalKeyboardShortcuts({
    isTerminalOpen,
    activeSessionId,
    activePaneId,
    splitPanesCount: splitPanes.length,
    sessionsCount: sessions.length,
    onCloseSplitPane: closeSplitPane,
    onCloseSession: handleCloseSession,
    onToggleTerminal: toggleTerminal
  });

  if (isFeatureFlagsLoading) {
    return <div className="flex-1 relative overflow-hidden animate-pulse" />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureTerminal)) {
    return <DisabledFeature />;
  }

  return (
    <AnyPermissionGuard
      permissions={['terminal:create', 'terminal:read', 'terminal:update']}
      loadingFallback={<div className="flex-1 relative overflow-hidden animate-pulse" />}
    >
      <div
        className="terminal-container flex h-full w-full flex-col overflow-hidden border-t border-[var(--terminal-border)]"
        ref={containerRef}
        data-slot="terminal"
        style={{
          background: 'var(--terminal-bg)',
          maxWidth: '100%',
          boxSizing: 'border-box',
          contain: 'inline-size'
        }}
      >
        <TerminalHeader
          sessions={sessions}
          activeSessionId={activeSessionId}
          sessionStatuses={sessionStatuses}
          sessionLimit={sessionLimit}
          maxSplits={maxSplits}
          splitPanesCount={splitPanes.length}
          onAddSession={addSession}
          onCloseSession={handleCloseSession}
          onSwitchSession={switchSession}
          onToggleTerminal={toggleTerminal}
          onAddSplitPane={addSplitPane}
          terminalPosition={terminalPosition}
          onTogglePosition={onTogglePosition}
          closeLabel={t('terminal.close')}
          newTabLabel={t('terminal.newTab')}
        />

        <div
          className="flex-1 relative overflow-hidden"
          style={{
            height: '100%',
            width: '100%',
            maxWidth: '100%',
            background: 'var(--terminal-bg)',
            boxSizing: 'border-box',
            contain: 'inline-size'
          }}
        >
          {sessions.map((session) => {
            const isActiveSession = session.id === activeSessionId;
            const hasMultiplePanes = session.splitPanes.length > 1;
            return (
              <div
                key={session.id}
                style={{
                  position: isActiveSession ? 'relative' : 'absolute',
                  visibility: isActiveSession ? 'visible' : 'hidden',
                  height: '100%',
                  width: '100%',
                  top: 0,
                  left: 0,
                  zIndex: isActiveSession ? 1 : 0
                }}
              >
                <ResizablePanelGroup direction="horizontal" className="h-full w-full">
                  {session.splitPanes.map((pane, index) => (
                    <React.Fragment key={pane.id}>
                      {index > 0 && (
                        <ResizableHandle
                          withHandle
                          className="bg-[var(--terminal-split-border)] hover:bg-[var(--terminal-split-active)] transition-colors duration-200 w-[2px] focus-visible:ring-0 focus-visible:ring-offset-0"
                        />
                      )}
                      <ResizablePanel
                        defaultSize={100 / session.splitPanes.length}
                        minSize={20}
                        className="flex flex-col"
                      >
                        {hasMultiplePanes && (
                          <SplitPaneHeader
                            paneIndex={index}
                            isActive={pane.id === activePaneId && isActiveSession}
                            canClose={session.splitPanes.length > 1}
                            totalPanes={session.splitPanes.length}
                            onFocus={() => {
                              if (!isActiveSession) {
                                switchSession(session.id);
                              }
                              focusPane(pane.id);
                            }}
                            onClose={() => closeSplitPane(pane.id)}
                            closeLabel={t('terminal.close')}
                          />
                        )}
                        <div
                          className="flex-1 relative"
                          style={{ height: hasMultiplePanes ? 'calc(100% - 24px)' : '100%' }}
                        >
                          <TerminalPane
                            key={`${session.id}-${pane.terminalId}`}
                            isActive={pane.id === activePaneId && isActiveSession}
                            isTerminalOpen={isTerminalOpen}
                            canCreate={canCreate}
                            canUpdate={canUpdate}
                            setFitAddonRef={setFitAddonRef}
                            terminalId={pane.terminalId}
                            onFocus={() => {
                              if (!isActiveSession) {
                                switchSession(session.id);
                              }
                              focusPane(pane.id);
                            }}
                            onStatusChange={getStatusChangeHandler(pane.terminalId)}
                            exitHandler={{
                              splitPanesCount: splitPanes.length,
                              sessionsCount: sessions.length,
                              activePaneId,
                              activeSessionId,
                              onCloseSplitPane: closeSplitPane,
                              onCloseSession: handleCloseSession,
                              onToggleTerminal: toggleTerminal
                            }}
                          />
                        </div>
                      </ResizablePanel>
                    </React.Fragment>
                  ))}
                </ResizablePanelGroup>
              </div>
            );
          })}
        </div>
      </div>
    </AnyPermissionGuard>
  );
};
