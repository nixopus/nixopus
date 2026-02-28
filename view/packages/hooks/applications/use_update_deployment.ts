import { useEffect, useRef } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { useWebSocket } from '@/packages/hooks/shared/socket-provider';
import { useUpdateDeploymentMutation } from '@/redux/services/deploy/applicationsApi';
import { UpdateDeploymentRequest, Environment } from '@/redux/types/applications';
import { useGetAllDomainsQuery } from '@/redux/services/settings/domainsApi';
import { parsePort, SHARED_DOMAIN_REGEX } from '@/packages/utils/util';
import { useAppSelector } from '@/redux/hooks';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

interface ComposeDomainEntry {
  domain: string;
  service_name: string;
  port: number;
}

interface UseUpdateDeploymentProps {
  name?: string;
  environment?: string;
  pre_run_command?: string;
  post_run_command?: string;
  build_variables?: Record<string, string>;
  environment_variables?: Record<string, string>;
  port?: number;
  id?: string;
  force?: boolean;
  DockerfilePath?: string;
  base_path?: string;
  domains?: string[];
  compose_domains?: ComposeDomainEntry[];
}

function useUpdateDeployment({
  name = '',
  environment = '',
  pre_run_command = '',
  post_run_command = '',
  build_variables = {},
  environment_variables = {},
  port = 3000,
  id = '',
  force = false,
  DockerfilePath = '/Dockerfile',
  base_path = '/',
  domains = [],
  compose_domains = []
}: UseUpdateDeploymentProps = {}) {
  const { t } = useTranslation();
  const { isReady, message, sendJsonMessage } = useWebSocket();
  const [updateDeployment, { isLoading }] = useUpdateDeploymentMutation();
  const router = useRouter();
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const { data: organizationDomains } = useGetAllDomainsQuery();

  const deploymentFormSchema = z.object({
    name: z
      .string()
      .min(3, { message: t('selfHost.deployForm.validation.applicationName.minLength') })
      .regex(/^[a-zA-Z0-9_-]+$/, {
        message: t('selfHost.deployForm.validation.applicationName.invalidFormat')
      })
      .optional(),
    environment: z
      .string()
      .regex(/^[a-z0-9]+(-[a-z0-9]+)*$/, {
        message: t('selfHost.deployForm.validation.environment.invalidValue')
      })
      .optional(),
    pre_run_command: z.string().optional(),
    post_run_command: z.string().optional(),
    build_variables: z.record(z.string(), z.string()).optional().default({}),
    environment_variables: z.record(z.string(), z.string()).optional().default({}),
    port: z.string().optional(),
    id: z.string().optional(),
    force: z.boolean().optional().default(false),
    DockerfilePath: z.string().optional().default(DockerfilePath),
    base_path: z.string().optional().default(base_path),
    domains: z
      .array(z.string())
      .optional()
      .default([])
      .refine(
        (arr) => {
          const nonEmpty = arr?.filter((d) => d && d.trim() !== '') ?? [];
          return nonEmpty.length <= 5;
        },
        { message: t('selfHost.deployForm.validation.domain.maxDomains') }
      )
      .refine(
        (arr) => {
          const nonEmpty = arr?.filter((d) => d && d.trim() !== '') ?? [];
          const unique = new Set(nonEmpty.map((d) => d.trim().toLowerCase()));
          return unique.size === nonEmpty.length;
        },
        { message: t('selfHost.deployForm.validation.domain.duplicate') }
      )
      .refine(
        (arr) => {
          const nonEmpty = arr?.filter((d) => d && d.trim() !== '') ?? [];
          return nonEmpty.every((d) => !d.trim() || SHARED_DOMAIN_REGEX.test(d.trim()));
        },
        { message: t('selfHost.deployForm.validation.domain.invalidFormat') }
      ),
    compose_domains: z
      .array(
        z.object({
          domain: z.string(),
          service_name: z.string().optional().default(''),
          port: z.number().optional().default(0)
        })
      )
      .optional()
      .default([])
      .refine(
        (arr) => {
          const nonEmpty = arr?.filter((d) => d.domain && d.domain.trim() !== '') ?? [];
          return nonEmpty.length <= 5;
        },
        { message: t('selfHost.deployForm.validation.domain.maxDomains') }
      )
      .refine(
        (arr) => {
          const nonEmpty = arr?.filter((d) => d.domain && d.domain.trim() !== '') ?? [];
          const unique = new Set(nonEmpty.map((d) => d.domain.trim().toLowerCase()));
          return unique.size === nonEmpty.length;
        },
        { message: t('selfHost.deployForm.validation.domain.duplicate') }
      )
      .refine(
        (arr) => {
          const nonEmpty = arr?.filter((d) => d.domain && d.domain.trim() !== '') ?? [];
          return nonEmpty.every(
            (d) => !d.domain.trim() || SHARED_DOMAIN_REGEX.test(d.domain.trim())
          );
        },
        { message: t('selfHost.deployForm.validation.domain.invalidFormat') }
      )
  });

  const form = useForm<z.infer<typeof deploymentFormSchema>>({
    resolver: zodResolver(deploymentFormSchema),
    defaultValues: {
      name,
      environment,
      pre_run_command,
      post_run_command,
      build_variables,
      environment_variables,
      port: port.toString(),
      id,
      force,
      DockerfilePath,
      base_path,
      domains: domains.length > 0 ? domains : [''],
      compose_domains:
        compose_domains.length > 0 ? compose_domains : [{ domain: '', service_name: '', port: 0 }]
    }
  });

  // Track which application id the form has been initialized for
  // This prevents resetting user changes on re renders while allowing
  // initialization when switching to a different application
  const initializedForIdRef = useRef<string | null>(null);

  useEffect(() => {
    // Wait for actual application data to load (name indicates data is ready)
    if (!id || !name) {
      return;
    }

    // If already initialized for this specific application, don't reset values
    if (initializedForIdRef.current === id) {
      return;
    }

    // Initialize form with server values
    form.setValue('name', name);
    form.setValue('domains', domains && domains.length > 0 ? domains : ['']);
    form.setValue(
      'compose_domains',
      compose_domains && compose_domains.length > 0
        ? compose_domains
        : [{ domain: '', service_name: '', port: 0 }]
    );
    if (environment) form.setValue('environment', environment);
    if (pre_run_command) form.setValue('pre_run_command', pre_run_command);
    if (post_run_command) form.setValue('post_run_command', post_run_command);
    if (build_variables && Object.keys(build_variables).length > 0)
      form.setValue('build_variables', build_variables);
    if (environment_variables && Object.keys(environment_variables).length > 0)
      form.setValue('environment_variables', environment_variables);
    if (port) form.setValue('port', port.toString());
    form.setValue('id', id);
    if (DockerfilePath) form.setValue('DockerfilePath', DockerfilePath);
    if (base_path) form.setValue('base_path', base_path);
    form.setValue('force', force);

    initializedForIdRef.current = id;
  }, [id, name, domains, compose_domains]);

  async function onSubmit(values: z.infer<typeof deploymentFormSchema>) {
    try {
      const filteredDomains = values.domains?.filter((d) => d && d.trim() !== '') ?? [];
      const filteredComposeDomains =
        values.compose_domains?.filter((d) => d.domain && d.domain.trim() !== '') ?? [];

      const hasComposeDomains = filteredComposeDomains.length > 0;

      const updateData: UpdateDeploymentRequest = {
        name: values.name,
        environment: values.environment as Environment | undefined,
        pre_run_command: values.pre_run_command,
        post_run_command: values.post_run_command,
        build_variables: values.build_variables,
        environment_variables: values.environment_variables,
        port: parsePort(values.port?.toString() || '3000') || 3000,
        id: values.id,
        force: values.force,
        dockerfile_path: values.DockerfilePath,
        base_path: values.base_path,
        domains: hasComposeDomains ? undefined : filteredDomains,
        compose_domains: hasComposeDomains ? filteredComposeDomains : undefined
      };

      const data = await updateDeployment(updateData).unwrap();

      if (data?.id) {
        router.push('/apps/application/' + data.id);
        toast.success(t('selfHost.deployForm.success.update'));
      }
    } catch (error) {
      toast.error(t('selfHost.deployForm.errors.updateFailed'));
    }
  }

  const validateEnvVar = (
    input: string
  ): { isValid: boolean; error?: string; key?: string; value?: string } => {
    if (!input.trim())
      return {
        isValid: false,
        error: t('selfHost.deployForm.validation.envVariables.emptyInput')
      };

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
      return {
        isValid: false,
        error: t('selfHost.deployForm.validation.envVariables.emptyKey')
      };
    }

    return {
      isValid: true,
      key: key.trim(),
      value: input.substring(key.length + 1)
    };
  };

  return {
    validateEnvVar,
    deploymentFormSchema,
    form,
    onSubmit,
    isLoading,
    organizationDomains
  };
}

export default useUpdateDeployment;
