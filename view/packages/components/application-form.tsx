'use client';
import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import FormInputField from '@/components/ui/form-input-field';
import FormSelectField from '@/components/ui/form-select-field';
import { MultipleDomainInput } from '@/packages/components/multi-domains';
import { EnvVariablesEditor } from '@/components/ui/env-variables-editor';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { BuildPack, Environment } from '@/redux/types/deploy-form';
import useUpdateDeployment from '@/packages/hooks/applications/use_update_deployment';
import { useDeploymentConfiguration } from '@/packages/hooks/applications/use_deployment_configuration';
import { parsePort } from '@/packages/utils/util';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { ResourceGuard, AnyPermissionGuard } from '@/packages/components/rbac';
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
import { useQuickDeployForm } from '@/packages/hooks/applications/use_quick_deploy_form';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { Plus } from 'lucide-react';

interface DeployConfigureProps {
  application_name?: string;
  environment?: Environment;
  branch?: string;
  port?: string;
  domains?: string[];
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
  domains: applicationDomains = [],
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

  const { validateEnvVar, form, onSubmit, isLoading } = useUpdateDeployment({
    name: application_name,
    environment: environment,
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

  const {
    environmentOptions,
    dockerConfigFields,
    envVariableEditors,
    commandFields,
    readOnlyFields
  } = useDeploymentConfiguration({
    branch,
    domains: applicationDomains,
    build_pack,
    env_variables,
    build_variables
  });

  const renderReadOnlyField = (
    label: string,
    value: string | undefined,
    description: string,
    isDomains = false
  ) => {
    const [isExpanded, setIsExpanded] = useState(false);
    const displayValue = value || '-';
    const shouldShowMore = displayValue.length > 50;

    // For domains, render as a list
    if (isDomains && value && value !== '-') {
      const domainList = value.split(', ').filter((d) => d.trim() !== '');
      return (
        <div className="space-y-2">
          <label className="text-sm font-medium">{label}</label>
          <div className="px-3 py-2 border rounded-md bg-muted text-muted-foreground">
            {domainList.length > 0 ? (
              <div className="flex flex-wrap gap-2">
                {domainList.map((domain, index) => (
                  <a
                    key={index}
                    href={`https://${domain.trim()}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs font-mono text-primary hover:underline"
                  >
                    {domain.trim()}
                  </a>
                ))}
              </div>
            ) : (
              <span>-</span>
            )}
          </div>
          <p className="text-sm text-muted-foreground">{description}</p>
        </div>
      );
    }

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

  const renderDescriptionWithLink = (descriptionText: string) => (
    <span>
      {descriptionText}{' '}
      <a
        href="https://docs.nixopus.com/self-host/#docker-configuration"
        target="_blank"
        rel="noopener noreferrer"
        className="text-primary hover:underline"
      >
        Learn more
      </a>
    </span>
  );

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
                <FormSelectField
                  form={form}
                  label={t('selfHost.configuration.fields.environment.label')}
                  name="environment"
                  placeholder={t('selfHost.deployForm.fields.environment.placeholder')}
                  selectOptions={environmentOptions}
                  required={false}
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
                    {dockerConfigFields.map((field) => (
                      <FormInputField
                        key={field.name}
                        form={form}
                        label={field.label}
                        name={field.name}
                        placeholder={field.placeholder}
                        required={false}
                        description={renderDescriptionWithLink(field.descriptionText)}
                      />
                    ))}
                  </div>
                </CollapsibleSection>

                <CollapsibleSection
                  title={t('selfHost.configuration.sections.environmentVariables')}
                  icon={CodeIcon}
                  badge="Optional"
                  description="Runtime and build-time variables"
                  defaultOpen={false}
                >
                  <div className="space-y-6">
                    {envVariableEditors.map((editor) => (
                      <EnvVariablesEditor
                        key={editor.name}
                        form={form}
                        label={editor.label}
                        name={editor.name}
                        required={false}
                        validator={validateEnvVar}
                        defaultValues={editor.defaultValues}
                      />
                    ))}
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
                    {commandFields.map((field) => (
                      <FormInputField
                        key={field.name}
                        form={form}
                        label={field.label}
                        name={field.name}
                        placeholder={field.placeholder}
                        required={false}
                      />
                    ))}
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
                {readOnlyFields.slice(0, 2).map((field, index) => (
                  <React.Fragment key={index}>
                    {renderReadOnlyField(
                      field.label,
                      field.value,
                      field.description,
                      field.label.toLowerCase().includes('domain')
                    )}
                  </React.Fragment>
                ))}
              </div>
              <div className="grid sm:grid-cols-2 gap-4">
                {readOnlyFields.slice(2).map((field, index) => (
                  <React.Fragment key={index + 2}>
                    {renderReadOnlyField(
                      field.label,
                      field.value,
                      field.description,
                      field.label.toLowerCase().includes('domain')
                    )}
                  </React.Fragment>
                ))}
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

interface QuickDeployFormProps {
  repository?: string;
  repository_full_name?: string;
  application_name?: string;
}

export const QuickDeployForm = ({
  repository,
  repository_full_name,
  application_name = ''
}: QuickDeployFormProps) => {
  const { form, formFields, headerContent, handleCreate, buttonLabel, isCreatingProject } =
    useQuickDeployForm({
      repository,
      repository_full_name,
      application_name
    });

  return (
    <ResourceGuard
      resource="deploy"
      action="create"
      loadingFallback={<Skeleton className="h-96" />}
    >
      <CardWrapper
        header={headerContent}
        className="w-full max-w-2xl mx-auto border-0 shadow-none bg-transparent"
        contentClassName="pt-6"
      >
        <Form {...form}>
          <form className="space-y-6">
            <div className="grid sm:grid-cols-2 gap-4">
              {formFields.map((field) => {
                if (field.type === 'select' && field.isLoading) {
                  return (
                    <div key={field.key} className="space-y-2">
                      <div className="flex gap-2">
                        <label className="text-sm font-medium">{field.label}</label>
                        {field.required && <span className="text-destructive">*</span>}
                      </div>
                      <Skeleton className="h-10 w-full" />
                    </div>
                  );
                }

                if (field.type === 'select') {
                  return (
                    <FormSelectField
                      key={field.key}
                      form={form}
                      label={field.label}
                      name={field.name}
                      placeholder={field.placeholder}
                      selectOptions={field.selectOptions}
                      required={field.required}
                    />
                  );
                }

                if (field.type === 'multi-domains') {
                  return (
                    <MultipleDomainInput
                      key={field.key}
                      form={form}
                      label={field.label}
                      name={field.name}
                      placeholder={field.placeholder}
                      required={field.required}
                      maxDomains={5}
                    />
                  );
                }

                return (
                  <FormInputField
                    key={field.key}
                    form={form}
                    label={field.label}
                    name={field.name}
                    placeholder={field.placeholder}
                    required={field.required}
                  />
                );
              })}
            </div>

            <Button
              type="button"
              onClick={handleCreate}
              disabled={isCreatingProject}
              className="w-full gap-2"
            >
              <Plus className="h-4 w-4" />
              {buttonLabel}
            </Button>
          </form>
        </Form>
      </CardWrapper>
    </ResourceGuard>
  );
};
