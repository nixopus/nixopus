import { useGetApplicationByIdQuery, useGetApplicationDeploymentsQuery } from '@/redux/services/deploy/applicationsApi';
import { useParams, useSearchParams } from 'next/navigation';
import { useEffect, useState, useRef } from 'react';
import { useApplicationWebSocket } from './use_application_websocket';
import { Application, ApplicationDeployment, ApplicationDeploymentStatus } from '@/redux/types/applications';

interface WebSocketMessage {
  action: string;
  data: {
    action: string;
    application_id: string;
    data: ApplicationDeployment | ApplicationDeploymentStatus;
    table: string;
  };
  topic: string;
}

function useApplicationDetails() {
  const { id } = useParams();
  const applicationId = id as string;
  const [deploymentsPage, setDeploymentsPage] = useState(1);
  const [deploymentsPerPage] = useState(9);
  
  const { data: applicationData } = useGetApplicationByIdQuery(
    { id: applicationId },
    { skip: !applicationId }
  );

  const { data: deploymentsData } = useGetApplicationDeploymentsQuery(
    { 
      id: applicationId,
      page: deploymentsPage,
      limit: deploymentsPerPage
    },
    { skip: !applicationId }
  );

  const [application, setApplication] = useState<Application | undefined>(applicationData);
  const applicationRef = useRef<Application | undefined>(applicationData);
  const [currentPage, setCurrentPage] = useState(1);
  const searchParams = useSearchParams();
  const defaultTab = searchParams.get('logs') === 'true' ? 'logs' : 'monitoring';
  const { message } = useApplicationWebSocket(applicationId);

  useEffect(() => {
    if (applicationData) {
      const initialApplication = {
        ...applicationData,
        deployments: deploymentsData?.deployments || []
      };
      setApplication(initialApplication);
      applicationRef.current = initialApplication;
    }
  }, [applicationData, deploymentsData]);

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
      return {};
    }
  };

  const envVariables = application?.environment_variables
    ? parseEnvVariables(application.environment_variables)
    : {};

  const buildVariables = application?.build_variables
    ? parseEnvVariables(application.build_variables)
    : {};

  useEffect(() => {
    if (!message || !applicationRef.current) return;

    try {
      const parsedMessage = JSON.parse(message) as WebSocketMessage;
      
      if (parsedMessage.action !== 'message' || !parsedMessage.data) return;

      const { action, table, data } = parsedMessage.data;
      
      if (!table || !action || !data) return;
      
      if (table === 'application_deployment') {
        const deployment = data as ApplicationDeployment;
        if (action === 'INSERT') {
          const updatedApplication = {
            ...applicationRef.current,
            deployments: [deployment, ...(applicationRef.current.deployments || [])]
          };
          setApplication(updatedApplication);
          applicationRef.current = updatedApplication;
        } else if (action === 'UPDATE') {
          const updatedApplication = {
            ...applicationRef.current,
            deployments: (applicationRef.current.deployments || []).map(d => 
              d.id === deployment.id ? deployment : d
            )
          };
          setApplication(updatedApplication);
          applicationRef.current = updatedApplication;
        }
      } else if (table === 'application_deployment_status') {
        const status = data as ApplicationDeploymentStatus;
        if (action === 'INSERT' || action === 'UPDATE') {
          const updatedApplication = {
            ...applicationRef.current,
            deployments: (applicationRef.current.deployments || []).map(d => {
              if (d.id === status.application_deployment_id) {
                return { ...d, status };
              }
              return d;
            })
          };
          setApplication(updatedApplication);
          applicationRef.current = updatedApplication;
        }
      }
    } catch (error) {
    }
  }, [message]);

  return {
    application,
    currentPage,
    setCurrentPage,
    defaultTab,
    envVariables,
    buildVariables,
    deploymentsPage,
    setDeploymentsPage,
    deploymentsPerPage,
    totalDeployments: deploymentsData?.total_count || 0
  };
}

export default useApplicationDetails;

