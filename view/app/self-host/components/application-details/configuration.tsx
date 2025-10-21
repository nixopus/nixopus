'use client';
import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import FormInputField from '@/components/ui/form-input-field';
import { FormSelectTagInputField } from '@/components/ui/form-select-tag-field';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { BuildPack, Environment } from '@/redux/types/deploy-form';
import useUpdateDeployment from '../../hooks/use_update_deployment';
import { parsePort } from '../../utils/parsePort';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard, AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import {
  ChevronDownIcon,
  ChevronRightIcon,
  SettingsIcon,
  ServerIcon,
  CodeIcon,
  InfoIcon,
  Terminal
} from 'lucide-react';

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

interface CollapsibleSectionProps {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
  icon?: React.ComponentType<{ size?: number; className?: string }>;
  badge?: string;
  description?: string;
}

const CollapsibleSection = ({
  title,
  children,
  defaultOpen = false,
  icon: Icon,
  badge,
  description
}: CollapsibleSectionProps) => {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  return (
    <div className="border rounded-lg overflow-hidden">
      <Collapsible open={isOpen} onOpenChange={setIsOpen}>
        <CollapsibleTrigger className="w-full p-4 flex items-center justify-between hover:bg-muted/50 transition-colors group">
          <div className="flex items-center gap-3">
            {Icon && (
              <Icon
                size={20}
                className="text-muted-foreground group-hover:text-foreground transition-colors"
              />
            )}
            <div className="text-left">
              <h3 className="text-sm font-medium">{title}</h3>
              {description && <p className="text-xs text-muted-foreground mt-1">{description}</p>}
            </div>
            {badge && (
              <span
                className={`px-2 py-1 ${badge.toLowerCase() === 'required' ? 'bg-destructive/10 text-destructive' : badge.toLowerCase() === 'read-only' ? 'bg-muted/10 text-muted-foreground' : 'bg-primary/10 text-primary'} text-xs rounded-full font-medium`}
              >
                {badge}
              </span>
            )}
          </div>
          {isOpen ? (
            <ChevronDownIcon className="h-4 w-4 text-muted-foreground transition-transform duration-200" />
          ) : (
            <ChevronRightIcon className="h-4 w-4 text-muted-foreground transition-transform duration-200" />
          )}
        </CollapsibleTrigger>
        <CollapsibleContent className="border-t bg-muted/20">
          <div className="p-4 space-y-4">{children}</div>
        </CollapsibleContent>
      </Collapsible>
    </div>
  );
};

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
    <ResourceGuard resource="deploy" action="read" loadingFallback={<Skeleton className="h-96" />}>
      <AnyPermissionGuard permissions={['deploy:update']} loadingFallback={null}>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <CollapsibleSection
              title={t('selfHost.configuration.sections.basicConfiguration')}
              icon={SettingsIcon}
              badge="Required"
              description="Essential settings for your application"
              defaultOpen={true}
            >
              <div className="grid sm:grid-cols-2 gap-4">
                <FormInputField
                  form={form}
                  label={t('selfHost.configuration.fields.applicationName.label')}
                  name="name"
                  placeholder={t('selfHost.configuration.fields.applicationName.label')}
                />
                {build_pack !== BuildPack.Static && (
                  <FormInputField
                    form={form}
                    label={t('selfHost.configuration.fields.port.label')}
                    name="port"
                    placeholder="3000"
                    validator={(value) => parsePort(value) !== null}
                  />
                )}
              </div>
            </CollapsibleSection>

            {build_pack !== BuildPack.Static && (
              <>
                <CollapsibleSection
                  title={t('selfHost.configuration.sections.dockerConfiguration')}
                  icon={ServerIcon}
                  badge="Optional"
                  description="Container and deployment configuration"
                  defaultOpen={false}
                >
                  <div className="grid sm:grid-cols-2 gap-4">
                    <FormInputField
                      form={form}
                      label={t('selfHost.configuration.fields.basePath.label')}
                      name="base_path"
                      placeholder="/"
                      required={false}
                    />
                    <FormInputField
                      form={form}
                      label={t('selfHost.configuration.fields.dockerfilePath.label')}
                      name="DockerfilePath"
                      placeholder="Dockerfile"
                      required={false}
                    />
                  </div>
                </CollapsibleSection>

                <CollapsibleSection
                  title={t('selfHost.configuration.sections.environmentVariables')}
                  icon={CodeIcon}
                  badge="Optional"
                  description="Runtime and build-time variables"
                  defaultOpen={false}
                >
                  <div className="grid sm:grid-cols-2 gap-4">
                    <FormSelectTagInputField
                      form={form}
                      label={t('selfHost.configuration.fields.environmentVariables.label')}
                      name="environment_variables"
                      placeholder={t(
                        'selfHost.configuration.fields.environmentVariables.placeholder'
                      )}
                      required={false}
                      validator={validateEnvVar}
                      defaultValues={env_variables}
                    />
                    <FormSelectTagInputField
                      form={form}
                      label={t('selfHost.configuration.fields.buildVariables.label')}
                      name="build_variables"
                      placeholder={t('selfHost.configuration.fields.buildVariables.placeholder')}
                      required={false}
                      validator={validateEnvVar}
                      defaultValues={build_variables}
                    />
                  </div>
                </CollapsibleSection>

                <CollapsibleSection
                  title={t('selfHost.configuration.sections.commands')}
                  icon={Terminal}
                  badge="Optional"
                  description="Pre and post deployment scripts"
                  defaultOpen={false}
                >
                  <div className="grid sm:grid-cols-2 gap-4">
                    <FormInputField
                      form={form}
                      label={t('selfHost.configuration.fields.preRunCommands.label')}
                      name="pre_run_command"
                      placeholder={t('selfHost.configuration.fields.preRunCommands.placeholder')}
                      required={false}
                    />
                    <FormInputField
                      form={form}
                      label={t('selfHost.configuration.fields.postRunCommands.label')}
                      name="post_run_command"
                      placeholder={t('selfHost.configuration.fields.postRunCommands.placeholder')}
                      required={false}
                    />
                  </div>
                </CollapsibleSection>
              </>
            )}

            <CollapsibleSection
              title={t('selfHost.configuration.sections.deploymentInformation')}
              icon={InfoIcon}
              badge="Read-only"
              description="Current deployment settings and metadata"
              defaultOpen={false}
            >
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
            </CollapsibleSection>

            <div className="pt-4 flex justify-end">
              <Button type="submit" className="w-fit cursor-pointer" disabled={isLoading}>
                {isLoading
                  ? t('selfHost.configuration.buttons.updating')
                  : t('selfHost.configuration.buttons.update')}
              </Button>
            </div>
          </form>
        </Form>
      </AnyPermissionGuard>
    </ResourceGuard>
  );
};
