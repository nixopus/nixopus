import React from 'react';
import { WelcomeStep } from './steps/welcome-step';
import { InstallAppStep } from './steps/install-app-step';
import { GitHubAppCredentials } from '@/redux/types/github';

interface StepPanelProps {
  Stepper: any;
  currentStepId: string;
  credentials: GitHubAppCredentials | null;
  error: string | null;
  organization?: string;
  onRegistrationSuccess: (creds: GitHubAppCredentials) => void;
  onRegistrationError: (error: Error) => void;
  onInstallationSuccess: () => void;
  onInstallationError: (error: Error) => void;
  onCreateClick?: (createFn: () => void) => void;
}

const renderStepContent = ({
  currentStepId,
  credentials,
  error,
  organization,
  onRegistrationSuccess,
  onRegistrationError,
  onInstallationSuccess,
  onInstallationError,
  onCreateClick
}: Omit<StepPanelProps, 'Stepper'>): React.ReactNode => {
  if (currentStepId === 'create-app') {
    return (
      <WelcomeStep
        error={error}
        organization={organization}
        onSuccess={onRegistrationSuccess}
        onError={onRegistrationError}
        onCreateClick={onCreateClick}
      />
    );
  }

  if (currentStepId === 'install-app' && credentials) {
    return (
      <InstallAppStep
        credentials={credentials}
        organization={organization}
        onSuccess={onInstallationSuccess}
        onError={onInstallationError}
      />
    );
  }

  return null;
};

export const StepPanel: React.FC<StepPanelProps> = (props) => {
  const { Stepper, ...rest } = props;

  return (
    <Stepper.Panel>
      {renderStepContent(rest)}
    </Stepper.Panel>
  );
};

