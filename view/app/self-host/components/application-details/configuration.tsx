'use client';
import React from 'react';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import FormInputField from '@/components/ui/form-input-field';
import { FormSelectTagInputField } from '@/components/ui/form-select-tag-field';
import { BuildPack, Environment } from '@/redux/types/deploy-form';
import useUpdateDeployment from '../../hooks/use_update_deployment';
import { Domain } from '@/redux/types/domain';
import { parsePort } from '../../utils/parsePort';

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
  application_id = ''
}: DeployConfigureProps) => {
  const { validateEnvVar, form, onSubmit, isLoading, domains } = useUpdateDeployment({
    name: application_name,
    pre_run_command: pre_run_commands,
    post_run_command: post_run_commands,
    build_variables,
    environment_variables: env_variables,
    port: parsePort(port) || 3000,
    force: true,
    id: application_id
  });

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <div className="grid sm:grid-cols-2 gap-4">
          <FormInputField
            form={form}
            label="Application Name"
            name="name"
            description="Application name"
            placeholder="Application name"
          />
          <FormInputField
            form={form}
            label="Port"
            name="port"
            description="Port on which your application will be available"
            placeholder="3000"
            validator={(value) => parsePort(value) !== null}
          />
          <FormInputField
            form={form}
            label="Dockerfile Path"
            name="dockerfile_path"
            description="Path of the dockerfile in case of mono repo"
            placeholder="Dockerfile"
            required={false}
          />
        </div>

        <div className="grid sm:grid-cols-2 gap-4">
          <FormSelectTagInputField
            form={form}
            label="Environment Variables"
            name="environment_variables"
            description="Type KEY=VALUE and press Enter to add"
            placeholder="NODE_ENV=production"
            required={false}
            validator={validateEnvVar}
            defaultValues={env_variables}
          />
          <FormSelectTagInputField
            form={form}
            label="Build Variables"
            name="build_variables"
            description="Type KEY=VALUE and press Enter to add"
            placeholder="DEBUG=false"
            required={false}
            validator={validateEnvVar}
            defaultValues={build_variables}
          />
        </div>

        <div className="grid sm:grid-cols-2 gap-4">
          <FormInputField
            form={form}
            label="Pre Run Commands"
            name="pre_run_command"
            description="Commands to run before deployment"
            placeholder="npm install"
            required={false}
          />
          <FormInputField
            form={form}
            label="Post Run Commands"
            name="post_run_command"
            description="Commands to run after deployment"
            placeholder="npm run test"
            required={false}
          />
        </div>

        <div className="grid sm:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Environment</label>
            <div className="px-3 py-2 border rounded-md bg-muted text-muted-foreground">
              {environment}
            </div>
            <p className="text-sm text-muted-foreground">Environment of the deployment</p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Branch</label>
            <div className="px-3 py-2 border rounded-md bg-muted text-muted-foreground">
              {branch}
            </div>
            <p className="text-sm text-muted-foreground">Branch from where you want to deploy</p>
          </div>
        </div>

        <div className="grid sm:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Domain</label>
            <div className="px-3 py-2 border rounded-md bg-muted text-muted-foreground">
              {domains?.find((dm: Domain) => dm.id === domain)?.name}
            </div>
            <p className="text-sm text-muted-foreground">
              Domain on which your application is available
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Build Pack</label>
            <div className="px-3 py-2 border rounded-md bg-muted text-muted-foreground">
              {build_pack}
            </div>
            <p className="text-sm text-muted-foreground">Build pack used for the application</p>
          </div>
        </div>

        <Button type="submit" className="w-full cursor-pointer" disabled={isLoading}>
          {isLoading ? 'Updating...' : 'Update Deployment'}
        </Button>
      </form>
    </Form>
  );
};
