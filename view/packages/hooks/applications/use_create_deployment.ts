import { BuildPack, Environment } from '@/redux/types/deploy-form';
import { z } from 'zod';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { useWatch } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useWebSocket } from '@/packages/hooks/shared/socket-provider';
import { useRouter } from 'next/navigation';
import { useCreateDeploymentMutation } from '@/redux/services/deploy/applicationsApi';
import { toast } from 'sonner';
import { useAppSelector } from '@/redux/hooks';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

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
  DockerfilePath,
  base_path = '/'
}: DeploymentFormValues) {
  // Set default DockerfilePath based on build_pack
  const defaultDockerfilePath =
    DockerfilePath ||
    (build_pack === BuildPack.DockerCompose ? '/docker-compose.yml' : '/Dockerfile');
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
      .regex(/^[0-9]+$/, { message: t('selfHost.deployForm.validation.port.invalidFormat') })
      .optional(),
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
      .enum([BuildPack.Dockerfile, BuildPack.DockerCompose /* BuildPack.Static */])
      .refine(
        (value) => value === BuildPack.Dockerfile || value === BuildPack.DockerCompose,
        // Static build pack option commented out for deployment
        // value === BuildPack.Static,
        {
          message: t('selfHost.deployForm.validation.buildPack.invalidValue')
        }
      ),
    env_variables: z.record(z.string(), z.string()).optional().default({}),
    build_variables: z.record(z.string(), z.string()).optional().default({}),
    pre_run_commands: z.string().optional(),
    post_run_commands: z.string().optional(),
    DockerfilePath: z.string().optional().default(defaultDockerfilePath),
    base_path: z.string().optional().default(base_path)
  });

  // Static build pack option commented out for deployment - default to Dockerfile if Static is provided
  const validBuildPack = build_pack === BuildPack.Static ? BuildPack.Dockerfile : build_pack;

  const form = useForm<z.infer<typeof deploymentFormSchema>>({
    resolver: zodResolver(deploymentFormSchema),
    defaultValues: {
      application_name,
      environment,
      branch,
      port,
      domain,
      repository,
      build_pack: validBuildPack,
      env_variables,
      build_variables,
      pre_run_commands,
      post_run_commands,
      DockerfilePath: defaultDockerfilePath,
      base_path
    }
  });

  useEffect(() => {
    if (application_name) form.setValue('application_name', application_name);
    if (environment) form.setValue('environment', environment);
    if (branch) form.setValue('branch', branch);
    // Only set port if build_pack is not DockerCompose
    const isDockerCompose =
      build_pack === BuildPack.DockerCompose ||
      (build_pack as string) === 'docker-compose' ||
      (build_pack as string) === 'dockerCompose';
    if (port && !isDockerCompose) form.setValue('port', port);
    if (domain) form.setValue('domain', domain);
    if (repository) form.setValue('repository', repository);
    // Static build pack option commented out for deployment - default to Dockerfile if Static is provided
    if (build_pack) {
      const validBuildPack = build_pack === BuildPack.Static ? BuildPack.Dockerfile : build_pack;
      form.setValue('build_pack', validBuildPack);
    }
    if (env_variables && Object.keys(env_variables).length > 0)
      form.setValue('env_variables', env_variables);
    if (build_variables && Object.keys(build_variables).length > 0)
      form.setValue('build_variables', build_variables);
    if (pre_run_commands) form.setValue('pre_run_commands', pre_run_commands);
    if (post_run_commands) form.setValue('post_run_commands', post_run_commands);
    // Set DockerfilePath - use provided value or default based on build_pack
    if (DockerfilePath) {
      form.setValue('DockerfilePath', DockerfilePath);
    } else {
      const currentBuildPack = form.getValues('build_pack');
      const expectedPath =
        currentBuildPack === BuildPack.DockerCompose ? '/docker-compose.yml' : '/Dockerfile';
      form.setValue('DockerfilePath', expectedPath);
    }
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

  // Watch build_pack changes and update DockerfilePath accordingly
  const currentBuildPack = useWatch({ control: form.control, name: 'build_pack' });
  useEffect(() => {
    if (currentBuildPack) {
      const currentPath = form.getValues('DockerfilePath');
      const expectedPath =
        currentBuildPack === BuildPack.DockerCompose ? '/docker-compose.yml' : '/Dockerfile';
      // Only update if current path matches one of the defaults (user hasn't customized it)
      if (currentPath === '/Dockerfile' || currentPath === '/docker-compose.yml') {
        form.setValue('DockerfilePath', expectedPath);
      }
    }
  }, [currentBuildPack, form]);

  async function onSubmit(values: z.infer<typeof deploymentFormSchema>) {
    try {
      // Check if build_pack is DockerCompose
      const isDockerCompose =
        values.build_pack === BuildPack.DockerCompose ||
        (values.build_pack as string) === 'docker-compose' ||
        (values.build_pack as string) === 'dockerCompose';

      const data = await createDeployment({
        name: values.application_name,
        environment: values.environment,
        branch: values.branch,
        port:
          values.port && values.port.trim()
            ? parseInt(values.port, 10)
            : isDockerCompose
              ? 0
              : 3000,
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
