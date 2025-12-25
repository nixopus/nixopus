import { useState, useCallback, useMemo, useRef, useEffect } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { ComposeSourceType } from '@/redux/types/deploy-form';
import { useGetGithubRepositoryBranchesMutation } from '@/redux/services/connector/githubConnectorApi';
import { toast } from 'sonner';

interface UseDeployFormStepperProps {
  form: UseFormReturn<any>;
  repository_full_name?: string;
  needsRepository: boolean;
  isDockerCompose: boolean;
  watchedComposeSource: ComposeSourceType | undefined;
}

interface StepperStep {
  id: string;
  title: string;
}

export const useDeployFormStepper = ({
  form,
  repository_full_name,
  needsRepository,
  isDockerCompose,
  watchedComposeSource
}: UseDeployFormStepperProps) => {
  const [currentStepId, setCurrentStepId] = useState('application');
  const stepperMethodsRef = useRef<any>(null);
  const [getGithubRepositoryBranches, { isLoading: isLoadingBranches }] =
    useGetGithubRepositoryBranchesMutation();
  const [availableBranches, setAvailableBranches] = useState<{ label: string; value: string }[]>(
    []
  );

  const stepperSteps = useMemo<StepperStep[]>(() => {
    return [
      { id: 'application', title: 'Application' },
      { id: 'source', title: 'Source' },
      { id: 'variables', title: 'Variables' }
    ];
  }, []);

  const fetchRepositoryBranches = useCallback(async () => {
    if (!repository_full_name) return;

    try {
      const result = await getGithubRepositoryBranches(repository_full_name).unwrap();
      const branchOptions = result.map((branch) => ({
        label: branch.name,
        value: branch.name
      }));
      setAvailableBranches(branchOptions);

      const current = form.getValues('branch');
      const defaultBranch =
        branchOptions.find((b) => b.value === 'main') ||
        branchOptions.find((b) => b.value === 'master') ||
        branchOptions[0];
      if (!current || !branchOptions.some((b) => b.value === current)) {
        if (defaultBranch) {
          form.setValue('branch', defaultBranch.value);
        } else {
          form.setValue('branch', '');
        }
      }
    } catch (error) {
      toast.error('Failed to fetch repository branches');
    }
  }, [getGithubRepositoryBranches, form, repository_full_name]);

  useEffect(() => {
    if (repository_full_name && needsRepository) {
      fetchRepositoryBranches();
    } else if (!needsRepository) {
      setAvailableBranches([]);
    }
  }, [repository_full_name, fetchRepositoryBranches, needsRepository]);

  useEffect(() => {
    if (stepperMethodsRef.current && stepperMethodsRef.current.current.id !== currentStepId) {
      stepperMethodsRef.current.goTo(currentStepId);
    }
  }, [currentStepId]);

  const validateCurrentStep = useCallback(async (): Promise<boolean> => {
    const currentStep = currentStepId;
    let fieldsToValidate: string[] = [];

    switch (currentStep) {
      case 'application':
        fieldsToValidate = ['application_name', 'environment', 'build_pack', 'domain'];
        break;
      case 'source':
        if (isDockerCompose) {
          fieldsToValidate = ['compose_source'];
          if (watchedComposeSource === ComposeSourceType.Repository) {
            fieldsToValidate.push('branch');
          } else if (watchedComposeSource === ComposeSourceType.URL) {
            fieldsToValidate.push('compose_file_url');
          } else if (watchedComposeSource === ComposeSourceType.Raw) {
            fieldsToValidate.push('compose_file_content');
          }
        } else {
          // Dockerfile
          fieldsToValidate = ['branch', 'port'];
        }
        break;
      case 'variables':
        // Variables are optional, no strict validation
        break;
      default:
        return true;
    }

    if (fieldsToValidate.length === 0) return true;
    const isValid = await form.trigger(fieldsToValidate as any);
    return isValid;
  }, [currentStepId, form, isDockerCompose, watchedComposeSource]);

  const handleNext = useCallback(async () => {
    const isValid = await validateCurrentStep();
    if (!isValid) {
      toast.warning('Please fix the errors before proceeding');
      return;
    }
    const currentIndex = stepperSteps.findIndex((step) => step.id === currentStepId);
    if (currentIndex < stepperSteps.length - 1) {
      setCurrentStepId(stepperSteps[currentIndex + 1].id);
    }
  }, [currentStepId, stepperSteps, validateCurrentStep]);

  const handlePrev = useCallback(() => {
    const currentIndex = stepperSteps.findIndex((step) => step.id === currentStepId);
    if (currentIndex > 0) {
      setCurrentStepId(stepperSteps[currentIndex - 1].id);
    }
  }, [currentStepId, stepperSteps]);

  const handleStepClick = useCallback((stepId: string) => {
    setCurrentStepId(stepId);
  }, []);

  const setStepperMethods = useCallback((methods: any) => {
    stepperMethodsRef.current = methods;
  }, []);

  const isFirstStep = stepperSteps[0]?.id === currentStepId;
  const isLastStep = stepperSteps[stepperSteps.length - 1]?.id === currentStepId;

  return {
    currentStepId,
    stepperSteps,
    availableBranches,
    isLoadingBranches,
    isFirstStep,
    isLastStep,
    handleNext,
    handlePrev,
    handleStepClick,
    setStepperMethods,
    validateCurrentStep
  };
};
