'use client';
import { Application } from '@/redux/types/applications';
import React from 'react';
import { DeploymentStatusChart } from './deploymentStatusChart';

function Monitor({ application }: { application?: Application }) {
  if (!application) {
    return null;
  }

  const deployments = application.deployments || [];

  return (
    <div className="space-y-6">
      <div className="">
        <DeploymentStatusChart deployments={deployments} />
      </div>
    </div>
  );
}

export default Monitor;
