import { BuildPack, Environment } from '@/redux/types/deploy-form';
import { z } from 'zod';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useWebSocket } from '@/hooks/socket-provider';
import { useRouter } from 'next/navigation';
import { useCreateDeploymentMutation } from '@/redux/services/deploy/applicationsApi';
import { toast } from 'sonner';
import { useAppSelector } from '@/redux/hooks';
import { useTranslation } from '@/hooks/use-translation';

interface DeploymentFormValues {
  application_name: string;
  environment: Environment;
  branch: string;
  port: string;
  domain: string;
  repository: string;
  build_pack: BuildPack;
  env_variables: Record<string, string>;
  build_variables: Record<string, string>;
  pre_run_commands: string;
  post_run_commands: string;
  DockerfilePath: string;
  base_path: string;
}

function useCreateDeployment({
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
}: DeploymentFormValues) {
  const { isReady, message, sendJsonMessage } = useWebSocket();
  const [createDeployment, { isLoading }] = useCreateDeploymentMutation();
  const router = useRouter();
  const { t } = useTranslation();

  const deploymentFormSchema = z.object({
    application_name: z
      .string()
      .min(3, { message: t('selfHost.deployForm.validation.applicationName.minLength') })
      .regex(/^[a-zA-Z0-9_-]+$/, {
        message: t('selfHost.deployForm.validation.applicationName.invalidFormat')
      }),
    environment: z
      .enum([Environment.Production, Environment.Staging, Environment.Development])
      .refine((value) => value === 'production' || value === 'staging' || value === 'development', {
        message: t('selfHost.deployForm.validation.environment.invalidValue')
      }),
    branch: z.string().min(3, { message: t('selfHost.deployForm.validation.branch.minLength') }),
    port: z
      .string()
      .regex(/^[0-9]+$/, { message: t('selfHost.deployForm.validation.port.invalidFormat') }),
    domain: z
      .string()
      .min(3, { message: t('selfHost.deployForm.validation.domain.minLength') })
      .regex(
        /^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9](\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9])*$/,
        {
          message: t('selfHost.deployForm.validation.domain.invalidFormat')
        }
      ),
    repository: z
      .string()
      .min(3, { message: t('selfHost.deployForm.validation.repository.minLength') })
      .regex(/^[a-zA-Z0-9_-]+$/, {
        message: t('selfHost.deployForm.validation.repository.invalidFormat')
      }),
    build_pack: z
      .enum([BuildPack.Dockerfile, BuildPack.DockerCompose, BuildPack.Static])
      .refine(
        (value) =>
          value === BuildPack.Dockerfile ||
          value === BuildPack.DockerCompose ||
          value === BuildPack.Static,
        {
          message: t('selfHost.deployForm.validation.buildPack.invalidValue')
        }
      ),
    env_variables: z.record(z.string(), z.string()).optional().default({}),
    build_variables: z.record(z.string(), z.string()).optional().default({}),
    pre_run_commands: z.string().optional(),
    post_run_commands: z.string().optional(),
    DockerfilePath: z.string().optional().default(DockerfilePath),
    base_path: z.string().optional().default(base_path)
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
      post_run_commands,
      DockerfilePath,
      base_path
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
    if (DockerfilePath) form.setValue('DockerfilePath', DockerfilePath);
    if (base_path) form.setValue('base_path', base_path);
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
    post_run_commands,
    DockerfilePath,
    base_path
  ]);

  async function onSubmit(values: z.infer<typeof deploymentFormSchema>) {
    try {
      const data = await createDeployment({
        name: values.application_name,
        environment: values.environment,
        branch: values.branch,
        port: parseInt(values.port, 10),
        domain: values.domain,
        repository: values.repository,
        build_pack: values.build_pack,
        environment_variables: values.env_variables,
        build_variables: values.build_variables,
        pre_run_command: values.pre_run_commands as string,
        post_run_command: values.post_run_commands as string,
        dockerfile_path: values.DockerfilePath,
        base_path: values.base_path
      }).unwrap();

      if (data?.deployments?.[0]?.id) {
        router.push('/self-host/application/' + data.id + '/deployments/' + data.deployments[0].id);
      } else {
        router.push('/self-host/application/' + data.id + '?logs=true');
      }
    } catch (error) {
      toast.error(t('selfHost.deployForm.errors.createFailed'));
    }
  }

  const validateEnvVar = (
    input: string
  ): { isValid: boolean; error?: string; key?: string; value?: string } => {
    if (!input.trim())
      return { isValid: false, error: t('selfHost.deployForm.validation.envVariables.emptyInput') };

    const regex = /^([^=]+)=(.*)$/;
    const isValid = regex.test(input);

    if (!isValid) {
      return {
        isValid: false,
        error: t('selfHost.deployForm.validation.envVariables.invalidFormat')
      };
    }

    const [, key] = input.match(regex) as RegExpMatchArray;

    if (!key.trim()) {
      return { isValid: false, error: t('selfHost.deployForm.validation.envVariables.emptyKey') };
    }

    return {
      isValid: true,
      key: key.trim(),
      value: input.substring(key.length + 1)
    };
  };

  const parsePort = (port: string) => {
    const parsedPort = parseInt(port, 10);
    return isNaN(parsedPort) ? null : parsedPort;
  };

  return { validateEnvVar, deploymentFormSchema, form, onSubmit, parsePort };
}

export default useCreateDeployment;
