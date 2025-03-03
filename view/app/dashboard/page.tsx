'use client';
import DashboardPageHeader from '@/components/dashboard-page-header';
import { useAppSelector } from '@/redux/hooks';
import { useRouter, useSearchParams } from 'next/navigation';
import React, { useEffect } from 'react';
import GitHubAppSetup from './components/github-connector/github-app-setup';
import {
  useGetAllGithubConnectorQuery,
  useUpdateGithubConnectorMutation
} from '@/redux/services/connector/githubConnectorApi';

function page() {
  const user = useAppSelector((state) => state.auth.user);
  const authenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const { data: connectors, refetch: GetGithubConnectors } = useGetAllGithubConnectorQuery();
  const searchParams = useSearchParams();
  const [updateGithubConnector] = useUpdateGithubConnectorMutation();

  const router = useRouter();

  useEffect(() => {
    const installationId = searchParams.get('installation_id');
    if (installationId) {
      const githubConnector = async () => {
        await updateGithubConnector({
          installation_id: installationId,
        });
      }
      githubConnector();
      router.push('/dashboard');
    }
  }, [searchParams, router]);

  if (!user || !authenticated) {
    router.push('/login');
    return null;
  }

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader label="Dashboard" description={`Welcome, ${user.username}`} />
      {
        connectors?.length === 0 ? (
          <GitHubAppSetup GetGithubConnectors={GetGithubConnectors} />
        ) : (
          <div className="flex flex-col h-full justify-center items-center gap-4">
            <h2 className="text-xl font-medium text-center text-foreground">Github already connected</h2>
          </div>
        )
      }
    </div>
  );
}

export default page;
