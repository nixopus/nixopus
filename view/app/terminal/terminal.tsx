'use client';

import React, { useRef } from 'react';
import '@xterm/xterm/css/xterm.css';
import { useTranslation } from '@/hooks/use-translation';
import { useFeatureFlags } from '@/hooks/features_provider';
import DisabledFeature from '@/components/features/disabled-feature';
import Skeleton from '@/app/file-manager/components/skeleton/Skeleton';
import { FeatureNames } from '@/types/feature-flags';
import { AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import { useRBAC } from '@/lib/rbac';

import { TerminalSession, TerminalHeader } from './components';
import { useTerminalSessions, useTerminalDimensions, useTerminalStyles } from './hooks';

type TerminalProps = {
  isOpen: boolean;
  toggleTerminal: () => void;
  isTerminalOpen: boolean;
  setFitAddonRef: React.Dispatch<React.SetStateAction<any | null>>;
};

export const Terminal: React.FC<TerminalProps> = ({
  isOpen,
  toggleTerminal,
  isTerminalOpen,
  setFitAddonRef
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
    sessionStatuses,
    sessionLimit,
    addSession,
    closeSession,
    switchSession,
    getStatusChangeHandler
  } = useTerminalSessions();

  const dimensions = useTerminalDimensions(containerRef, isTerminalOpen);

  useTerminalStyles();

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureTerminal)) {
    return <DisabledFeature />;
  }

  return (
    <AnyPermissionGuard
      permissions={['terminal:create', 'terminal:read', 'terminal:update']}
      loadingFallback={<Skeleton />}
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
          onAddSession={addSession}
          onCloseSession={closeSession}
          onSwitchSession={switchSession}
          onToggleTerminal={toggleTerminal}
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
          {sessions.map((session) => (
            <div
              key={session.id}
              style={{
                display: session.id === activeSessionId ? 'block' : 'none',
                height: '100%',
                width: '100%',
                maxWidth: '100%',
                overflow: 'hidden',
                contain: 'inline-size'
              }}
            >
              <TerminalSession
                isActive={session.id === activeSessionId}
                isTerminalOpen={isTerminalOpen}
                dimensions={dimensions}
                canCreate={canCreate}
                canUpdate={canUpdate}
                setFitAddonRef={setFitAddonRef}
                terminalId={session.id}
                onStatusChange={getStatusChangeHandler(session.id)}
              />
            </div>
          ))}
        </div>
      </div>
    </AnyPermissionGuard>
  );
};
