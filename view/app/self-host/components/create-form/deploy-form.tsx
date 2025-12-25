'use client';
import React, { useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { BuildPack, ComposeSourceType, Environment } from '@/redux/types/deploy-form';
import useCreateDeployment from '../../hooks/use_create_deployment';
import { useDeployFormStepper } from '../../hooks/use_deploy_form_stepper';
import { DeployFormFields } from './deploy-form-fields';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import { defineStepper } from '@/components/stepper';
import { useIsMobile } from '@/hooks/use-mobile';

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
  DockerfilePath = '',
  base_path = '/'
}: DeployFormProps) => {
  const { t } = useTranslation();
  const isMobileView = useIsMobile();

  const { validateEnvVar, form, onSubmit, parsePort, isLoading } = useCreateDeployment({
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
    base_path,
    compose_source: ComposeSourceType.Repository,
    compose_file_url: '',
    compose_file_content: ''
  });

  const watchedBuildPack = form.watch('build_pack');
  const watchedComposeSource = form.watch('compose_source');
  const isDockerCompose = watchedBuildPack === BuildPack.DockerCompose;
  const isComposeFromRepo =
    isDockerCompose && watchedComposeSource === ComposeSourceType.Repository;
  const needsRepository = !isDockerCompose || isComposeFromRepo;

  const {
    currentStepId,
    stepperSteps,
    availableBranches,
    isLoadingBranches,
    isFirstStep,
    isLastStep,
    handleNext,
    handlePrev,
    handleStepClick,
    setStepperMethods
  } = useDeployFormStepper({
    form,
    repository_full_name,
    needsRepository,
    isDockerCompose,
    watchedComposeSource
  });

  const { Stepper } = useMemo(() => defineStepper(...stepperSteps), [stepperSteps]);

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
                className="sm:flex-row flex-col mb-12"
              >
                {stepperSteps.map((step) => {
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
                  <Stepper.Panel key={currentStepId}>
                    <DeployFormFields
                      form={form}
                      currentStepId={currentStepId}
                      isDockerCompose={isDockerCompose}
                      watchedComposeSource={watchedComposeSource}
                      parsePort={parsePort}
                      validateEnvVar={validateEnvVar}
                      isLoadingBranches={isLoadingBranches}
                      availableBranches={availableBranches}
                    />
                  </Stepper.Panel>

                  <Stepper.Controls>
                    <div className="flex justify-between w-full mt-8">
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
                          <Button type="submit" className="cursor-pointer" disabled={isLoading}>
                            {isLoading ? 'Deploying...' : t('selfHost.deployForm.submit')}
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
