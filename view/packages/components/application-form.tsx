'use client';
import React, { useState } from 'react';
import { Button } from '@nixopus/ui';
import { Form } from '@nixopus/ui';
import { FormInputField } from '@nixopus/ui';
import FormSelectField from '@/components/ui/form-select-field';
import { MultipleDomainInput } from '@/packages/components/multi-domains';
import { ComposeDomainInput } from '@/packages/components/compose-domain-input';
import { EnvVariablesEditor } from '@/components/ui/env-variables-editor';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@nixopus/ui';
import { BuildPack } from '@/redux/types/deploy-form';
import { EnvironmentInputField } from '@/packages/components/environment-input-field';
import useUpdateDeployment from '@/packages/hooks/applications/use_update_deployment';
import { useDeploymentConfiguration } from '@/packages/hooks/applications/use_deployment_configuration';
import { parsePort } from '@/packages/utils/util';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { ResourceGuard, AnyPermissionGuard } from '@/packages/components/rbac';
import { Skeleton } from '@nixopus/ui';
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
import { CardWrapper } from '@nixopus/ui';
import { Plus } from 'lucide-react';
import { useGetComposeServicesQuery } from '@/redux/services/deploy/applicationsApi';

interface ComposeDomainEntry {
  domain: string;
  service_name: string;
  port: number;
}

interface DeployConfigureProps {
  application_name?: string;
  environment?: string;
  branch?: string;
  port?: string;
  domains?: string[];
  compose_domains?: ComposeDomainEntry[];
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
    <div className="border border-border rounded-lg overflow-hidden">
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
        <CollapsibleContent className="border-t border-border bg-muted/20">
          <div className="p-4 space-y-4">{children}</div>
        </CollapsibleContent>
      </Collapsible>
    </div>
  );
};

export const DeployConfigureForm = ({
  application_name = '',
  environment = 'production',
  branch = '',
  port = '3000',
  domains: applicationDomains = [],
  compose_domains: initialComposeDomains = [],
  repository = '',
  build_pack = BuildPack.Dockerfile,
  env_variables = {},
  build_variables = {},
  pre_run_commands = '',
  post_run_commands = '',
  application_id = '',
  dockerFilePath,
  base_path = '/'
}: DeployConfigureProps) => {
  const resolvedDockerFilePath =
    dockerFilePath ||
    (build_pack === BuildPack.DockerCompose ? '/docker-compose.yml' : '/Dockerfile');
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
    DockerfilePath: resolvedDockerFilePath,
    base_path,
    domains: applicationDomains,
    compose_domains: initialComposeDomains
  });

  const { dockerConfigFields, envVariableEditors, commandFields, readOnlyFields } =
    useDeploymentConfiguration({
      branch,
      domains: applicationDomains,
      build_pack,
      env_variables,
      build_variables,
      domainsEditable: true
    });

  const isDockerCompose = build_pack === BuildPack.DockerCompose;

  const { data: composeServices = [] } = useGetComposeServicesQuery(
    { id: application_id },
    { skip: !isDockerCompose || !application_id }
  );

  const isComposeMode = isDockerCompose && composeServices.length > 0;

  const handleFormSubmit = (values: any) => {
    if (isComposeMode) {
      values.domains = [];
    } else {
      values.compose_domains = [];
    }
    return onSubmit(values);
  };

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
          <div className="px-3 py-2 border border-border rounded-md bg-muted text-muted-foreground">
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
        <div className="px-3 py-2 border border-border rounded-md bg-muted text-muted-foreground overflow-hidden">
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
        href="https://docs.nixopus.com/apps/#docker-configuration"
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
          <form onSubmit={form.handleSubmit(handleFormSubmit)} className="space-y-6">
            <CollapsibleSection
              title={t('selfHost.configuration.sections.basicConfiguration')}
              icon={SettingsIcon}
              defaultOpen={true}
            >
              <div className="grid sm:grid-cols-2 gap-4">
                <FormInputField
                  form={form}
                  label={t('selfHost.configuration.fields.applicationName.label')}
                  name="name"
                  placeholder={t('selfHost.configuration.fields.applicationName.label')}
                />
                <EnvironmentInputField
                  form={form}
                  name="environment"
                  label={t('selfHost.configuration.fields.environment.label')}
                  required={false}
                />
                {build_pack !== BuildPack.Static && !isDockerCompose && (
                  <FormInputField
                    form={form}
                    label={t('selfHost.configuration.fields.port.label')}
                    name="port"
                    placeholder="3000"
                    validator={(value) => parsePort(value) !== null}
                  />
                )}
              </div>
              <div className="mt-4">
                {isDockerCompose && composeServices.length > 0 ? (
                  <ComposeDomainInput
                    form={form}
                    label={t('selfHost.configuration.fields.domain.label')}
                    name="compose_domains"
                    composeServices={composeServices}
                    placeholder="example.com"
                    required={false}
                    maxDomains={5}
                  />
                ) : (
                  <MultipleDomainInput
                    form={form}
                    label={t('selfHost.configuration.fields.domain.label')}
                    name="domains"
                    placeholder="example.com"
                    required={false}
                    maxDomains={5}
                  />
                )}
              </div>
            </CollapsibleSection>

            {build_pack !== BuildPack.Static && (
              <>
                <CollapsibleSection
                  title={t('selfHost.configuration.sections.dockerConfiguration')}
                  icon={ServerIcon}
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
  const {
    form,
    formFields,
    headerContent,
    handleCreate,
    buttonLabel,
    isCreatingProject,
    composeServices,
    isPreviewingCompose,
    previewError
  } = useQuickDeployForm({
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

                if (field.type === 'compose-domains') {
                  return (
                    <div key={field.key} className="sm:col-span-2">
                      <ComposeDomainInput
                        form={form}
                        label={field.label}
                        name={field.name}
                        composeServices={composeServices}
                        placeholder={field.placeholder}
                        required={field.required}
                        maxDomains={5}
                      />
                      {isPreviewingCompose && (
                        <p className="text-xs text-muted-foreground mt-1">
                          Discovering compose services...
                        </p>
                      )}
                      {previewError && (
                        <p className="text-xs text-amber-500 mt-1">{previewError}</p>
                      )}
                    </div>
                  );
                }

                if (field.type === 'multi-domains') {
                  return (
                    <div key={field.key} className="sm:col-span-2">
                      <MultipleDomainInput
                        form={form}
                        label={field.label}
                        name={field.name}
                        placeholder={field.placeholder}
                        required={field.required}
                        maxDomains={5}
                      />
                    </div>
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
              disabled={isCreatingProject || isPreviewingCompose}
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
