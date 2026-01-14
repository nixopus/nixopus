'use client';

import {
  useGetApplicationByIdQuery,
  useGetApplicationDeploymentsQuery
} from '@/redux/services/deploy/applicationsApi';
import { useParams, useSearchParams } from 'next/navigation';
import { useEffect, useState, useRef, useMemo } from 'react';
import { useApplicationWebSocket } from './use_application_websocket';
import {
  Application,
  ApplicationDeployment,
  ApplicationDeploymentStatus
} from '@/redux/types/applications';
import { BuildPack, Environment } from '@/redux/types/deploy-form';
import type { TabItem } from '@/components/ui/tabs-wrapper';
import { Activity, Settings, Layers, ScrollText } from 'lucide-react';
import DeploymentsList, {
  ApplicationLogs,
  Monitor
} from '@/packages/components/application-details';
import { DeployConfigureForm } from '@/packages/components/application-form';
import { useTranslation } from '../shared/use-translation';

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
  const { t } = useTranslation();
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
  const [activeTab, setActiveTab] = useState(defaultTab);
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

  const totalDeployments = deploymentsData?.total_count || 0;
  const totalPages = Math.ceil(totalDeployments / deploymentsPerPage);

  const tabs: TabItem[] = useMemo(
    () => [
      {
        value: 'monitoring',
        label: t('selfHost.application.tabs.monitoring'),
        icon: Activity,
        content: <Monitor application={application} />
      },
      {
        value: 'configuration',
        label: t('selfHost.application.tabs.configuration'),
        icon: Settings,
        content: (
          <DeployConfigureForm
            application_name={application?.name}
            domains={application?.domains?.map((d) => d.domain)}
            environment={application?.environment as Environment | undefined}
            env_variables={envVariables}
            build_variables={buildVariables}
            build_pack={application?.build_pack as BuildPack}
            branch={application?.branch}
            port={application?.port?.toString()}
            repository={application?.repository}
            pre_run_commands={application?.pre_run_command}
            post_run_commands={application?.post_run_command}
            application_id={application?.id}
            dockerFilePath={application?.dockerfile_path}
            base_path={application?.base_path}
          />
        )
      },
      {
        value: 'deployments',
        label: t('selfHost.application.tabs.deployments'),
        icon: Layers,
        content: (
          <DeploymentsList
            deployments={application?.deployments}
            currentPage={deploymentsPage}
            totalPages={totalPages}
            onPageChange={setDeploymentsPage}
          />
        )
      },
      {
        value: 'logs',
        label: t('selfHost.application.tabs.logs'),
        icon: ScrollText,
        content: (
          <ApplicationLogs
            id={application?.id || ''}
            currentPage={currentPage}
            setCurrentPage={setCurrentPage}
          />
        )
      }
    ],
    [
      t,
      application,
      envVariables,
      buildVariables,
      deploymentsPage,
      totalPages,
      currentPage,
      setCurrentPage,
      setDeploymentsPage
    ]
  );

  const sharedTabTriggerClassName =
    'rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2';

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
            deployments: (applicationRef.current.deployments || []).map((d) =>
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
            deployments: (applicationRef.current.deployments || []).map((d) => {
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
    } catch (error) {}
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
    totalDeployments,
    totalPages,
    activeTab,
    setActiveTab,
    tabs,
    sharedTabTriggerClassName
  };
}

export default useApplicationDetails;
