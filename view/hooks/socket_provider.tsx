import { WEBSOCKET_URL } from "@/redux/conf";
import { createContext, useContext, useEffect, useRef, useState, ReactNode } from "react";

type WebSocketContextValue = {
  isReady: boolean;
  message: string | null;
  sendMessage: (data: string) => void;
  sendJsonMessage: (data: any) => void;
};

const WebSocketContext = createContext<WebSocketContextValue>({
  isReady: false,
  message: null,
  sendMessage: () => { },
  sendJsonMessage: () => { },
});

interface WebSocketProviderProps {
  children: ReactNode;
  url?: string;
}

export const WebSocketProvider = ({
  children,
  url = WEBSOCKET_URL + "?token=" + localStorage.getItem("token"),
}: WebSocketProviderProps) => {
  const [isReady, setIsReady] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const socket = new WebSocket(url);

    socket.onopen = () => {
      console.log("WebSocket connection established");
      setIsReady(true);
    };


    socket.onclose = () => {
      console.log("WebSocket connection closed");
      setIsReady(false);
    };

    socket.onmessage = (event) => {
      console.log("WebSocket message received:", event.data);
      setMessage(event.data);
    };

    socket.onerror = (error) => {
      console.log("WebSocket error:", error);
    };

    wsRef.current = socket;


    return () => {
      console.log("Closing WebSocket connection");
      socket.close();
    };
  }, [url]);


  const sendMessage = (data: string) => {
    if (wsRef.current && isReady) {
      wsRef.current.send(data);
    } else {
      console.warn("Cannot send message, WebSocket is not connected");
    }
  };

  const sendJsonMessage = (data: any) => {
    if (wsRef.current && isReady) {
      wsRef.current.send(JSON.stringify(data));
    } else {
      console.warn("Cannot send message, WebSocket is not connected");
    }
  };

  const contextValue: WebSocketContextValue = {
    isReady,
    message,
    sendMessage,
    sendJsonMessage
  };

  return (
    <WebSocketContext.Provider value={contextValue}>
      {children}
    </WebSocketContext.Provider>
  );
};


export const useWebSocket = () => {
  const context = useContext(WebSocketContext);

  if (context === undefined) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }

  return context;
};