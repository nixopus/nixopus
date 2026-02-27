import { useEffect, useMemo, useRef } from 'react';
import { v4 as uuidv4 } from 'uuid';
import { useTerminal } from '@/packages/hooks/terminal/use-terminal';
import { useContainerReady } from '@/packages/hooks/terminal/use-container-ready';
import { useWebSocket } from '@/packages/hooks/shared/socket-provider';

export const useContainerTerminal = (containerId: string) => {
  const sessionId = useMemo(() => `container-${containerId}-${uuidv4()}`, [containerId]);
  const { sendJsonMessage, isReady } = useWebSocket();

  const {
    terminalRef: termRef,
    initializeTerminal,
    terminalInstance
  } = useTerminal(true, 0, 0, true, sessionId);

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

      const dockerCmd = `if docker ps >/dev/null 2>&1; then D=docker; else D='sudo -n docker'; fi; $D exec -it ${containerId} /bin/bash || $D exec -it ${containerId} /bin/sh`;
      const maxRetries = 3;
      const retryDelay = 500;
      const initialDelay = 150;
      const clearDelay = 1800;

      // Send docker exec command with retries
      for (let i = 0; i < maxRetries; i++) {
        setTimeout(
          () => {
            sendJsonMessage({
              action: 'terminal',
              data: { value: `${dockerCmd}\r`, terminalId: sessionId }
            });
          },
          initialDelay + i * retryDelay
        );
      }

      // Clear terminal after connection is established
      setTimeout(
        () => {
          sendJsonMessage({
            action: 'terminal',
            data: { value: 'clear\r', terminalId: sessionId }
          });
        },
        initialDelay + (maxRetries - 1) * retryDelay + clearDelay
      );
    }
  }, [terminalInstance, isReady, sendJsonMessage, containerId, sessionId]);

  return {
    terminalRef: termRef,
    sessionId
  };
};
