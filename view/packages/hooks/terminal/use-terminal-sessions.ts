import { useState, useCallback, useRef, useMemo, useEffect } from 'react';
import { v4 as uuidv4 } from 'uuid';
import type { SessionStatus, SplitPane, Session } from '../../types/terminal';

const SESSION_LIMIT = 5;
const MAX_SPLITS = 4;

const initialSessionId = uuidv4();
const initialPaneId = uuidv4();

export const useTerminalSessions = () => {
  const [sessions, setSessions] = useState<Session[]>([
    {
      id: initialSessionId,
      label: 'Terminal 1',
      splitPanes: [{ id: initialPaneId, label: 'Pane 1', terminalId: uuidv4() }]
    }
  ]);
  const [activeSessionId, setActiveSessionId] = useState(initialSessionId);
  const [activePaneId, setActivePaneId] = useState(initialPaneId);
  const [paneStatuses, setPaneStatuses] = useState<Record<string, SessionStatus>>({});
  const statusChangeHandlers = useRef<Record<string, (status: SessionStatus) => void>>({});

  // Track active pane per session
  const [activePaneBySession, setActivePaneBySession] = useState<Record<string, string>>({
    [initialSessionId]: initialPaneId
  });

  // Compute session-level statuses from pane statuses
  const sessionStatuses = useMemo(() => {
    const statuses: Record<string, SessionStatus> = {};

    sessions.forEach((session) => {
      const paneStatusList = session.splitPanes.map(
        (pane) => paneStatuses[pane.terminalId] || 'idle'
      );

      // Session is 'loading' if any pane is loading
      if (paneStatusList.includes('loading')) {
        statuses[session.id] = 'loading';
      }
      // Session is 'active' if any pane is active
      else if (paneStatusList.includes('active')) {
        statuses[session.id] = 'active';
      }
      // Otherwise session is 'idle'
      else {
        statuses[session.id] = 'idle';
      }
    });

    return statuses;
  }, [sessions, paneStatuses]);

  useEffect(() => {
    if (sessions.length === 0) {
      const newSessionId = uuidv4();
      const newPaneId = uuidv4();
      const newTerminalId = uuidv4();
      const newSession: Session = {
        id: newSessionId,
        label: 'Terminal 1',
        splitPanes: [{ id: newPaneId, label: 'Pane 1', terminalId: newTerminalId }]
      };
      setSessions([newSession]);
      setActiveSessionId(newSessionId);
      setActivePaneId(newPaneId);
      setActivePaneBySession({ [newSessionId]: newPaneId });
      setPaneStatuses({ [newTerminalId]: 'loading' });
    }
  }, [sessions.length]);

  const addSession = useCallback(() => {
    if (sessions.length >= SESSION_LIMIT) {
      return;
    }
    const newSessionId = uuidv4();
    const newPaneId = uuidv4();
    const newTerminalId = uuidv4();
    const newSession: Session = {
      id: newSessionId,
      label: `Terminal ${sessions.length + 1}`,
      splitPanes: [{ id: newPaneId, label: 'Pane 1', terminalId: newTerminalId }]
    };
    setSessions((prev) => [...prev, newSession]);
    setActiveSessionId(newSessionId);
    setActivePaneId(newPaneId);
    setActivePaneBySession((prev) => ({
      ...prev,
      [newSessionId]: newPaneId
    }));
    setPaneStatuses((prev) => ({ ...prev, [newTerminalId]: 'loading' }));
  }, [sessions.length]);

  const closeSession = useCallback(
    (id: string, force: boolean = false) => {
      setSessions((prev) => {
        // Prevent closing last session unless forced
        if (!force && prev.length <= 1) {
          return prev;
        }

        const idx = prev.findIndex((s) => s.id === id);
        const closedSession = prev.find((s) => s.id === id);
        const newSessions = prev.filter((s) => s.id !== id);

        // Clean up pane statuses for closed session
        if (closedSession) {
          setPaneStatuses((prevStatuses) => {
            const updated = { ...prevStatuses };
            closedSession.splitPanes.forEach((pane) => {
              delete updated[pane.terminalId];
              delete statusChangeHandlers.current[pane.terminalId];
            });
            return updated;
          });
        }

        if (id === activeSessionId) {
          if (newSessions.length > 0) {
            const newActiveSession = newSessions[Math.max(0, idx - 1)];
            const newActivePaneId =
              activePaneBySession[newActiveSession.id] || newActiveSession.splitPanes[0]?.id || '';
            setActiveSessionId(newActiveSession.id);
            setActivePaneId(newActivePaneId);
          } else {
            setActiveSessionId('');
            setActivePaneId('');
          }
        }
        return newSessions;
      });
      setActivePaneBySession((prev) => {
        const updated = { ...prev };
        delete updated[id];
        return updated;
      });
    },
    [activeSessionId, activePaneBySession]
  );

  const switchSession = useCallback(
    (id: string) => {
      const session = sessions.find((s) => s.id === id);
      if (session) {
        setActiveSessionId(id);
        // Restore the active pane for this session, or use the first one
        const savedActivePane = activePaneBySession[id] || session.splitPanes[0]?.id || '';
        setActivePaneId(savedActivePane);
      }
    },
    [sessions, activePaneBySession]
  );

  const addSplitPane = useCallback(() => {
    const activeSession = sessions.find((s) => s.id === activeSessionId);
    if (!activeSession || activeSession.splitPanes.length >= MAX_SPLITS) {
      return;
    }

    const newPaneId = uuidv4();
    const newTerminalId = uuidv4();
    const newPane: SplitPane = {
      id: newPaneId,
      label: `Pane ${activeSession.splitPanes.length + 1}`,
      terminalId: newTerminalId
    };

    setSessions((prev) =>
      prev.map((session) =>
        session.id === activeSessionId
          ? { ...session, splitPanes: [...session.splitPanes, newPane] }
          : session
      )
    );
    setActivePaneId(newPaneId);
    setActivePaneBySession((prev) => ({
      ...prev,
      [activeSessionId]: newPaneId
    }));
    setPaneStatuses((prev) => ({ ...prev, [newTerminalId]: 'loading' }));
  }, [sessions, activeSessionId]);

  const closeSplitPane = useCallback(
    (paneId: string) => {
      const activeSession = sessions.find((s) => s.id === activeSessionId);
      if (!activeSession || activeSession.splitPanes.length <= 1) {
        return;
      }

      setSessions((prev) =>
        prev.map((session) => {
          if (session.id === activeSessionId) {
            const idx = session.splitPanes.findIndex((p) => p.id === paneId);
            const paneToClose = session.splitPanes.find((p) => p.id === paneId);
            const newPanes = session.splitPanes.filter((p) => p.id !== paneId);

            // Clean up pane status
            if (paneToClose) {
              setPaneStatuses((prevStatuses) => {
                const updated = { ...prevStatuses };
                delete updated[paneToClose.terminalId];
                delete statusChangeHandlers.current[paneToClose.terminalId];
                return updated;
              });
            }

            if (paneId === activePaneId && newPanes.length > 0) {
              const newActivePaneId = newPanes[Math.max(0, idx - 1)].id;
              setActivePaneId(newActivePaneId);
              setActivePaneBySession((prevMapping) => ({
                ...prevMapping,
                [activeSessionId]: newActivePaneId
              }));
            }
            return { ...session, splitPanes: newPanes };
          }
          return session;
        })
      );
    },
    [sessions, activeSessionId, activePaneId]
  );

  const focusPane = useCallback(
    (paneId: string) => {
      setActivePaneId(paneId);
      setActivePaneBySession((prev) => ({
        ...prev,
        [activeSessionId]: paneId
      }));
    },
    [activeSessionId]
  );

  const getStatusChangeHandler = useCallback((terminalId: string) => {
    if (!statusChangeHandlers.current[terminalId]) {
      statusChangeHandlers.current[terminalId] = (status: SessionStatus) => {
        setPaneStatuses((prev) => {
          if (prev[terminalId] === status) return prev;
          return { ...prev, [terminalId]: status };
        });
      };
    }
    return statusChangeHandlers.current[terminalId];
  }, []);

  // Get current active session's split panes
  const activeSession = sessions.find((s) => s.id === activeSessionId);
  const splitPanes = activeSession?.splitPanes || [];
  const activePaneIdForSession = activePaneBySession[activeSessionId] || splitPanes[0]?.id || '';

  return {
    sessions,
    activeSessionId,
    activePaneId: activePaneIdForSession,
    sessionStatuses,
    sessionLimit: SESSION_LIMIT,
    maxSplits: MAX_SPLITS,
    splitPanes,
    addSession,
    closeSession,
    switchSession,
    addSplitPane,
    closeSplitPane,
    focusPane,
    getStatusChangeHandler
  };
};
