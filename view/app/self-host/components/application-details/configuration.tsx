'use client';
import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import FormInputField from '@/components/ui/form-input-field';
import { FormSelectTagInputField } from '@/components/ui/form-select-tag-field';
import { BuildPack, Environment } from '@/redux/types/deploy-form';
import useUpdateDeployment from '../../hooks/use_update_deployment';
import { parsePort } from '../../utils/parsePort';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard, AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';

interface DeployConfigureProps {
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
  application_id?: string;
  dockerFilePath?: string;
  base_path?: string;
}

export const DeployConfigureForm = ({
  application_name = '',
  environment = Environment.Production,
  branch = '',
  port = '3000',
  domain = '',
  repository = '',
  build_pack = BuildPack.Dockerfile,
  env_variables = {},
  build_variables = {},
  pre_run_commands = '',
  post_run_commands = '',
  application_id = '',
  dockerFilePath = '/Dockerfile',
  base_path = '/'
}: DeployConfigureProps) => {
  const { t } = useTranslation();

  const { validateEnvVar, form, onSubmit, isLoading, domains } = useUpdateDeployment({
    name: application_name,
    pre_run_command: pre_run_commands,
    post_run_command: post_run_commands,
    build_variables,
    environment_variables: env_variables,
    port: parsePort(port) || 3000,
    force: true,
    id: application_id,
    DockerfilePath: dockerFilePath,
    base_path
  });

  const renderReadOnlyField = (label: string, value: string | undefined, description: string) => {
    const [isExpanded, setIsExpanded] = useState(false);
    const displayValue = value || '-';
    const shouldShowMore = displayValue.length > 50;

    return (
      <div className="space-y-2">
        <label className="text-sm font-medium">{label}</label>
        <div className="px-3 py-2 border rounded-md bg-muted text-muted-foreground overflow-hidden">
          <div className={`${!isExpanded ? 'truncate' : ''}`}>{displayValue}</div>
          {shouldShowMore && (
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className="text-xs text-primary hover:underline mt-1"
            >
              {isExpanded ? 'Show less' : 'Show more'}
            </button>
          )}
        </div>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
    );
  };

  return (
    <ResourceGuard 
      resource="deploy" 
      action="read"
      loadingFallback={<Skeleton className="h-96" />}
    >
      <AnyPermissionGuard 
        permissions={['deploy:update']}
        loadingFallback={null}
      >
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
            <div className="grid sm:grid-cols-2 gap-4">
              <FormInputField
                form={form}
                label={t('selfHost.configuration.fields.applicationName.label')}
                name="name"
                description={t('selfHost.configuration.fields.applicationName.description')}
                placeholder={t('selfHost.configuration.fields.applicationName.label')}
              />
              {build_pack !== BuildPack.Static && (
                <FormInputField
                  form={form}
                  label={t('selfHost.configuration.fields.port.label')}
                  name="port"
                  description={t('selfHost.configuration.fields.port.description')}
                  placeholder="3000"
                  validator={(value) => parsePort(value) !== null}
                />
              )}
            </div>

            {build_pack !== BuildPack.Static && (
              <>
                <div className="grid sm:grid-cols-2 gap-4">
                  <FormInputField
                    form={form}
                    label={t('selfHost.configuration.fields.basePath.label')}
                    name="base_path"
                    description={t('selfHost.configuration.fields.basePath.description')}
                    placeholder="/"
                    required={false}
                  />
                  <FormInputField
                    form={form}
                    label={t('selfHost.configuration.fields.dockerfilePath.label')}
                    name="DockerfilePath"
                    description={t('selfHost.configuration.fields.dockerfilePath.description')}
                    placeholder="Dockerfile"
                    required={false}
                  />
                </div>

                <div className="grid sm:grid-cols-2 gap-4">
                  <FormSelectTagInputField
                    form={form}
                    label={t('selfHost.configuration.fields.environmentVariables.label')}
                    name="environment_variables"
                    description={t('selfHost.configuration.fields.environmentVariables.description')}
                    placeholder={t('selfHost.configuration.fields.environmentVariables.placeholder')}
                    required={false}
                    validator={validateEnvVar}
                    defaultValues={env_variables}
                  />
                  <FormSelectTagInputField
                    form={form}
                    label={t('selfHost.configuration.fields.buildVariables.label')}
                    name="build_variables"
                    description={t('selfHost.configuration.fields.buildVariables.description')}
                    placeholder={t('selfHost.configuration.fields.buildVariables.placeholder')}
                    required={false}
                    validator={validateEnvVar}
                    defaultValues={build_variables}
                  />
                </div>

                <div className="grid sm:grid-cols-2 gap-4">
                  <FormInputField
                    form={form}
                    label={t('selfHost.configuration.fields.preRunCommands.label')}
                    name="pre_run_command"
                    description={t('selfHost.configuration.fields.preRunCommands.description')}
                    placeholder={t('selfHost.configuration.fields.preRunCommands.placeholder')}
                    required={false}
                  />
                  <FormInputField
                    form={form}
                    label={t('selfHost.configuration.fields.postRunCommands.label')}
                    name="post_run_command"
                    description={t('selfHost.configuration.fields.postRunCommands.description')}
                    placeholder={t('selfHost.configuration.fields.postRunCommands.placeholder')}
                    required={false}
                  />
                </div>
              </>
            )}

            <div className="grid sm:grid-cols-2 gap-4">
              {renderReadOnlyField(
                t('selfHost.configuration.fields.environment.label'),
                environment,
                t('selfHost.configuration.fields.environment.description')
              )}
              {renderReadOnlyField(
                t('selfHost.configuration.fields.branch.label'),
                branch,
                t('selfHost.configuration.fields.branch.description')
              )}
            </div>

            <div className="grid sm:grid-cols-2 gap-4">
              {renderReadOnlyField(
                t('selfHost.configuration.fields.domain.label'),
                domain,
                t('selfHost.configuration.fields.domain.description')
              )}
              {renderReadOnlyField(
                t('selfHost.configuration.fields.buildPack.label'),
                build_pack,
                t('selfHost.configuration.fields.buildPack.description')
              )}
            </div>

            <Button type="submit" className="w-full cursor-pointer" disabled={isLoading}>
              {isLoading
                ? t('selfHost.configuration.buttons.updating')
                : t('selfHost.configuration.buttons.update')}
            </Button>
          </form>
        </Form>
      </AnyPermissionGuard>
    </ResourceGuard>
  );
};
