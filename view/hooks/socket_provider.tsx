import { getToken } from '@/lib/auth';
import { WEBSOCKET_URL } from '@/redux/conf';
import { createContext, useContext, useEffect, useRef, useState, ReactNode } from 'react';

type WebSocketContextValue = {
  isReady: boolean;
  message: string | null;
  sendMessage: (data: string) => void;
  sendJsonMessage: (data: any) => void;
};

const WebSocketContext = createContext<WebSocketContextValue>({
  isReady: false,
  message: null,
  sendMessage: () => {},
  sendJsonMessage: () => {}
});

interface WebSocketProviderProps {
  children: ReactNode;
  url?: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

export const WebSocketProvider = ({
  children,
  url = WEBSOCKET_URL + '?token=' + getToken(),
  reconnectInterval = 3000,
  maxReconnectAttempts = 5
}: WebSocketProviderProps) => {
  const [isReady, setIsReady] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const isConnectingRef = useRef(false);
  const isMountedRef = useRef(true);

  const connectWebSocket = () => {
    // Don't attempt to connect if the component is unmounting
    if (!isMountedRef.current) {
      return;
    }

    if (isConnectingRef.current) {
      console.log('Connection already in progress, skipping');
      return;
    }

    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    // Only close the existing connection if it's not in CONNECTING state
    if (wsRef.current) {
      if (wsRef.current.readyState === WebSocket.OPEN) {
        console.log('WebSocket connection already active, skipping connection attempt');
        return;
      }

      if (wsRef.current.readyState !== WebSocket.CONNECTING) {
        try {
          wsRef.current.close();
        } catch (e) {
          console.warn('Error closing existing WebSocket:', e);
        }
        wsRef.current = null;
      } else {
        // If it's connecting, wait for it
        console.log('WebSocket is currently connecting, waiting...');
        return;
      }
    }

    isConnectingRef.current = true;
    console.log('Initiating WebSocket connection...');

    try {
      const socket = new WebSocket(url);

      socket.onopen = () => {
        if (!isMountedRef.current) {
          socket.close();
          return;
        }

        console.log('WebSocket connection established');
        setIsReady(true);
        reconnectAttemptsRef.current = 0;
        isConnectingRef.current = false;
      };

      socket.onclose = (event) => {
        if (!isMountedRef.current) {
          return;
        }

        console.log(`WebSocket connection closed: ${event.code} ${event.reason}`);
        setIsReady(false);
        isConnectingRef.current = false;

        if (!event.wasClean) {
          handleReconnect();
        }
      };

      socket.onmessage = (event) => {
        if (!isMountedRef.current) {
          return;
        }

        setMessage(event.data);
      };

      socket.onerror = (error) => {
        if (!isMountedRef.current) {
          return;
        }

        console.log('WebSocket error:', error);
        isConnectingRef.current = false;

        // Don't try to reconnect here - wait for onclose which will be called after an error
      };

      wsRef.current = socket;
    } catch (error) {
      console.error('Error creating WebSocket:', error);
      isConnectingRef.current = false;

      if (isMountedRef.current) {
        handleReconnect();
      }
    }
  };

  const handleReconnect = () => {
    if (!isMountedRef.current) {
      return;
    }

    if (reconnectAttemptsRef.current < maxReconnectAttempts) {
      console.log(
        `Attempting to reconnect (${reconnectAttemptsRef.current + 1}/${maxReconnectAttempts})...`
      );
      reconnectAttemptsRef.current += 1;

      // Use exponential backoff for reconnection
      const backoffTime = reconnectInterval * Math.pow(1.5, reconnectAttemptsRef.current - 1);

      reconnectTimeoutRef.current = setTimeout(() => {
        if (isMountedRef.current) {
          connectWebSocket();
        }
      }, backoffTime);
    } else {
      console.error(`Failed to reconnect after ${maxReconnectAttempts} attempts`);
    }
  };

  // Reset connection attempt on URL change
  useEffect(() => {
    reconnectAttemptsRef.current = 0;
    isConnectingRef.current = false;

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    connectWebSocket();
  }, [url]);

  // Main setup and cleanup effect
  useEffect(() => {
    isMountedRef.current = true;

    if (!wsRef.current) {
      connectWebSocket();
    }

    return () => {
      isMountedRef.current = false;
      isConnectingRef.current = false;

      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }

      if (wsRef.current) {
        // Remove event handlers to prevent callbacks after unmount
        wsRef.current.onopen = null;
        wsRef.current.onclose = null;
        wsRef.current.onmessage = null;
        wsRef.current.onerror = null;

        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, []);

  const sendMessage = (data: string) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(data);
    } else {
      console.warn('Cannot send message, WebSocket is not connected');
    }
  };

  const sendJsonMessage = (data: any) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data));
    } else {
      console.warn('Cannot send message, WebSocket is not connected');
    }
  };

  const contextValue: WebSocketContextValue = {
    isReady,
    message,
    sendMessage,
    sendJsonMessage
  };

  return <WebSocketContext.Provider value={contextValue}>{children}</WebSocketContext.Provider>;
};

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);

  if (context === undefined) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }

  return context;
};
