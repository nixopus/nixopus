'use client';
import React, { useEffect, useState, useRef, useCallback } from 'react';
import '@xterm/xterm/css/xterm.css';
import { useTerminal } from './utils/useTerminal';
import { useContainerReady } from './utils/isContainerReady';
import { Plus, X, SplitSquareVertical, Maximize2, Minimize2 } from 'lucide-react';
import { useTranslation, translationKey } from '@/hooks/use-translation';
import { useFeatureFlags } from '@/hooks/features_provider';
import DisabledFeature from '@/components/features/disabled-feature';
import Skeleton from '@/app/file-manager/components/skeleton/Skeleton';
import { FeatureNames } from '@/types/feature-flags';
import { AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import { useRBAC } from '@/lib/rbac';
import { Button } from '@/components/ui/button';
import { v4 as uuidv4 } from 'uuid';
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from '@/components/ui/resizable';

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

type SplitPane = {
  id: string;
  label: string;
  terminalId: string;
};

type Session = {
  id: string;
  label: string;
  splitPanes: SplitPane[];
};

const TerminalSession: React.FC<{
  isActive: boolean;
  isTerminalOpen: boolean;
  canCreate: boolean;
  canUpdate: boolean;
  setFitAddonRef: React.Dispatch<React.SetStateAction<any | null>>;
  terminalId: string;
  onFocus: () => void;
  containerRef?: React.RefObject<HTMLDivElement>;
}> = ({
  isActive,
  isTerminalOpen,
  canCreate,
  canUpdate,
  setFitAddonRef,
  terminalId,
  onFocus,
  containerRef
}) => {
  const paneRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });
  const resizeTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);

  const updateDimensions = useCallback(() => {
    if (!paneRef.current) return;

    if (resizeTimeoutRef.current) {
      clearTimeout(resizeTimeoutRef.current);
    }

    resizeTimeoutRef.current = setTimeout(() => {
      if (paneRef.current) {
        setDimensions({
          width: paneRef.current.offsetWidth,
          height: paneRef.current.offsetHeight
        });
      }
    }, 100);
  }, []);

  useEffect(() => {
    if (!paneRef.current) return;

    // Force immediate dimension check
    const immediateCheck = () => {
      if (paneRef.current) {
        const rect = paneRef.current.getBoundingClientRect();
        if (rect.width > 0 && rect.height > 0) {
          setDimensions({
            width: rect.width,
            height: rect.height
          });
        }
      }
    };

    // Check immediately
    immediateCheck();

    // Also use the callback-based update
    updateDimensions();

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        if (entry.target === paneRef.current) {
          updateDimensions();
        }
      }
    });
    resizeObserver.observe(paneRef.current);

    // Multiple delayed checks to catch cases where ResizeObserver doesn't fire immediately
    const delayedChecks = [100, 200, 500].map((delay) =>
      setTimeout(() => {
        immediateCheck();
        updateDimensions();
      }, delay)
    );

    return () => {
      resizeObserver.disconnect();
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
      }
      delayedChecks.forEach((timeout) => clearTimeout(timeout));
    };
  }, [isTerminalOpen, updateDimensions]);

  const { terminalRef, fitAddonRef, initializeTerminal, destroyTerminal, isWebSocketReady } = useTerminal(
    isTerminalOpen,
    dimensions.width,
    dimensions.height,
    canCreate || canUpdate,
    terminalId
  );
  const isContainerReady = useContainerReady(
    isTerminalOpen,
    terminalRef as React.RefObject<HTMLDivElement>
  );

  useEffect(() => {
    if (!isTerminalOpen || !isWebSocketReady) return;

    let initialized = false;

    const attemptInitialization = () => {
      // Prevent multiple initializations
      if (initialized) return true;

      // Check if terminalRef is attached
      if (!terminalRef?.current) return false;

      // Try to get dimensions from paneRef if state dimensions are 0
      let finalWidth = dimensions.width;
      let finalHeight = dimensions.height;
      
      if (finalWidth === 0 || finalHeight === 0) {
        if (paneRef.current) {
          const rect = paneRef.current.getBoundingClientRect();
          if (rect.width > 0 && rect.height > 0) {
            finalWidth = rect.width;
            finalHeight = rect.height;
            // Update dimensions state immediately
            setDimensions({
              width: finalWidth,
              height: finalHeight
            });
          }
        }
      }

      // Also check terminalRef dimensions as fallback
      if (finalWidth === 0 || finalHeight === 0) {
        if (terminalRef.current) {
          const rect = terminalRef.current.getBoundingClientRect();
          if (rect.width > 0 && rect.height > 0) {
            finalWidth = rect.width;
            finalHeight = rect.height;
          }
        }
      }

      // Initialize if we have valid dimensions and container is ready
      if (finalWidth > 0 && finalHeight > 0 && (isContainerReady || terminalRef.current)) {
        initializeTerminal();
        initialized = true;
        return true;
      }
      return false;
    };

    // Try immediately
    if (attemptInitialization()) {
      return;
    }

    // Retry with delays to handle async updates
    const retryDelays = [50, 100, 200, 500];
    const timeouts: NodeJS.Timeout[] = [];

    retryDelays.forEach((delay) => {
      const timeout = setTimeout(() => {
        if (attemptInitialization()) {
          // Clear remaining timeouts if initialization succeeds
          timeouts.forEach((t) => clearTimeout(t));
        }
      }, delay);
      timeouts.push(timeout);
    });

    return () => {
      timeouts.forEach((timeout) => clearTimeout(timeout));
    };
  }, [isTerminalOpen, isContainerReady, initializeTerminal, dimensions.width, dimensions.height, isWebSocketReady]);

  // Cleanup: destroy terminal when component unmounts
  useEffect(() => {
    return () => {
      destroyTerminal();
    };
  }, [destroyTerminal]);

  // Re-fit terminal when dimensions change - but only if WebSocket is ready
  useEffect(() => {
    if (fitAddonRef?.current && dimensions.width > 0 && dimensions.height > 0 && isWebSocketReady) {
      requestAnimationFrame(() => {
        try {
          fitAddonRef.current?.fit();
        } catch (error) {
          // Ignore fit errors
        }
      });
    }
  }, [fitAddonRef, dimensions.width, dimensions.height, isWebSocketReady]);

  useEffect(() => {
    if (fitAddonRef) {
      setFitAddonRef(fitAddonRef);
    }
  }, [fitAddonRef, setFitAddonRef]);

  return (
    <div
      ref={paneRef}
      className="flex-1 relative"
      style={{
        minHeight: '200px',
        minWidth: '200px',
        padding: '4px',
        overflow: 'hidden',
        backgroundColor: '#1e1e1e',
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
        className="absolute inset-0"
        style={{
          padding: '4px',
          height: '100%',
          width: '100%'
        }}
      />
    </div>
  );
};

const SplitPaneHeader: React.FC<{
  pane: SplitPane;
  isActive: boolean;
  canClose: boolean;
  onFocus: () => void;
  onClose: () => void;
  t: (key: translationKey, params?: Record<string, string>) => string;
}> = ({ pane, isActive, canClose, onFocus, onClose, t }) => {
  return (
    <div
      className={`flex items-center justify-between h-7 px-2 border-r cursor-pointer ${
        isActive
          ? 'bg-[#2d2d2d] border-[#007acc]'
          : 'bg-[#1e1e1e] border-[#2d2d2d] hover:bg-[#252525]'
      }`}
      onClick={onFocus}
    >
      <span className="text-xs text-[#cccccc]">{pane.label}</span>
      <div className="flex items-center gap-1">
        {canClose && (
          <button
            className="p-1 text-[#666] hover:text-[#ccc] rounded"
            onClick={(e) => {
              e.stopPropagation();
              onClose();
            }}
            title={t('terminal.close')}
          >
            <X className="h-3 w-3" />
          </button>
        )}
      </div>
    </div>
  );
};

export const Terminal: React.FC<TerminalProps> = ({
  isOpen,
  toggleTerminal,
  isTerminalOpen,
  setFitAddonRef
}) => {
  const { t } = useTranslation();
  const initialSessionId = uuidv4();
  const initialPaneId = uuidv4();
  const [sessions, setSessions] = useState<Session[]>([
    {
      id: initialSessionId,
      label: 'Session 1',
      splitPanes: [{ id: initialPaneId, label: 'Terminal 1', terminalId: uuidv4() }]
    }
  ]);
  const [activeSessionId, setActiveSessionId] = useState<string>(initialSessionId);
  const [activePaneId, setActivePaneId] = useState<string>(initialPaneId);
  const containerRef = useRef<HTMLDivElement>(null);

  // Track active pane per session
  const [activePaneBySession, setActivePaneBySession] = useState<Record<string, string>>({
    [initialSessionId]: initialPaneId
  });

  // Get current active session's split panes
  const activeSession = sessions.find((s) => s.id === activeSessionId);
  const splitPanes = activeSession?.splitPanes || [];
  const activePaneIdForSession = activePaneBySession[activeSessionId] || splitPanes[0]?.id || '';
  const { canAccessResource } = useRBAC();

  const canCreate = canAccessResource('terminal', 'create');
  const canUpdate = canAccessResource('terminal', 'update');
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();
  const SESSION_LIMIT = 3;
  const MAX_SPLITS = 3;

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
    const newPaneId = uuidv4();
    const newSessionId = uuidv4();
    const newSession: Session = {
      id: newSessionId,
      label: `Session ${sessions.length + 1}`,
      splitPanes: [{ id: newPaneId, label: 'Terminal 1', terminalId: uuidv4() }]
    };
    setSessions((prev) => [...prev, newSession]);
    setActiveSessionId(newSessionId);
    setActivePaneId(newPaneId);
    setActivePaneBySession((prev) => ({
      ...prev,
      [newSessionId]: newPaneId
    }));
  };

  const closeSession = (id: string) => {
    setSessions((prev) => {
      const idx = prev.findIndex((s) => s.id === id);
      const newSessions = prev.filter((s) => s.id !== id);
      if (id === activeSessionId && newSessions.length > 0) {
        const newActiveSession = newSessions[Math.max(0, idx - 1)];
        const newActivePaneId = activePaneBySession[newActiveSession.id] || newActiveSession.splitPanes[0]?.id || '';
        setActiveSessionId(newActiveSession.id);
        setActivePaneId(newActivePaneId);
      }
      return newSessions;
    });
    // Clean up active pane tracking for closed session
    setActivePaneBySession((prev) => {
      const updated = { ...prev };
      delete updated[id];
      return updated;
    });
  };

  const switchSession = (id: string) => {
    const session = sessions.find((s) => s.id === id);
    if (session) {
      setActiveSessionId(id);
      // Restore the active pane for this session, or use the first one
      const savedActivePane = activePaneBySession[id] || session.splitPanes[0]?.id || '';
      setActivePaneId(savedActivePane);
    }
  };

  const addSplitPane = () => {
    if (splitPanes.length >= MAX_SPLITS) return;
    
    const newPane: SplitPane = {
      id: uuidv4(),
      label: `Terminal ${splitPanes.length + 1}`,
      terminalId: uuidv4()
    };
    setSessions((prev) =>
      prev.map((session) =>
        session.id === activeSessionId
          ? { ...session, splitPanes: [...session.splitPanes, newPane] }
          : session
      )
    );
    setActivePaneId(newPane.id);
  };

  const closeSplitPane = (paneId: string) => {
    if (splitPanes.length <= 1) return;
    setSessions((prev) =>
      prev.map((session) => {
        if (session.id === activeSessionId) {
          const idx = session.splitPanes.findIndex((p) => p.id === paneId);
          const newPanes = session.splitPanes.filter((p) => p.id !== paneId);
          if (paneId === activePaneId && newPanes.length > 0) {
            setActivePaneId(newPanes[Math.max(0, idx - 1)].id);
          }
          return { ...session, splitPanes: newPanes };
        }
        return session;
      })
    );
  };

  const focusPane = (paneId: string) => {
    setActivePaneId(paneId);
    // Save the active pane for the current session
    setActivePaneBySession((prev) => ({
      ...prev,
      [activeSessionId]: paneId
    }));
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
            {splitPanes.length < MAX_SPLITS && (
              <button
                className="ml-2 text-[#666] hover:text-[#ccc]"
                onClick={addSplitPane}
                title="Split Terminal"
              >
                <SplitSquareVertical className="h-3 w-3" />
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
        <div className="flex-1 relative overflow-hidden" style={{ height: '100%', width: '100%' }}>
          {sessions.map((session) => {
            const isActiveSession = session.id === activeSessionId;
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
                <ResizablePanelGroup direction="horizontal" className="h-full">
                  {session.splitPanes.map((pane, index) => (
                    <React.Fragment key={pane.id}>
                      <ResizablePanel
                        defaultSize={100 / session.splitPanes.length}
                        minSize={20}
                        className="flex flex-col"
                      >
                        <SplitPaneHeader
                          pane={pane}
                          isActive={pane.id === activePaneIdForSession && isActiveSession}
                          canClose={session.splitPanes.length > 1}
                          onFocus={() => {
                            if (!isActiveSession) {
                              switchSession(session.id);
                            }
                            focusPane(pane.id);
                          }}
                          onClose={() => closeSplitPane(pane.id)}
                          t={t}
                        />
                        <div className="flex-1 relative" style={{ height: 'calc(100% - 28px)' }}>
                          <TerminalSession
                            key={`${session.id}-${pane.terminalId}`}
                            isActive={pane.id === activePaneIdForSession && isActiveSession}
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
                            containerRef={containerRef as React.RefObject<HTMLDivElement>}
                          />
                        </div>
                      </ResizablePanel>
                      {index < session.splitPanes.length - 1 && (
                        <ResizableHandle withHandle className="bg-[#2d2d2d] hover:bg-[#3d3d3d]" />
                      )}
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
