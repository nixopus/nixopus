'use client';
import React, { useState, useRef } from 'react';
import { useGitHubAppSetup } from '../../hooks/use-github-app-setup';
import { StepperNavigation } from './stepper-navigation';
import { StepperControls } from './stepper-controls';
import { StepPanel } from './step-panel';

interface GitHubAppSetupProps {
  organization?: string;
  GetGithubConnectors: () => void;
}

const GitHubAppSetup: React.FC<GitHubAppSetupProps> = ({ organization, GetGithubConnectors }) => {
  const handleGetGithubConnectors = async () => {
    await GetGithubConnectors();
  };

  const createAppRef = useRef<(() => void) | null>(null);

  const {
    Stepper,
    utils,
    credentials,
    error,
    setStepperMethods,
    handleRegistrationSuccess,
    handleRegistrationError,
    handleInstallationSuccess,
    handleInstallationError,
    handleNext,
    handleBack,
    getCanGoNext
  } = useGitHubAppSetup(handleGetGithubConnectors);

  const handleCreateClick = (createFn: () => void) => {
    createAppRef.current = createFn;
  };

  const handleCreateApp = () => {
    if (createAppRef.current) {
      createAppRef.current();
    }
  };

  return (
    <div className="flex flex-col items-center w-full max-w-4xl mx-auto p-4 space-y-6">
      <Stepper.Provider
        variant="horizontal"
        labelOrientation="vertical"
        initialStep="create-app"
        className="w-full"
      >
        {({ methods }) => {
          setStepperMethods(methods);
          const { current } = methods;
          const currentStepId = current.id;
          const isFirstStep = utils.getIndex(currentStepId) === 0;
          const isLastStep = utils.getLast().id === currentStepId;
          const canGoNext = Boolean(getCanGoNext(currentStepId));

          return (
            <>
              <StepperNavigation Stepper={Stepper} />
              <div className="flex justify-center w-full">
                <div className="w-full max-w-2xl">
                  <StepPanel
                    Stepper={Stepper}
                    currentStepId={currentStepId}
                    credentials={credentials}
                    error={error}
                    organization={organization}
                    onRegistrationSuccess={handleRegistrationSuccess}
                    onRegistrationError={handleRegistrationError}
                    onInstallationSuccess={handleInstallationSuccess}
                    onInstallationError={handleInstallationError}
                    onCreateClick={handleCreateClick}
                  />
                  <StepperControls
                    Stepper={Stepper}
                    isFirstStep={isFirstStep}
                    isLastStep={isLastStep}
                    canGoNext={canGoNext}
                    onBack={handleBack}
                    onNext={handleNext}
                    onCreateApp={handleCreateApp}
                    currentStepId={currentStepId}
                  />
                </div>
              </div>
            </>
          );
        }}
      </Stepper.Provider>
    </div>
  );
};

export default GitHubAppSetup;
