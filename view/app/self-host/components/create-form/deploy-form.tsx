'use client';
import React from 'react';
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

interface DeployFormProps {
  application_name?: string;
  environment?: Environment;
  branch?: string;
  port?: string;
  domain?: string;
  repository?: string;
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
  build_pack = BuildPack.Dockerfile,
  env_variables = {},
  build_variables = {},
  pre_run_commands = '',
  post_run_commands = '',
  DockerfilePath = '/Dockerfile',
  base_path = '/'
}: DeployFormProps) => {
  const { t } = useTranslation();

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

  const isStaticBuildPack = form.watch('build_pack') === BuildPack.Static;

  return (
    <ResourceGuard 
      resource="deploy" 
      action="create"
      loadingFallback={<Skeleton className="h-96" />}
    >
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <div className="grid sm:grid-cols-2 gap-4">
            <FormInputField
              form={form}
              label={t('selfHost.deployForm.fields.applicationName.label')}
              name="application_name"
              description={t('selfHost.deployForm.fields.applicationName.description')}
              placeholder={t('selfHost.deployForm.fields.applicationName.placeholder')}
            />
            <FormSelectField
              form={form}
              label={t('selfHost.deployForm.fields.environment.label')}
              name="environment"
              description={t('selfHost.deployForm.fields.environment.description')}
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
              description={t('selfHost.deployForm.fields.buildPack.description')}
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
                description={t('selfHost.deployForm.fields.port.description')}
                placeholder={t('selfHost.deployForm.fields.port.placeholder')}
                validator={(value) => parsePort(value) !== null}
              />
            )}
          </div>
          <div className="grid sm:grid-cols-2 gap-4">
            <FormInputField
              form={form}
              label={t('selfHost.deployForm.fields.domain.label')}
              name="domain"
              description={t('selfHost.deployForm.fields.domain.description')}
              placeholder={t('selfHost.deployForm.fields.domain.placeholder')}
            />
            <FormInputField
              form={form}
              label={t('selfHost.deployForm.fields.branch.label')}
              name="branch"
              description={t('selfHost.deployForm.fields.branch.description')}
              placeholder={t('selfHost.deployForm.fields.branch.placeholder')}
            />
          </div>
          {!isStaticBuildPack && (
            <>
              <div className="grid sm:grid-cols-2 gap-4">
                <FormInputField
                  form={form}
                  label={t('selfHost.deployForm.fields.basePath.label')}
                  name="base_path"
                  description={t('selfHost.deployForm.fields.basePath.description')}
                  placeholder={t('selfHost.deployForm.fields.basePath.placeholder')}
                  required={false}
                />
                <FormInputField
                  form={form}
                  label={t('selfHost.deployForm.fields.dockerfilePath.label')}
                  name="DockerfilePath"
                  description={t('selfHost.deployForm.fields.dockerfilePath.description')}
                  placeholder={t('selfHost.deployForm.fields.dockerfilePath.placeholder')}
                  required={false}
                />
              </div>
              <div className="grid sm:grid-cols-2 gap-4">
                <FormSelectTagInputField
                  form={form}
                  label={t('selfHost.deployForm.fields.envVariables.label')}
                  name="env_variables"
                  description={t('selfHost.deployForm.fields.envVariables.description')}
                  placeholder={t('selfHost.deployForm.fields.envVariables.placeholder')}
                  required={false}
                  validator={validateEnvVar}
                />
                <FormSelectTagInputField
                  form={form}
                  label={t('selfHost.deployForm.fields.buildVariables.label')}
                  name="build_variables"
                  description={t('selfHost.deployForm.fields.buildVariables.description')}
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
                  description={t('selfHost.deployForm.fields.preRunCommands.description')}
                  placeholder={t('selfHost.deployForm.fields.preRunCommands.placeholder')}
                  required={false}
                />
                <FormInputField
                  form={form}
                  label={t('selfHost.deployForm.fields.postRunCommands.label')}
                  name="post_run_commands"
                  description={t('selfHost.deployForm.fields.postRunCommands.description')}
                  placeholder={t('selfHost.deployForm.fields.postRunCommands.placeholder')}
                  required={false}
                />
              </div>
            </>
          )}
          <Button className="w-full cursor-pointer">{t('selfHost.deployForm.submit')}</Button>
        </form>
      </Form>
    </ResourceGuard>
  );
};
