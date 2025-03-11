import { useWebSocket } from '@/hooks/socket_provider';
import { SOCKET_EVENTS } from '@/redux/api-conf';
import { useGetApplicationByIdQuery } from '@/redux/services/deploy/applicationsApi';
import { SubscribeToTopic } from '@/redux/sockets/socket';
import { useParams } from 'next/navigation';
import { useEffect, useState } from 'react';

function useApplicationDetails() {
  const { sendJsonMessage } = useWebSocket();
  const { id } = useParams();
  const { data: application } = useGetApplicationByIdQuery({ id: id as string }, { skip: !id });
  const [currentPage, setCurrentPage] = useState(1);

  useEffect(() => {
    sendJsonMessage(SubscribeToTopic(id as string, SOCKET_EVENTS.MONITOR_APPLICATION_DEPLOYMENT));
  }, []);

  return {
    currentPage,
    setCurrentPage,
    application
  };
}

export default useApplicationDetails;
