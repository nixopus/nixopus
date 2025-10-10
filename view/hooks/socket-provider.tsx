'use client';
import { getWebsocketUrl } from '@/redux/conf';
import { createContext, useContext, useEffect, useRef, useState, ReactNode } from 'react';
import { getAccessToken } from 'supertokens-auth-react/recipe/session';

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
  url,
  reconnectInterval = 3000,
  maxReconnectAttempts = 5
}: WebSocketProviderProps) => {
  const [isReady, setIsReady] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const isConnectingRef = useRef(false);

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
          console.warn('Error closing existing WebSocket:', e);
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
        setMessage(event.data);
      };

      socket.onerror = (error) => {
        console.log('WebSocket error:', error);
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

      const backoffTime = reconnectInterval * Math.pow(1.5, reconnectAttemptsRef.current - 1);

      reconnectTimeoutRef.current = setTimeout(() => {
        connectWebSocket();
      }, backoffTime);
    } else {
      console.error(`Failed to reconnect after ${maxReconnectAttempts} attempts`);
    }
  };

  useEffect(() => {
    reconnectAttemptsRef.current = 0;
    isConnectingRef.current = false;

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    connectWebSocket();
  }, []);

  useEffect(() => {
    if (!wsRef.current) {
      connectWebSocket();
    }

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
