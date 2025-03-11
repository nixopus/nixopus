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
  url = WEBSOCKET_URL + '?token=' + localStorage.getItem('token'),
  reconnectInterval = 3000,
  maxReconnectAttempts = 5
}: WebSocketProviderProps) => {
  const [isReady, setIsReady] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const isConnectingRef = useRef(false);

  const connectWebSocket = () => {
    if (isConnectingRef.current) {
      console.log('Connection already in progress, skipping');
      return;
    }

    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      console.log('WebSocket connection already active, skipping connection attempt');
      return;
    }

    if (wsRef.current && wsRef.current.readyState !== WebSocket.OPEN) {
      try {
        wsRef.current.close();
      } catch (e) {
        console.warn('Error closing existing WebSocket:', e);
      }
      wsRef.current = null;
    }

    isConnectingRef.current = true;
    console.log('Initiating WebSocket connection...');

    try {
      const socket = new WebSocket(url);

      socket.onopen = () => {
        console.log('WebSocket connection established');
        setIsReady(true);
        reconnectAttemptsRef.current = 0;
        isConnectingRef.current = false;
      };

      socket.onclose = (event) => {
        console.log(`WebSocket connection closed: ${event.code} ${event.reason}`);
        setIsReady(false);
        isConnectingRef.current = false;

        if (!event.wasClean) {
          handleReconnect();
        }
      };

      socket.onmessage = (event) => {
        console.log('WebSocket message received:', event.data);
        setMessage(event.data);
      };

      socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        isConnectingRef.current = false;
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

      reconnectTimeoutRef.current = setTimeout(() => {
        connectWebSocket();
      }, reconnectInterval);
    } else {
      console.error(`Failed to reconnect after ${maxReconnectAttempts} attempts`);
    }
  };

  useEffect(() => {
    connectWebSocket();

    return () => {
      console.log('Cleaning up WebSocket');
      isConnectingRef.current = false;

      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }

      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
      }
    };
  }, [url]);

  const sendMessage = (data: string) => {
    if (wsRef.current && isReady) {
      wsRef.current.send(data);
    } else {
      console.warn('Cannot send message, WebSocket is not connected');
    }
  };

  const sendJsonMessage = (data: any) => {
    if (wsRef.current && isReady) {
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
