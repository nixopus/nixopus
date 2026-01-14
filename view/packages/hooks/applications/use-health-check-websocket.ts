'use client';

import { useWebSocket } from '@/packages/hooks/shared/socket-provider';
import { SOCKET_EVENTS } from '@/redux/api-conf';
import { SubscribeToTopic } from '@/redux/sockets/socket';
import { useEffect } from 'react';
import { useGetHealthCheckQuery } from '@/redux/services/deploy/healthcheckApi';
import { healthcheckApi } from '@/redux/services/deploy/healthcheckApi';
import { useAppDispatch } from '@/redux/hooks';

interface UseHealthCheckWebSocketProps {
  applicationId: string;
}

interface HealthCheckWebSocketPayload {
  application_id: string;
  health_check_id: string;
  status: string;
  response_time_ms: number;
  status_code?: number;
  error_message?: string;
  checked_at: string;
  consecutive_fails: number;
}

export function useHealthCheckWebSocket({ applicationId }: UseHealthCheckWebSocketProps) {
  const { isReady, sendJsonMessage, subscribe } = useWebSocket();
  const dispatch = useAppDispatch();
  const { data: healthCheck } = useGetHealthCheckQuery(applicationId, {
    skip: !applicationId
  });

  useEffect(() => {
    if (applicationId && isReady && healthCheck) {
      sendJsonMessage(SubscribeToTopic(applicationId, SOCKET_EVENTS.MONITOR_HEALTH_CHECK));
    }
  }, [applicationId, isReady, healthCheck, sendJsonMessage]);

  useEffect(() => {
    if (!applicationId) return;

    // Use subscribe to ensure we receive ALL messages, not just the latest one
    const unsubscribe = subscribe((rawMessage: string) => {
      try {
        const parsedMessage = typeof rawMessage === 'string' ? JSON.parse(rawMessage) : rawMessage;

        // Handle subscription confirmation
        if (parsedMessage.action === 'subscribed') {
          return;
        }

        // Handle health check result messages
        if (parsedMessage.action === 'message') {
          const expectedTopic = `${SOCKET_EVENTS.MONITOR_HEALTH_CHECK}:${applicationId}`;

          if (parsedMessage.topic === expectedTopic) {
            if (!parsedMessage.data) {
              console.warn(
                '[HealthCheck WS] Received message with null/undefined data:',
                parsedMessage
              );
              return;
            }

            const payload = parsedMessage.data as HealthCheckWebSocketPayload;

            // Update the health check cache directly with new data
            dispatch(
              healthcheckApi.util.updateQueryData('getHealthCheck', applicationId, (draft) => {
                if (draft) {
                  draft.consecutive_fails = payload.consecutive_fails;
                  draft.last_checked_at = payload.checked_at;
                  // Store error message if present
                  if (payload.error_message) {
                    draft.last_error_message = payload.error_message;
                  } else if (payload.status === 'error' || payload.status === 'unhealthy') {
                    // Keep existing error message if new one is not provided but status indicates error
                    // This ensures we don't lose the error message on subsequent checks
                  } else {
                    // Clear error message when health check passes
                    draft.last_error_message = undefined;
                  }
                }
              }) as any
            );

            // Invalidate stats to trigger refetch
            dispatch(
              healthcheckApi.util.invalidateTags([{ type: 'HealthCheckStats', id: applicationId }])
            );
          }
        }
      } catch (error) {
        console.error(
          '[HealthCheck WS] Error parsing WebSocket message:',
          error,
          'Raw message:',
          rawMessage
        );
      }
    });

    return () => {
      unsubscribe();
    };
  }, [applicationId, subscribe, dispatch]);
}
