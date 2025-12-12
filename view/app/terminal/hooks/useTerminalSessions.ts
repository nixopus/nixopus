import { useState, useCallback, useRef } from 'react';
import { v4 as uuidv4 } from 'uuid';
import type { SessionStatus } from '../components/TerminalSession';

type Session = {
  id: string;
  label: string;
};

const SESSION_LIMIT = 5;

export const useTerminalSessions = () => {
  const [sessions, setSessions] = useState<Session[]>([{ id: uuidv4(), label: 'Terminal 1' }]);
  const [activeSessionId, setActiveSessionId] = useState(sessions[0].id);
  const [sessionStatuses, setSessionStatuses] = useState<Record<string, SessionStatus>>({});
  const statusChangeHandlers = useRef<Record<string, (status: SessionStatus) => void>>({});

  const addSession = useCallback(() => {
    if (sessions.length >= SESSION_LIMIT) {
      return;
    }
    const newSession = {
      id: uuidv4(),
      label: `Terminal ${sessions.length + 1}`
    };
    setSessions((prev) => [...prev, newSession]);
    setActiveSessionId(newSession.id);
    setSessionStatuses((prev) => ({ ...prev, [newSession.id]: 'loading' }));
  }, [sessions.length]);

  const closeSession = useCallback(
    (id: string) => {
      setSessions((prev) => {
        const idx = prev.findIndex((s) => s.id === id);
        const newSessions = prev.filter((s) => s.id !== id);
        if (id === activeSessionId && newSessions.length > 0) {
          setActiveSessionId(newSessions[Math.max(0, idx - 1)].id);
        }
        return newSessions;
      });
      setSessionStatuses((prev) => {
        const newStatuses = { ...prev };
        delete newStatuses[id];
        return newStatuses;
      });
      delete statusChangeHandlers.current[id];
    },
    [activeSessionId]
  );

  const switchSession = useCallback((id: string) => {
    setActiveSessionId(id);
  }, []);

  const getStatusChangeHandler = useCallback((sessionId: string) => {
    if (!statusChangeHandlers.current[sessionId]) {
      statusChangeHandlers.current[sessionId] = (status: SessionStatus) => {
        setSessionStatuses((prev) => {
          if (prev[sessionId] === status) return prev;
          return { ...prev, [sessionId]: status };
        });
      };
    }
    return statusChangeHandlers.current[sessionId];
  }, []);

  return {
    sessions,
    activeSessionId,
    sessionStatuses,
    sessionLimit: SESSION_LIMIT,
    addSession,
    closeSession,
    switchSession,
    getStatusChangeHandler
  };
};
