'use client';
import React, { useState } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { GitHubAppCredentials } from '@/redux/types/github';
import GitHubAppManifestComponent from './github-manifest-flow';
import GithubInstaller from './github-app-installer';

interface GitHubAppSetupProps {
  organization?: string;
  GetGithubConnectors: () => void;
}

const GitHubAppSetup: React.FC<GitHubAppSetupProps> = ({ organization, GetGithubConnectors }) => {
  const [setupStage, setSetupStage] = useState<'registration' | 'installation'>('registration');
  const [credentials, setCredentials] = useState<GitHubAppCredentials | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleRegistrationSuccess = async (creds: GitHubAppCredentials) => {
    setCredentials(creds);
    setSetupStage('installation');
    // await GetGithubConnectors();
  };

  const handleRegistrationError = (error: Error) => {
    setError(`Registration failed: ${error.message}`);
  };

  const handleInstallationError = (error: Error) => {
    setError(`Installation failed: ${error.message}`);
  };

  return (
    <div className="flex flex-col items-center space-y-4 p-4">
      {setupStage === 'registration' ? (
        <GitHubAppManifestComponent
          organization={organization}
          onSuccess={handleRegistrationSuccess}
          onError={handleRegistrationError}
        />
      ) : credentials ? (
        <GithubInstaller appSlug={credentials.slug} organization={organization} callbackUrl={''} />
      ) : null}
    </div>
  );
};

export default GitHubAppSetup;
