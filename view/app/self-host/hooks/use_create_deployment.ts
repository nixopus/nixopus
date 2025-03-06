import { BuildPack, Environment } from '@/redux/types/deploy-form';
import { z } from 'zod';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useGetAllDomainsQuery } from '@/redux/services/settings/domainsApi';
import { useWebSocket } from '@/hooks/socket_provider';

interface DeploymentFormValues {
  application_name: string;
  environment: Environment;
  branch: string;
  port: number;
  domain: string;
  repository: string;
  build_pack: BuildPack;
  env_variables: Record<string, string>;
  build_variables: Record<string, string>;
  pre_run_commands: string;
  post_run_commands: string;
}

function useCreateDeployment({
  application_name = '',
  environment = Environment.Production,
  branch = '',
  port = 3000,
  domain = '',
  repository,
  build_pack = BuildPack.Dockerfile,
  env_variables = {},
  build_variables = {},
  pre_run_commands = '',
  post_run_commands = ''
}: DeploymentFormValues) {
  const { data: domains } = useGetAllDomainsQuery();
  const { isReady, message, sendJsonMessage } = useWebSocket();

  console.log(message)

  const deploymentFormSchema = z.object({
    application_name: z
      .string()
      .min(3, { message: 'Application name must be at least 3 characters.' })
      .regex(/^[a-zA-Z0-9_-]+$/, { message: 'Application name must be a valid name.' }),
    environment: z
      .enum([Environment.Production, Environment.Staging, Environment.Development])
      .refine((value) => value === 'production' || value === 'staging' || value === 'development', {
        message: 'Environment name must be production, staging, or development.'
      }),
    branch: z
      .string()
      .min(3, { message: 'Branch name must be at least 3 characters.' })
      .regex(/^[a-zA-Z0-9_-]+$/, { message: 'Branch name must be a valid name.' }),
    port: z.number().min(1024, { message: 'Port must be at least 1024.' }),
    domain: z
      .string()
      .min(3, { message: 'Domain name must be at least 3 characters.' })
      .regex(/^[a-zA-Z0-9.-]+$/, { message: 'Domain name must be a valid domain name.' }),
    repository: z
      .string()
      .min(3, { message: 'Repository name must be at least 3 characters.' })
      .regex(/^[a-zA-Z0-9_-]+$/, { message: 'Repository name must be a valid name.' }),
    build_pack: z
      .enum([BuildPack.Dockerfile, BuildPack.DockerCompose, BuildPack.Static])
      .refine(
        (value) =>
          value === BuildPack.Dockerfile ||
          value === BuildPack.DockerCompose ||
          value === BuildPack.Static,
        {
          message: 'Build pack must be Dockerfile, DockerCompose, or Static.'
        }
      ),
    env_variables: z.record(z.string(), z.string()).optional().default({}),
    build_variables: z.record(z.string(), z.string()).optional().default({}),
    pre_run_commands: z.string().optional(),
    post_run_commands: z.string().optional()
  });

  const form = useForm<z.infer<typeof deploymentFormSchema>>({
    resolver: zodResolver(deploymentFormSchema),
    defaultValues: {
      application_name,
      environment,
      branch,
      port,
      domain,
      repository,
      build_pack,
      env_variables,
      build_variables,
      pre_run_commands,
      post_run_commands
    }
  });

  useEffect(() => {
    if (application_name) form.setValue('application_name', application_name);
    if (environment) form.setValue('environment', environment);
    if (branch) form.setValue('branch', branch);
    if (port) form.setValue('port', port);
    if (domain) form.setValue('domain', domain);
    if (repository) form.setValue('repository', repository);
    if (build_pack) form.setValue('build_pack', build_pack);
    if (env_variables && Object.keys(env_variables).length > 0)
      form.setValue('env_variables', env_variables);
    if (build_variables && Object.keys(build_variables).length > 0)
      form.setValue('build_variables', build_variables);
    if (pre_run_commands) form.setValue('pre_run_commands', pre_run_commands);
    if (post_run_commands) form.setValue('post_run_commands', post_run_commands);
  }, [
    form,
    application_name,
    environment,
    branch,
    port,
    domain,
    repository,
    build_pack,
    env_variables,
    build_variables,
    pre_run_commands,
    post_run_commands
  ]);

  function onSubmit(values: z.infer<typeof deploymentFormSchema>) {
    sendJsonMessage({
      action: 'deploy',
      data: {
        name: values.application_name,
        environment: values.environment,
        branch: values.branch,
        port: values.port,
        domain_id: values.domain,
        repository: values.repository,
        build_pack: values.build_pack,
        env_variables: values.env_variables,
        build_variables: values.build_variables,
        pre_run_commands: values.pre_run_commands,
        post_run_commands: values.post_run_commands
      }
    })
  }

  const validateEnvVar = (
    input: string
  ): { isValid: boolean; error?: string; key?: string; value?: string } => {
    if (!input.trim()) return { isValid: false, error: 'Input cannot be empty' };

    const regex = /^([^=]+)=(.*)$/;
    const isValid = regex.test(input);

    if (!isValid) {
      return { isValid: false, error: 'Must be in format KEY=VALUE' };
    }

    const [, key] = input.match(regex) as RegExpMatchArray;

    if (!key.trim()) {
      return { isValid: false, error: 'Key cannot be empty' };
    }

    return {
      isValid: true,
      key: key.trim(),
      value: input.substring(key.length + 1)
    };
  };

  return { validateEnvVar, deploymentFormSchema, form, onSubmit, domains };
}

export default useCreateDeployment;
