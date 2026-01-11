import { SOCKET_ACTIONS, SOCKET_EVENTS } from '../api-conf';

export const SubscribeToTopic = (id: string, type: string) => {
  switch (type) {
    case SOCKET_EVENTS.MONITOR_APPLICATION_DEPLOYMENT:
      return {
        action: SOCKET_ACTIONS.SUBSCRIBE,
        topic: SOCKET_EVENTS.MONITOR_APPLICATION_DEPLOYMENT,
        data: {
          resource_id: id
        }
      };
    case SOCKET_EVENTS.MONITOR_HEALTH_CHECK:
      return {
        action: SOCKET_ACTIONS.SUBSCRIBE,
        topic: SOCKET_EVENTS.MONITOR_HEALTH_CHECK,
        data: {
          resource_id: id
        }
      };
    default:
      break;
  }
};
