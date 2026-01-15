'use client';
import { getWebsocketUrl } from '@/redux/conf';
import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  ReactNode,
  useCallback
} from 'react';
import { getAccessToken } from 'supertokens-auth-react/recipe/session';
import { useAppSelector } from '@/redux/hooks';
import { getAdvancedSettings } from '@/packages/utils/advanced-settings';

type WebSocketContextValue = {
  isReady: boolean;
  message: string | null;
  sendMessage: (data: string) => void;
  sendJsonMessage: (data: any) => void;
  subscribe: (listener: (data: string) => void) => () => void;
};

const WebSocketContext = createContext<WebSocketContextValue>({
  isReady: false,
  message: null,
  sendMessage: () => {},
  sendJsonMessage: () => {},
  subscribe: () => () => {}
});

interface WebSocketProviderProps {
  children: ReactNode;
  url?: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

export const WebSocketProvider = ({
  children,
  url,
  reconnectInterval: reconnectIntervalProp,
  maxReconnectAttempts: maxReconnectAttemptsProp
}: WebSocketProviderProps) => {
  const advancedSettings = getAdvancedSettings();
  const reconnectInterval = reconnectIntervalProp ?? advancedSettings.websocketReconnectInterval;
  const maxReconnectAttempts =
    maxReconnectAttemptsProp ?? advancedSettings.websocketReconnectAttempts;
  const [isReady, setIsReady] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const isConnectingRef = useRef(false);
  const messageQueueRef = useRef<string[]>([]);
  const listenersRef = useRef(new Set<(data: string) => void>());
  const { isAuthenticated, isInitialized } = useAppSelector((state) => state.auth);

  const connectWebSocket = async () => {
    if (isConnectingRef.current) {
      console.log('Connection already in progress, skipping');
      return;
    }

    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current) {
      if (wsRef.current.readyState === WebSocket.OPEN) {
        console.log('WebSocket connection already active, skipping connection attempt');
        return;
      }

      if (wsRef.current.readyState !== WebSocket.CONNECTING) {
        try {
          wsRef.current.close();
        } catch (e) {
          console.error('Error closing existing WebSocket:', e);
        }
        wsRef.current = null;
      } else {
        console.log('WebSocket is currently connecting, waiting...');
        return;
      }
    }

    isConnectingRef.current = true;
    console.log('Initiating WebSocket connection...');

    try {
      const token = await getAccessToken();
      const wsUrl = url || (await getWebsocketUrl()) + '?token=' + token;
      const socket = new WebSocket(wsUrl);

      socket.onopen = () => {
        console.log('WebSocket connection established');
        setIsReady(true);
        reconnectAttemptsRef.current = 0;
        isConnectingRef.current = false;

        while (messageQueueRef.current.length > 0) {
          const queuedMessage = messageQueueRef.current.shift();
          if (queuedMessage && socket.readyState === WebSocket.OPEN) {
            socket.send(queuedMessage);
          }
        }
      };

      socket.onclose = (event) => {
        console.log(`WebSocket connection closed: ${event.code} ${event.reason}`);
        setIsReady(false);
        isConnectingRef.current = false;
        messageQueueRef.current = [];

        if (!event.wasClean) {
          handleReconnect();
        }
      };

      socket.onmessage = (event) => {
        // Backwards compatible: some parts of the app consume only the latest message.
        setMessage(event.data);

        // Critical: terminals require *every* WS frame; React state can drop/coalesce updates under load.
        for (const listener of listenersRef.current) {
          try {
            listener(event.data);
          } catch (e) {
            console.error('WebSocket listener error:', e);
          }
        }
      };

      socket.onerror = () => {
        isConnectingRef.current = false;
        setIsReady(false);
      };

      wsRef.current = socket;
    } catch (error) {
      console.error('Error creating WebSocket:', error);
      isConnectingRef.current = false;
      handleReconnect();
    }
  };

  const handleReconnect = () => {
    if (reconnectAttemptsRef.current < maxReconnectAttempts) {
      console.log(
        `Attempting to reconnect (${reconnectAttemptsRef.current + 1}/${maxReconnectAttempts})...`
      );
      reconnectAttemptsRef.current += 1;

      const backoffTime = reconnectInterval * Math.pow(1.5, reconnectAttemptsRef.current - 1);

      reconnectTimeoutRef.current = setTimeout(() => {
        connectWebSocket();
      }, backoffTime);
    } else {
      console.error(`Failed to reconnect after ${maxReconnectAttempts} attempts`);
    }
  };

  useEffect(() => {
    if (!isInitialized || !isAuthenticated) {
      return;
    }

    reconnectAttemptsRef.current = 0;
    isConnectingRef.current = false;

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    connectWebSocket();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }

      if (wsRef.current) {
        wsRef.current.onopen = null;
        wsRef.current.onclose = null;
        wsRef.current.onmessage = null;
        wsRef.current.onerror = null;

        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [isAuthenticated, isInitialized]);

  const sendMessage = useCallback((data: string) => {
    const ws = wsRef.current;
    if (!ws) return;

    if (ws.readyState === WebSocket.OPEN) {
      ws.send(data);
    } else if (ws.readyState === WebSocket.CONNECTING) {
      messageQueueRef.current.push(data);
    }
  }, []);

  const sendJsonMessage = useCallback((data: any) => {
    const ws = wsRef.current;
    if (!ws) return;

    const jsonData = JSON.stringify(data);

    if (ws.readyState === WebSocket.OPEN) {
      ws.send(jsonData);
    } else if (ws.readyState === WebSocket.CONNECTING) {
      messageQueueRef.current.push(jsonData);
    }
  }, []);

  const subscribe = useCallback((listener: (data: string) => void) => {
    listenersRef.current.add(listener);
    return () => {
      listenersRef.current.delete(listener);
    };
  }, []);

  const contextValue: WebSocketContextValue = {
    isReady,
    message,
    sendMessage,
    sendJsonMessage,
    subscribe
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
