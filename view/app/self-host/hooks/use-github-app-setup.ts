import { useState, useRef, useMemo, useCallback } from 'react';
import { GitHubAppCredentials } from '@/redux/types/github';
import { defineStepper } from '@/components/stepper';

const STEPPER_STEPS = [
  { id: 'create-app', title: 'Create App' },
  { id: 'install-app', title: 'Install App' }
] as const;

export const useGitHubAppSetup = (onInstallationSuccess: () => Promise<void>) => {
  const [credentials, setCredentials] = useState<GitHubAppCredentials | null>(null);
  const [error, setError] = useState<string | null>(null);
  const stepperMethodsRef = useRef<any>(null);

  const stepperDefinition = useMemo(() => defineStepper(...STEPPER_STEPS), []);
  const { Stepper, utils } = stepperDefinition;

  const setStepperMethods = useCallback((methods: any) => {
    stepperMethodsRef.current = methods;
  }, []);

  const handleRegistrationSuccess = useCallback((creds: GitHubAppCredentials) => {
    setCredentials(creds);
    setError(null);
    if (stepperMethodsRef.current) {
      stepperMethodsRef.current.goTo('install-app');
    }
  }, []);

  const handleRegistrationError = useCallback((error: Error) => {
    setError(`Registration failed: ${error.message}`);
  }, []);

  const handleInstallationSuccess = useCallback(async () => {
    await onInstallationSuccess();
  }, [onInstallationSuccess]);

  const handleInstallationError = useCallback((error: Error) => {
    setError(`Installation failed: ${error.message}`);
  }, []);

  const handleNext = useCallback(() => {
    if (!stepperMethodsRef.current) return;
    
    const { current } = stepperMethodsRef.current;
    const nextStep = utils.getNext(current.id);
    
    if (nextStep) {
      stepperMethodsRef.current.goTo(nextStep.id);
      setError(null);
    }
  }, [utils]);

  const handleBack = useCallback(() => {
    if (!stepperMethodsRef.current) return;
    
    const { current } = stepperMethodsRef.current;
    const prevStep = utils.getPrev(current.id);
    
    if (prevStep) {
      stepperMethodsRef.current.goTo(prevStep.id);
      setError(null);
    }
  }, [utils]);

  const getCanGoNext = useCallback((currentStepId: string) => {
    return currentStepId === 'create-app' && credentials;
  }, [credentials]);

  return {
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
  };
};

