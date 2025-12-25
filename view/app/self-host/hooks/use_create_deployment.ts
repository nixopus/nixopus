import { BuildPack, ComposeSourceType, Environment } from '@/redux/types/deploy-form';
import { z } from 'zod';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useRouter } from 'next/navigation';
import { useCreateDeploymentMutation } from '@/redux/services/deploy/applicationsApi';
import { toast } from 'sonner';
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
  compose_source?: ComposeSourceType;
  compose_file_url?: string;
  compose_file_content?: string;
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
  DockerfilePath = '',
  base_path = '/',
  compose_source = ComposeSourceType.Repository,
  compose_file_url = '',
  compose_file_content = ''
}: DeploymentFormValues) {
  const [createDeployment, { isLoading }] = useCreateDeploymentMutation();
  const router = useRouter();
  const { t } = useTranslation();

  const deploymentFormSchema = z
    .object({
      application_name: z
        .string()
        .min(3, { message: t('selfHost.deployForm.validation.applicationName.minLength') })
        .regex(/^[a-zA-Z0-9_-]+$/, {
          message: t('selfHost.deployForm.validation.applicationName.invalidFormat')
        }),
      environment: z
        .enum([Environment.Production, Environment.Staging, Environment.Development])
        .refine(
          (value) => value === 'production' || value === 'staging' || value === 'development',
          {
            message: t('selfHost.deployForm.validation.environment.invalidValue')
          }
        ),
      branch: z.string().optional(),
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
      repository: z.string().optional(),
      build_pack: z
        .enum([BuildPack.Dockerfile, BuildPack.DockerCompose])
        .refine((value) => value === BuildPack.Dockerfile || value === BuildPack.DockerCompose, {
          message: t('selfHost.deployForm.validation.buildPack.invalidValue')
        }),
      env_variables: z.record(z.string(), z.string()).optional().default({}),
      build_variables: z.record(z.string(), z.string()).optional().default({}),
      pre_run_commands: z.string().optional(),
      post_run_commands: z.string().optional(),
      DockerfilePath: z.string().optional(),
      base_path: z.string().optional().default(base_path),
      compose_source: z.nativeEnum(ComposeSourceType).optional(),
      compose_file_url: z.string().optional(),
      compose_file_content: z.string().optional()
    })
    .superRefine((data, ctx) => {
      if (data.build_pack === BuildPack.Dockerfile) {
        if (!data.repository || data.repository.length < 3) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: t('selfHost.deployForm.validation.repository.minLength'),
            path: ['repository']
          });
        }
        if (!data.branch || data.branch.length < 1) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: t('selfHost.deployForm.validation.branch.minLength'),
            path: ['branch']
          });
        }
        if (!data.port) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: t('selfHost.deployForm.validation.port.invalidFormat'),
            path: ['port']
          });
        }
      }

      if (data.build_pack === BuildPack.DockerCompose) {
        const source = data.compose_source || ComposeSourceType.Repository;

        if (source === ComposeSourceType.Repository) {
          if (!data.repository || data.repository.length < 3) {
            ctx.addIssue({
              code: z.ZodIssueCode.custom,
              message: t('selfHost.deployForm.validation.repository.minLength'),
              path: ['repository']
            });
          }
          if (!data.branch || data.branch.length < 1) {
            ctx.addIssue({
              code: z.ZodIssueCode.custom,
              message: t('selfHost.deployForm.validation.branch.minLength'),
              path: ['branch']
            });
          }
        }

        if (source === ComposeSourceType.URL) {
          if (
            !data.compose_file_url ||
            (!data.compose_file_url.startsWith('http://') &&
              !data.compose_file_url.startsWith('https://'))
          ) {
            ctx.addIssue({
              code: z.ZodIssueCode.custom,
              message: t('selfHost.deployForm.validation.composeFileUrl.invalidFormat'),
              path: ['compose_file_url']
            });
          }
        }

        if (source === ComposeSourceType.Raw) {
          if (!data.compose_file_content || data.compose_file_content.length < 10) {
            ctx.addIssue({
              code: z.ZodIssueCode.custom,
              message: t('selfHost.deployForm.validation.composeFileContent.tooShort'),
              path: ['compose_file_content']
            });
          }
        }
      }
    });

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
      DockerfilePath,
      base_path,
      compose_source,
      compose_file_url,
      compose_file_content
    }
  });

  useEffect(() => {
    if (application_name) form.setValue('application_name', application_name);
    if (environment) form.setValue('environment', environment);
    if (branch) form.setValue('branch', branch);
    if (port) form.setValue('port', port);
    if (domain) form.setValue('domain', domain);
    if (repository) form.setValue('repository', repository);
    if (build_pack)
      form.setValue(
        'build_pack',
        build_pack === BuildPack.Static ? BuildPack.Dockerfile : build_pack
      );
    if (env_variables && Object.keys(env_variables).length > 0)
      form.setValue('env_variables', env_variables);
    if (build_variables && Object.keys(build_variables).length > 0)
      form.setValue('build_variables', build_variables);
    if (pre_run_commands) form.setValue('pre_run_commands', pre_run_commands);
    if (post_run_commands) form.setValue('post_run_commands', post_run_commands);
    if (DockerfilePath) form.setValue('DockerfilePath', DockerfilePath);
    if (base_path) form.setValue('base_path', base_path);
    if (compose_source) form.setValue('compose_source', compose_source);
    if (compose_file_url) form.setValue('compose_file_url', compose_file_url);
    if (compose_file_content) form.setValue('compose_file_content', compose_file_content);
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
    base_path,
    compose_source,
    compose_file_url,
    compose_file_content
  ]);

  // Update DockerfilePath default when build pack changes
  const watchedBuildPack = form.watch('build_pack');
  useEffect(() => {
    const currentPath = form.getValues('DockerfilePath');
    const isDefaultDockerfile =
      currentPath === '/Dockerfile' || currentPath === 'Dockerfile' || currentPath === '';
    const isDefaultCompose =
      currentPath === 'docker-compose.yml' || currentPath === 'docker-compose.yaml';

    // Only update if the current value is a default value (user hasn't customized it)
    if (watchedBuildPack === BuildPack.DockerCompose && isDefaultDockerfile) {
      form.setValue('DockerfilePath', 'docker-compose.yml');
    } else if (
      watchedBuildPack === BuildPack.Dockerfile &&
      (isDefaultCompose || currentPath === '')
    ) {
      form.setValue('DockerfilePath', 'Dockerfile');
    }
  }, [watchedBuildPack, form]);

  async function onSubmit(values: z.infer<typeof deploymentFormSchema>) {
    try {
      const isCompose = values.build_pack === BuildPack.DockerCompose;
      const composeSource = values.compose_source || ComposeSourceType.Repository;

      const data = await createDeployment({
        name: values.application_name,
        environment: values.environment,
        branch: isCompose && composeSource !== ComposeSourceType.Repository ? '' : values.branch,
        port: parseInt(values.port || '80', 10),
        domain: values.domain,
        repository:
          isCompose && composeSource !== ComposeSourceType.Repository ? '' : values.repository,
        build_pack: values.build_pack,
        environment_variables: values.env_variables || {},
        build_variables: values.build_variables || {},
        pre_run_command: values.pre_run_commands as string,
        post_run_command: values.post_run_commands as string,
        dockerfile_path: values.DockerfilePath || (isCompose ? 'docker-compose.yml' : 'Dockerfile'),
        base_path: values.base_path,
        compose_file_url:
          isCompose && composeSource === ComposeSourceType.URL ? values.compose_file_url : '',
        compose_file_content:
          isCompose && composeSource === ComposeSourceType.Raw ? values.compose_file_content : ''
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

  return { validateEnvVar, deploymentFormSchema, form, onSubmit, parsePort, isLoading };
}

export default useCreateDeployment;
