'use client';

import React, { useRef, useCallback } from 'react';
import '@xterm/xterm/css/xterm.css';
import { useTranslation } from '@/hooks/use-translation';
import { useFeatureFlags } from '@/hooks/features_provider';
import DisabledFeature from '@/components/features/disabled-feature';
import { FeatureNames } from '@/packages/types/feature-flags';
import { AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import { useRBAC } from '@/lib/rbac';
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from '@/components/ui/resizable';

import { TerminalHeader, TerminalPane, SplitPaneHeader } from './components';
import {
  useTerminalSessions,
  useTerminalDimensions,
  useTerminalStyles,
  useTerminalKeyboardShortcuts
} from './hooks';

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

  const dimensions = useTerminalDimensions(containerRef, isTerminalOpen);

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
                          className="bg-[#3a3a3a] hover:bg-[#4a4a4a] transition-colors duration-200 w-[2px] focus-visible:ring-0 focus-visible:ring-offset-0"
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
