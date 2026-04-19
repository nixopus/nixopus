'use client';

import { useMemo } from 'react';
import { useGetApplicationsQuery } from '@/redux/services/deploy/applicationsApi';
import type { Application } from '@/redux/types/applications';
import type { Machine } from './use-machines';

export function useMachineCard(machine: Machine) {
  const isProvisioning = machine.status === 'provisioning';
  const isFailed = machine.status === 'failed';

  const { data: applicationsData } = useGetApplicationsQuery(
    { page: 1, limit: 1000 },
    { skip: isProvisioning || isFailed }
  );

  const serverApplications = useMemo(() => {
    if (!applicationsData?.applications || !machine.ssh_key_id) return [];
    return applicationsData.applications.filter((app: Application) =>
      app.servers?.some((s) => s.server_id === machine.ssh_key_id)
    );
  }, [applicationsData, machine.ssh_key_id]);

  const customDomains = useMemo(() => {
    const domains = new Set<string>();
    serverApplications.forEach((app: Application) => {
      if (app.domains && app.domains.length > 0) {
        app.domains.forEach((domain) => {
          domains.add(domain.domain);
        });
      }
    });
    return Array.from(domains);
  }, [serverApplications]);

  return {
    serverApplications,
    customDomains,
    isProvisioning,
    isFailed,
    isInteractive: !isProvisioning && !isFailed
  };
}
