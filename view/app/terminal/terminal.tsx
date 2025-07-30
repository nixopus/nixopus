'use client';
import React, { useEffect, useState, useRef, useCallback } from 'react';
import '@xterm/xterm/css/xterm.css';
import { useTerminal } from './utils/useTerminal';
import { useContainerReady } from './utils/isContainerReady';
import { Plus, X } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { useFeatureFlags } from '@/hooks/features_provider';
import DisabledFeature from '@/components/features/disabled-feature';
import Skeleton from '@/app/file-manager/components/skeleton/Skeleton';
import { FeatureNames } from '@/types/feature-flags';
import { AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import { useRBAC } from '@/lib/rbac';
import { Button } from '@/components/ui/button';
import { v4 as uuidv4 } from 'uuid';

const globalStyles = `
  .xterm-viewport::-webkit-scrollbar {
    display: none;
  }
  .xterm-viewport {
    scrollbar-width: none;
    -ms-overflow-style: none;
  }
`;

type TerminalProps = {
  isOpen: boolean;
  toggleTerminal: () => void;
  isTerminalOpen: boolean;
  setFitAddonRef: React.Dispatch<React.SetStateAction<any | null>>;
};

const TerminalSession: React.FC<{
  isActive: boolean;
  isTerminalOpen: boolean;
  dimensions: { width: number; height: number };
  canCreate: boolean;
  canUpdate: boolean;
  setFitAddonRef: React.Dispatch<React.SetStateAction<any | null>>;
  terminalId: string;
}> = ({
  isActive,
  isTerminalOpen,
  dimensions,
  canCreate,
  canUpdate,
  setFitAddonRef,
  terminalId
}) => {
  const { terminalRef, fitAddonRef, initializeTerminal, destroyTerminal } = useTerminal(
    isTerminalOpen && isActive,
    dimensions.width,
    dimensions.height,
    canCreate || canUpdate,
    terminalId
  );
  const isContainerReady = useContainerReady(
    isTerminalOpen && isActive,
    terminalRef as React.RefObject<HTMLDivElement>
  );

  useEffect(() => {
    if (isTerminalOpen && isActive && isContainerReady) {
      initializeTerminal();
    }
  }, [isTerminalOpen, isActive, isContainerReady, initializeTerminal]);

  useEffect(() => {
    if (fitAddonRef) {
      setFitAddonRef(fitAddonRef);
    }
  }, [fitAddonRef, setFitAddonRef]);

  return (
    <div
      ref={terminalRef}
      className="flex-1 relative"
      style={{
        visibility: isTerminalOpen && isActive ? 'visible' : 'hidden',
        minHeight: '200px',
        padding: '4px',
        overflow: 'hidden',
        backgroundColor: '#1e1e1e',
        scrollbarWidth: 'none',
        msOverflowStyle: 'none',
        height: '100%',
        width: '100%'
      }}
    />
  );
};

export const Terminal: React.FC<TerminalProps> = ({
  isOpen,
  toggleTerminal,
  isTerminalOpen,
  setFitAddonRef
}) => {
  const { t } = useTranslation();
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });
  const [sessions, setSessions] = useState([{ id: uuidv4(), label: 'Session 1' }]);
  const [activeSessionId, setActiveSessionId] = useState(sessions[0].id);
  const containerRef = useRef<HTMLDivElement>(null);
  const resizeTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
  const { canAccessResource } = useRBAC();

  const canCreate = canAccessResource('terminal', 'create');
  const canUpdate = canAccessResource('terminal', 'update');
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();
  const SESSION_LIMIT = 3;

  const updateDimensions = useCallback(() => {
    if (!containerRef.current) return;

    if (resizeTimeoutRef.current) {
      clearTimeout(resizeTimeoutRef.current);
    }

    resizeTimeoutRef.current = setTimeout(() => {
      if (containerRef.current) {
        setDimensions({
          width: containerRef.current.offsetWidth,
          height: containerRef.current.offsetHeight
        });
      }
    }, 100);
  }, []);

  useEffect(() => {
    if (!containerRef.current) return;

    updateDimensions();

    const resizeObserver = new ResizeObserver(updateDimensions);
    resizeObserver.observe(containerRef.current);

    return () => {
      resizeObserver.disconnect();
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
      }
    };
  }, [isTerminalOpen, updateDimensions]);

  useEffect(() => {
    const style = document.createElement('style');
    style.textContent = globalStyles;
    document.head.appendChild(style);
    return () => {
      document.head.removeChild(style);
    };
  }, []);

  const addSession = () => {
    if (sessions.length >= SESSION_LIMIT) {
      return;
    }
    const newSession = {
      id: uuidv4(),
      label: `Session ${sessions.length + 1}`
    };
    setSessions((prev) => [...prev, newSession]);
    setActiveSessionId(newSession.id);
  };

  const closeSession = (id: string) => {
    setSessions((prev) => {
      const idx = prev.findIndex((s) => s.id === id);
      const newSessions = prev.filter((s) => s.id !== id);
      if (id === activeSessionId && newSessions.length > 0) {
        setActiveSessionId(newSessions[Math.max(0, idx - 1)].id);
      }
      return newSessions;
    });
  };

  const switchSession = (id: string) => {
    setActiveSessionId(id);
  };

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
        className="flex h-full flex-col overflow-hidden bg-[#1e1e1e]"
        ref={containerRef}
        data-slot="terminal"
      >
        <div className="flex h-8 items-center justify-between border-b border-[#2d2d2d] px-3">
          <div className="flex items-center gap-2">
            <span className="text-xs font-medium text-[#cccccc]">{t('terminal.title')}</span>
            <span className="text-xs text-[#666666]">{t('terminal.shortcut')}</span>
          </div>
          <div className="flex items-center gap-2 ml-auto">
            {sessions.map((session) => (
              <div
                key={session.id}
                className={`flex items-center px-2 py-1 rounded-t-md cursor-pointer ${
                  session.id === activeSessionId
                    ? 'bg-[#232323] border border-[#333]'
                    : 'bg-transparent'
                }`}
                onClick={() => switchSession(session.id)}
                style={{ marginLeft: 4 }}
              >
                <span className="text-xs text-[#cccccc] mr-1">{session.label}</span>
                {sessions.length > 1 && (
                  <button
                    className="ml-1 text-[#666] hover:text-[#ccc]"
                    onClick={(e) => {
                      e.stopPropagation();
                      closeSession(session.id);
                    }}
                    title={t('terminal.close')}
                  >
                    <X className="h-3 w-3" />
                  </button>
                )}
              </div>
            ))}
            {sessions.length < SESSION_LIMIT && (
              <button
                className="ml-2 text-[#666] hover:text-[#ccc]"
                onClick={addSession}
                title={t('terminal.newTab')}
              >
                <Plus className="h-3 w-3" />
              </button>
            )}
            <Button
              variant="ghost"
              size="icon"
              onClick={toggleTerminal}
              title={t('terminal.close')}
            >
              <X className="h-3 w-3 text-[#666666] hover:text-[#cccccc]" />
            </Button>
          </div>
        </div>
        <div className="flex-1 relative" style={{ height: '100%', width: '100%' }}>
          {sessions.map((session) => (
            <div
              key={session.id}
              style={{ display: session.id === activeSessionId ? 'block' : 'none', height: '100%' }}
            >
              <TerminalSession
                isActive={session.id === activeSessionId}
                isTerminalOpen={isTerminalOpen}
                dimensions={dimensions}
                canCreate={canCreate}
                canUpdate={canUpdate}
                setFitAddonRef={setFitAddonRef}
                terminalId={session.id}
              />
            </div>
          ))}
        </div>
      </div>
    </AnyPermissionGuard>
  );
};
