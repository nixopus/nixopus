import { useWebSocket } from '@/packages/hooks/shared/socket-provider';
import { SOCKET_EVENTS } from '@/redux/api-conf';
import { SubscribeToTopic } from '@/redux/sockets/socket';
import { useEffect } from 'react';

export function useApplicationWebSocket(id: string) {
  const { isReady, message, sendJsonMessage } = useWebSocket();

  useEffect(() => {
    if (id && isReady) {
      sendJsonMessage(SubscribeToTopic(id, SOCKET_EVENTS.MONITOR_APPLICATION_DEPLOYMENT));
    }
  }, [id, isReady, sendJsonMessage]);

  return { message };
}
