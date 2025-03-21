'use client';
import React from 'react';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import FormInputField from '@/components/ui/form-input-field';
import FormSelectField from '@/components/ui/form-select-field';
import { FormSelectTagInputField } from '@/components/ui/form-select-tag-field';
import { BuildPack, Environment } from '@/redux/types/deploy-form';
import useCreateDeployment from '../../hooks/use_create_deployment';

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
  DockerfilePath = '/Dockerfile'
}: DeployFormProps) => {
  const { validateEnvVar, form, onSubmit, domains, parsePort } = useCreateDeployment({
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
    DockerfilePath
  });

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <div className="grid sm:grid-cols-2 gap-4">
          <FormInputField
            form={form}
            label="Application Name"
            name="application_name"
            description="Application name"
            placeholder="Application name"
          />
          <FormSelectField
            form={form}
            label="Environment"
            name="environment"
            description="Environments helps you deploy to different spaces"
            placeholder="Environment"
            selectOptions={[
              { label: 'Staging', value: 'staging' },
              { label: 'Production', value: 'production' },
              { label: 'Development', value: 'development' }
            ]}
          />
        </div>
        <div className="grid sm:grid-cols-2 gap-4">
          <FormInputField
            form={form}
            label="Branch"
            name="branch"
            description="Branch from where you want to deploy"
            placeholder="Branch"
          />
          <FormInputField
            form={form}
            label="Dockerfile Path"
            name="dockerfile_path"
            description="Path of the dockerfile in case of mono repo"
            placeholder="Dockerfile"
            required={false}
          />
          <FormInputField
            form={form}
            label="Port"
            name="port"
            description="Port on which your application will be available"
            placeholder="3000"
            validator={(value) => parsePort(value) !== null}
          />
        </div>
        <div className="grid sm:grid-cols-2 gap-4">
          <FormSelectField
            form={form}
            label="Domain"
            name="domain"
            description="Domain on which your application will be available"
            placeholder="Domain"
            selectOptions={domains?.map((domain) => ({ label: domain.name, value: domain.id }))}
          />
          <FormSelectField
            form={form}
            label="Build Pack"
            name="build_pack"
            description="Choose the one that best fits your application structure"
            placeholder="Build Pack"
            selectOptions={[
              { label: 'Dockerfile', value: BuildPack.Dockerfile },
              { label: 'DockerCompose', value: BuildPack.DockerCompose },
              { label: 'Static', value: BuildPack.Static }
            ]}
          />
        </div>
        <div className="grid sm:grid-cols-2 gap-4">
          <FormSelectTagInputField
            form={form}
            label="Environment Variables"
            name="env_variables"
            description="Type KEY=VALUE and press Enter to add"
            placeholder="NODE_ENV=production"
            required={false}
            validator={validateEnvVar}
          />
          <FormSelectTagInputField
            form={form}
            label="Build Variables"
            name="build_variables"
            description="Type KEY=VALUE and press Enter to add"
            placeholder="DEBUG=false"
            required={false}
            validator={validateEnvVar}
          />
        </div>
        <div className="grid sm:grid-cols-2 gap-4">
          <FormInputField
            form={form}
            label="Pre Run Commands"
            name="pre_run_commands"
            description="Commands to run before deployment"
            placeholder="npm install"
            required={false}
          />
          <FormInputField
            form={form}
            label="Post Run Commands"
            name="post_run_commands"
            description="Commands to run after deployment"
            placeholder="npm run test"
            required={false}
          />
        </div>
        <Button className="w-full cursor-pointer">Deploy</Button>
      </form>
    </Form>
  );
};
