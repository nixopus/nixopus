'use client';
import React, { useState, useCallback, useMemo, useRef, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import FormInputField from '@/components/ui/form-input-field';
import FormSelectField from '@/components/ui/form-select-field';
import { FormSelectTagInputField } from '@/components/ui/form-select-tag-field';
import { BuildPack, Environment } from '@/redux/types/deploy-form';
import useCreateDeployment from '../../hooks/use_create_deployment';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import { defineStepper } from '@/components/stepper';
import { useIsMobile } from '@/hooks/use-mobile';
import { toast } from 'sonner';
import { useGetGithubRepositoryBranchesMutation } from '@/redux/services/connector/githubConnectorApi';

interface DeployFormProps {
  application_name?: string;
  environment?: Environment;
  branch?: string;
  port?: string;
  domain?: string;
  repository?: string;
  repository_full_name?: string;
  build_pack?: BuildPack;
  env_variables?: Record<string, string>;
  build_variables?: Record<string, string>;
  pre_run_commands?: string;
  post_run_commands?: string;
  DockerfilePath?: string;
  base_path?: string;
}

export const DeployForm = ({
  application_name = '',
  environment = Environment.Production,
  branch = '',
  port = '3000',
  domain = '',
  repository,
  repository_full_name,
  build_pack = BuildPack.Dockerfile,
  env_variables = {},
  build_variables = {},
  pre_run_commands = '',
  post_run_commands = '',
  DockerfilePath = '/Dockerfile',
  base_path = '/'
}: DeployFormProps) => {
  const { t } = useTranslation();
  const isMobileView = useIsMobile();

  const [getGithubRepositoryBranches, { isLoading: isLoadingBranches }] =
    useGetGithubRepositoryBranchesMutation();
  const [availableBranches, setAvailableBranches] = useState<{ label: string; value: string }[]>(
    []
  );

  const { validateEnvVar, form, onSubmit, parsePort } = useCreateDeployment({
    application_name,
    environment,
    branch,
    port,
    domain,
    repository: repository || '',
    build_pack,
    env_variables,
    build_variables,
    pre_run_commands,
    post_run_commands,
    DockerfilePath,
    base_path
  });

  const [currentStepId, setCurrentStepId] = useState('basic-info');
  const stepperMethodsRef = useRef<any>(null);

  const isStaticBuildPack = form.watch('build_pack') === BuildPack.Static;

  const stepperSteps = useMemo(
    () =>
      isStaticBuildPack
        ? [
            { id: 'basic-info', title: 'Basic Information' },
            { id: 'repository', title: 'Repository & Branch' }
          ]
        : [
            { id: 'basic-info', title: 'Basic Information' },
            { id: 'repository', title: 'Repository & Branch' },
            { id: 'configuration', title: 'Configuration' },
            { id: 'variables', title: 'Variables & Commands' }
          ],
    [isStaticBuildPack]
  );

  const fetchRepositoryBranches = useCallback(async () => {
    if (!repository_full_name) {
      return;
    }

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
    if (isStaticBuildPack && (currentStepId === 'configuration' || currentStepId === 'variables')) {
      setCurrentStepId('repository');
    }
  }, [isStaticBuildPack, currentStepId]);

  useEffect(() => {
    if (repository_full_name) {
      fetchRepositoryBranches();
    } else {
      setAvailableBranches([]);
      form.setValue('branch', '');
    }
  }, [repository_full_name, fetchRepositoryBranches, form]);

  useEffect(() => {
    if (stepperMethodsRef.current && stepperMethodsRef.current.current.id !== currentStepId) {
      stepperMethodsRef.current.goTo(currentStepId);
    }
  }, [currentStepId]);

  const { Stepper } = useMemo(() => defineStepper(...stepperSteps), [stepperSteps]);

  const validateCurrentStep = useCallback(async (): Promise<boolean> => {
    const currentStep = currentStepId;
    let fieldsToValidate: string[] = [];

    switch (currentStep) {
      case 'basic-info':
        fieldsToValidate = ['application_name', 'environment', 'build_pack'];
        if (!isStaticBuildPack) {
          fieldsToValidate.push('port');
        }
        break;
      case 'repository':
        fieldsToValidate = ['branch', 'domain'];
        break;
      case 'configuration':
        fieldsToValidate = ['base_path', 'DockerfilePath'];
        break;
      case 'variables':
        fieldsToValidate = [
          'env_variables',
          'build_variables',
          'pre_run_commands',
          'post_run_commands'
        ];
        break;
      default:
        return true;
    }

    const isValid = await form.trigger(fieldsToValidate as any);
    return isValid;
  }, [currentStepId, form, isStaticBuildPack]);

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

  const isFirstStep = stepperSteps[0]?.id === currentStepId;
  const isLastStep = stepperSteps[stepperSteps.length - 1]?.id === currentStepId;

  const renderStepContent = useCallback(() => {
    switch (currentStepId) {
      case 'basic-info':
        return (
          <div className="space-y-12">
            <div className="grid sm:grid-cols-2 gap-4">
              <FormInputField
                form={form}
                label={t('selfHost.deployForm.fields.applicationName.label')}
                name="application_name"
                placeholder={t('selfHost.deployForm.fields.applicationName.placeholder')}
              />
              <FormSelectField
                form={form}
                label={t('selfHost.deployForm.fields.environment.label')}
                name="environment"
                placeholder={t('selfHost.deployForm.fields.environment.placeholder')}
                selectOptions={[
                  {
                    label: t('selfHost.deployForm.fields.environment.options.staging'),
                    value: 'staging'
                  },
                  {
                    label: t('selfHost.deployForm.fields.environment.options.production'),
                    value: 'production'
                  },
                  {
                    label: t('selfHost.deployForm.fields.environment.options.development'),
                    value: 'development'
                  }
                ]}
              />
            </div>
            <div className="grid sm:grid-cols-2 gap-4">
              <FormSelectField
                form={form}
                label={t('selfHost.deployForm.fields.buildPack.label')}
                name="build_pack"
                placeholder={t('selfHost.deployForm.fields.buildPack.placeholder')}
                selectOptions={[
                  {
                    label: t('selfHost.deployForm.fields.buildPack.options.dockerfile'),
                    value: BuildPack.Dockerfile
                  },
                  {
                    label: t('selfHost.deployForm.fields.buildPack.options.static'),
                    value: BuildPack.Static
                  }
                ]}
              />
              {!isStaticBuildPack && (
                <FormInputField
                  form={form}
                  label={t('selfHost.deployForm.fields.port.label')}
                  name="port"
                  placeholder={t('selfHost.deployForm.fields.port.placeholder')}
                  validator={(value) => parsePort(value) !== null}
                />
              )}
            </div>
          </div>
        );

      case 'repository':
        return (
          <div className="space-y-12">
            <div className="grid sm:grid-cols-2 gap-4">
              <FormInputField
                form={form}
                label={t('selfHost.deployForm.fields.domain.label')}
                name="domain"
                placeholder={t('selfHost.deployForm.fields.domain.placeholder')}
                required={true}
              />
              {isLoadingBranches ? (
                <div className="space-y-2">
                  <div className="flex gap-2">
                    <label className="text-sm font-medium">
                      {t('selfHost.deployForm.fields.branch.label')}
                    </label>
                    <span className="text-destructive">*</span>
                  </div>
                  <Skeleton className="h-10 w-full" />
                </div>
              ) : (
                <FormSelectField
                  form={form}
                  label={t('selfHost.deployForm.fields.branch.label')}
                  name="branch"
                  placeholder={
                    availableBranches.length === 0
                      ? 'No branches available'
                      : t('selfHost.deployForm.fields.branch.placeholder')
                  }
                  selectOptions={availableBranches}
                  required={true}
                />
              )}
            </div>
          </div>
        );

      case 'configuration':
        return (
          <div className="space-y-12">
            <div className="grid sm:grid-cols-2 gap-4">
              <FormInputField
                form={form}
                label={t('selfHost.deployForm.fields.basePath.label')}
                name="base_path"
                placeholder={t('selfHost.deployForm.fields.basePath.placeholder')}
                required={false}
              />
              <FormInputField
                form={form}
                label={t('selfHost.deployForm.fields.dockerfilePath.label')}
                name="DockerfilePath"
                placeholder={t('selfHost.deployForm.fields.dockerfilePath.placeholder')}
                required={false}
              />
            </div>
          </div>
        );

      case 'variables':
        return (
          <div className="space-y-12">
            <div className="grid sm:grid-cols-2 gap-4">
              <FormSelectTagInputField
                form={form}
                label={t('selfHost.deployForm.fields.envVariables.label')}
                name="env_variables"
                placeholder={t('selfHost.deployForm.fields.envVariables.placeholder')}
                required={false}
                validator={validateEnvVar}
              />
              <FormSelectTagInputField
                form={form}
                label={t('selfHost.deployForm.fields.buildVariables.label')}
                name="build_variables"
                placeholder={t('selfHost.deployForm.fields.buildVariables.placeholder')}
                required={false}
                validator={validateEnvVar}
              />
            </div>
            <div className="grid sm:grid-cols-2 gap-4">
              <FormInputField
                form={form}
                label={t('selfHost.deployForm.fields.preRunCommands.label')}
                name="pre_run_commands"
                placeholder={t('selfHost.deployForm.fields.preRunCommands.placeholder')}
                required={false}
              />
              <FormInputField
                form={form}
                label={t('selfHost.deployForm.fields.postRunCommands.label')}
                name="post_run_commands"
                placeholder={t('selfHost.deployForm.fields.postRunCommands.placeholder')}
                required={false}
              />
            </div>
          </div>
        );

      default:
        return null;
    }
  }, [currentStepId, form, t, isStaticBuildPack, parsePort, validateEnvVar]);

  const setStepperMethods = useCallback((methods: any) => {
    stepperMethodsRef.current = methods;
  }, []);

  return (
    <ResourceGuard
      resource="deploy"
      action="create"
      loadingFallback={<Skeleton className="h-96" />}
    >
      <Stepper.Provider
        variant={isMobileView ? 'vertical' : 'horizontal'}
        labelOrientation={isMobileView ? 'horizontal' : 'vertical'}
        className="sm:space-y-4 space-y-2"
      >
        {({ methods }) => {
          setStepperMethods(methods);

          return (
            <>
              <Stepper.Navigation
                aria-label="Deployment Form Steps"
                className="sm:flex-row flex-col mb-24"
              >
                {stepperSteps.map((step, index) => {
                  return (
                    <Stepper.Step
                      key={step.id}
                      of={step.id}
                      onClick={() => handleStepClick(step.id)}
                      className="sm:flex-row flex-col sm:gap-2 gap-1"
                    >
                      <Stepper.Title className="sm:text-base text-sm">{step.title}</Stepper.Title>
                    </Stepper.Step>
                  );
                })}
              </Stepper.Navigation>

              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8 mt-8">
                  <Stepper.Panel key={currentStepId}>{renderStepContent()}</Stepper.Panel>

                  <Stepper.Controls>
                    <div className="flex justify-between w-full mt-16">
                      <Button
                        type="button"
                        variant="outline"
                        onClick={handlePrev}
                        disabled={isFirstStep}
                      >
                        Previous
                      </Button>
                      <div className="flex gap-2">
                        {!isLastStep && (
                          <Button type="button" onClick={handleNext}>
                            Next
                          </Button>
                        )}
                        {isLastStep && (
                          <Button type="submit" className="cursor-pointer">
                            {t('selfHost.deployForm.submit')}
                          </Button>
                        )}
                      </div>
                    </div>
                  </Stepper.Controls>
                </form>
              </Form>
            </>
          );
        }}
      </Stepper.Provider>
    </ResourceGuard>
  );
};
