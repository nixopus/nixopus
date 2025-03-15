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


  const parseEnvVariables = (variablesString: string | undefined): Record<string, string> => {
    if (!variablesString) return {};

    try {
      return variablesString
        .split(/\s+/)
        .filter((pair) => pair.includes('='))
        .reduce<Record<string, string>>((acc, curr) => {
          const [key, value] = curr.split('=');
          return key && value ? { ...acc, [key]: value } : acc;
        }, {});
    } catch (error) {
      console.error('Error parsing variables:', error);
      return {};
    }
  };

  const envVariables = application?.environment_variables
    ? parseEnvVariables(application.environment_variables)
    : {};

  const buildVariables = application?.build_variables
    ? parseEnvVariables(application.build_variables)
    : {};

  return {
    currentPage,
    setCurrentPage,
    application,
    envVariables,
    buildVariables
  };
}

export default useApplicationDetails;
