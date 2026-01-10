'use client';

import React from 'react';
import { DeploymentLogsTable } from '../deployment-logs';

interface ApplicationLogsProps {
  id: string;
  currentPage?: number;
  setCurrentPage?: (page: number) => void;
}

const ApplicationLogs = ({ id }: ApplicationLogsProps) => {
  return <DeploymentLogsTable id={id} isDeployment={false} />;
};

export default ApplicationLogs;
