'use client';
import React from 'react';
import GitHubAppSetup from './components/github-connector/github-app-setup';
import { useGetAllGithubConnectorQuery } from '@/redux/services/connector/githubConnectorApi';
import ListRepositories from './components/github-repositories/list-repositories';
import Loading from '@/components/ui/loading';

function page() {
  const {
    data: connectors,
    refetch: GetGithubConnectors,
    isLoading
  } = useGetAllGithubConnectorQuery();

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      {connectors?.length === 0 ? (
        <GitHubAppSetup GetGithubConnectors={GetGithubConnectors} />
      ) : (
        <ListRepositories />
      )}
    </div>
  );
}

export default page;
