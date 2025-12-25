import React from 'react';
import GithubInstaller from '../github-app-installer';
import { GitHubAppCredentials } from '@/redux/types/github';

interface InstallAppStepProps {
  credentials: GitHubAppCredentials;
  organization?: string;
  onSuccess: () => void;
  onError: (error: Error) => void;
}

export const InstallAppStep: React.FC<InstallAppStepProps> = ({
  credentials,
  organization,
  onSuccess,
  onError
}) => {
  return (
    <div className="pt-8 pb-6 px-6">
      <div className="flex justify-center">
        <GithubInstaller
          appSlug={credentials.slug}
          organization={organization}
          callbackUrl=""
          onSuccess={onSuccess}
          onError={onError}
        />
      </div>
    </div>
  );
};

