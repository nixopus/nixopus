'use client';
import React, { useEffect, useMemo, useRef } from 'react';
import '@xterm/xterm/css/xterm.css';
import { v4 as uuidv4 } from 'uuid';
import { useTerminal } from '@/app/terminal/utils/useTerminal';
import { useContainerReady } from '@/app/terminal/utils/isContainerReady';
import { useWebSocket } from '@/hooks/socket-provider';

type TerminalProps = {
  containerId: string;
};

export const Terminal: React.FC<TerminalProps> = ({ containerId }) => {
  const terminalRef = useRef<HTMLDivElement | null>(null);
  const sessionId = useMemo(() => `container-${containerId}-${uuidv4()}`, [containerId]);
  const { sendJsonMessage, isReady } = useWebSocket();

  const { terminalRef: termRef, initializeTerminal, terminalInstance } = useTerminal(
    true,
    0,
    0,
    true,
    sessionId
  );

  const isMounted = useContainerReady(true, termRef as React.RefObject<HTMLDivElement>);

  useEffect(() => {
    if (isMounted) {
      initializeTerminal();
    }
  }, [isMounted, initializeTerminal]);

  const hasSentInitRef = useRef(false);
  useEffect(() => {
    if (!hasSentInitRef.current && terminalInstance && isReady) {
      hasSentInitRef.current = true;
      // TODO: optimize this such that backend handles this instead of client.
      const cmd = `if docker ps >/dev/null 2>&1; then D=docker; else D='sudo -n docker'; fi; $D exec -it ${containerId} /bin/bash || $D exec -it ${containerId} /bin/sh\r`;
      setTimeout(() => {
        sendJsonMessage({ action: 'terminal', data: { value: cmd, terminalId: sessionId } });
      }, 150);
    }
  }, [terminalInstance, isReady, sendJsonMessage, containerId, sessionId]);

  return (
    <div
      ref={(el) => {
        terminalRef.current = el;
        // @ts-ignore
        if (termRef) termRef.current = el;
      }}
      className="relative"
      style={{ height: '60vh', minHeight: 300, backgroundColor: '#1e1e1e' }}
    />
  );
};

export default Terminal;


