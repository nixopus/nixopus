'use client';
import React from 'react';
import { UseFormReturn } from 'react-hook-form';
import FormInputField from '@/components/ui/form-input-field';
import FormSelectField from '@/components/ui/form-select-field';
import { FormSelectTagInputField } from '@/components/ui/form-select-tag-field';
import FormTextareaField from '@/components/ui/form-textarea-field';
import { BuildPack, ComposeSourceType } from '@/redux/types/deploy-form';
import { useTranslation } from '@/hooks/use-translation';
import { Skeleton } from '@/components/ui/skeleton';

interface DeployFormFieldsProps {
  form: UseFormReturn<any>;
  currentStepId: string;
  isDockerCompose: boolean;
  watchedComposeSource: ComposeSourceType | undefined;
  parsePort: (value: string) => number | null;
  validateEnvVar: (input: string) => {
    isValid: boolean;
    error?: string;
    key?: string;
    value?: string;
  };
  isLoadingBranches: boolean;
  availableBranches: { label: string; value: string }[];
}

export const DeployFormFields = ({
  form,
  currentStepId,
  isDockerCompose,
  watchedComposeSource,
  parsePort,
  validateEnvVar,
  isLoadingBranches,
  availableBranches
}: DeployFormFieldsProps) => {
  const { t } = useTranslation();

  switch (currentStepId) {
    // Step 1: Application (4 fields: name, environment, build_pack, domain)
    case 'application':
      return (
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
                label: t('selfHost.deployForm.fields.buildPack.options.dockerCompose'),
                value: BuildPack.DockerCompose
              }
            ]}
          />
          <FormInputField
            form={form}
            label={t('selfHost.deployForm.fields.domain.label')}
            name="domain"
            placeholder={t('selfHost.deployForm.fields.domain.placeholder')}
            required={true}
          />
        </div>
      );

    // Step 2: Source (varies by build pack, max 4 fields)
    case 'source':
      if (isDockerCompose) {
        return (
          <div className="space-y-4">
            <div className="grid sm:grid-cols-2 gap-4">
              <FormSelectField
                form={form}
                label={t('selfHost.deployForm.fields.composeSource.label')}
                name="compose_source"
                placeholder={t('selfHost.deployForm.fields.composeSource.placeholder')}
                selectOptions={[
                  {
                    label: t('selfHost.deployForm.fields.composeSource.options.repository'),
                    value: ComposeSourceType.Repository
                  },
                  {
                    label: t('selfHost.deployForm.fields.composeSource.options.url'),
                    value: ComposeSourceType.URL
                  },
                  {
                    label: t('selfHost.deployForm.fields.composeSource.options.raw'),
                    value: ComposeSourceType.Raw
                  }
                ]}
              />
              <FormInputField
                form={form}
                label={t('selfHost.deployForm.fields.port.label')}
                name="port"
                placeholder={t('selfHost.deployForm.fields.port.placeholder')}
                validator={(value) => !value || parsePort(value) !== null}
                required={false}
              />
            </div>

            {/* Repository source: show branch and compose file path */}
            {watchedComposeSource === ComposeSourceType.Repository && (
              <div className="grid sm:grid-cols-2 gap-4">
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
                <FormInputField
                  form={form}
                  label={t('selfHost.deployForm.fields.composeFilePath.label')}
                  name="DockerfilePath"
                  placeholder={t('selfHost.deployForm.fields.composeFilePath.placeholder')}
                  required={false}
                />
              </div>
            )}

            {/* URL source: show URL input */}
            {watchedComposeSource === ComposeSourceType.URL && (
              <FormInputField
                form={form}
                label={t('selfHost.deployForm.fields.composeFileUrl.label')}
                name="compose_file_url"
                placeholder={t('selfHost.deployForm.fields.composeFileUrl.placeholder')}
                required={true}
              />
            )}

            {/* Raw source: show textarea */}
            {watchedComposeSource === ComposeSourceType.Raw && (
              <FormTextareaField
                form={form}
                label={t('selfHost.deployForm.fields.composeFileContent.label')}
                name="compose_file_content"
                placeholder={t('selfHost.deployForm.fields.composeFileContent.placeholder')}
                required={true}
                rows={10}
              />
            )}
          </div>
        );
      }

      // Dockerfile: branch, port, base_path, dockerfile_path
      return (
        <div className="grid sm:grid-cols-2 gap-4">
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
          <FormInputField
            form={form}
            label={t('selfHost.deployForm.fields.port.label')}
            name="port"
            placeholder={t('selfHost.deployForm.fields.port.placeholder')}
            validator={(value) => parsePort(value) !== null}
            required={true}
          />
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
      );

    // Step 3: Variables (4 fields: env_vars, build_vars, pre_run, post_run)
    case 'variables':
      return (
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
      );

    default:
      return null;
  }
};
