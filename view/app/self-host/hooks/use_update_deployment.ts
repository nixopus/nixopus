import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { useWebSocket } from '@/hooks/socket_provider';
import { useUpdateDeploymentMutation } from '@/redux/services/deploy/applicationsApi';
import { UpdateDeploymentRequest } from '@/redux/types/applications';
import { useGetAllDomainsQuery } from '@/redux/services/settings/domainsApi';
import { parsePort } from '../utils/parsePort';

interface UseUpdateDeploymentProps {
  name?: string;
  pre_run_command?: string;
  post_run_command?: string;
  build_variables?: Record<string, string>;
  environment_variables?: Record<string, string>;
  port?: number;
  id?: string;
  force?: boolean;
  DockerfilePath?: string;
}

function useUpdateDeployment({
  name = '',
  pre_run_command = '',
  post_run_command = '',
  build_variables = {},
  environment_variables = {},
  port = 3000,
  id = '',
  force = false,
  DockerfilePath = '/Dockerfile'
}: UseUpdateDeploymentProps = {}) {
  const { isReady, message, sendJsonMessage } = useWebSocket();
  const [updateDeployment, { isLoading }] = useUpdateDeploymentMutation();
  const router = useRouter();
  const { data: domains } = useGetAllDomainsQuery();

  const deploymentFormSchema = z.object({
    name: z
      .string()
      .min(3, { message: 'Application name must be at least 3 characters.' })
      .regex(/^[a-zA-Z0-9_-]+$/, { message: 'Application name must be a valid name.' })
      .optional(),
    pre_run_command: z.string().optional(),
    post_run_command: z.string().optional(),
    build_variables: z.record(z.string(), z.string()).optional().default({}),
    environment_variables: z.record(z.string(), z.string()).optional().default({}),
    port: z.string().optional(),
    id: z.string().optional(),
    force: z.boolean().optional().default(false),
    DockerfilePath: z.string().optional().default(DockerfilePath)
  });

  const form = useForm<z.infer<typeof deploymentFormSchema>>({
    resolver: zodResolver(deploymentFormSchema),
    defaultValues: {
      name,
      pre_run_command,
      post_run_command,
      build_variables,
      environment_variables,
      port: port.toString(),
      id,
      force,
      DockerfilePath
    }
  });

  useEffect(() => {
    if (name) form.setValue('name', name);
    if (pre_run_command) form.setValue('pre_run_command', pre_run_command);
    if (post_run_command) form.setValue('post_run_command', post_run_command);
    if (build_variables && Object.keys(build_variables).length > 0)
      form.setValue('build_variables', build_variables);
    if (environment_variables && Object.keys(environment_variables).length > 0)
      form.setValue('environment_variables', environment_variables);
    if (port) form.setValue('port', port.toString());
    if (id) form.setValue('id', id);
    if (DockerfilePath) form.setValue('DockerfilePath', DockerfilePath);
    form.setValue('force', force);
  }, [
    form,
    name,
    pre_run_command,
    post_run_command,
    build_variables,
    environment_variables,
    port,
    id,
    force,
    DockerfilePath
  ]);

  async function onSubmit(values: z.infer<typeof deploymentFormSchema>) {
    try {
      const updateData: UpdateDeploymentRequest = {
        name: values.name,
        pre_run_command: values.pre_run_command,
        post_run_command: values.post_run_command,
        build_variables: values.build_variables,
        environment_variables: values.environment_variables,
        port: parsePort(values.port?.toString() || '3000') || 3000,
        id: values.id,
        force: values.force,
        dockerfile_path: values.DockerfilePath
      };

      const data = await updateDeployment(updateData).unwrap();

      if (data?.id) {
        router.push('/self-host/application/' + data.id);
        toast.success('Deployment updated successfully');
      }
    } catch (error) {
      toast.error('Failed to update deployment');
    }
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

  return {
    validateEnvVar,
    deploymentFormSchema,
    form,
    onSubmit,
    isLoading,
    domains
  };
}

export default useUpdateDeployment;
