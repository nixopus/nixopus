import { useState, useCallback, useEffect, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { BuildPack } from '@/redux/types/deploy-form';
import { useGetGithubRepositoryBranchesMutation } from '@/redux/services/connector/githubConnectorApi';
import { useCreateProjectMutation } from '@/redux/services/deploy/applicationsApi';
import { z } from 'zod';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Github } from 'lucide-react';

interface UseQuickDeployFormProps {
  repository?: string;
  repository_full_name?: string;
  application_name?: string;
}

export function useQuickDeployForm({
  repository,
  repository_full_name,
  application_name = ''
}: UseQuickDeployFormProps) {
  const { t } = useTranslation();
  const router = useRouter();

  const [getGithubRepositoryBranches, { isLoading: isLoadingBranches }] =
    useGetGithubRepositoryBranchesMutation();
  const [createProject, { isLoading: isCreatingProject }] = useCreateProjectMutation();

  const [availableBranches, setAvailableBranches] = useState<{ label: string; value: string }[]>(
    []
  );

  const quickDeploySchema = useMemo(
    () =>
      z.object({
        application_name: z
          .string()
          .min(3, { message: t('selfHost.deployForm.validation.applicationName.minLength') })
          .regex(/^[a-zA-Z0-9_-]+$/, {
            message: t('selfHost.deployForm.validation.applicationName.invalidFormat')
          }),
        domain: z
          .string()
          .min(3, { message: t('selfHost.deployForm.validation.domain.minLength') })
          .regex(
            /^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9](\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9])*$/,
            {
              message: t('selfHost.deployForm.validation.domain.invalidFormat')
            }
          ),
        branch: z
          .string()
          .min(1, { message: t('selfHost.deployForm.validation.branch.minLength') }),
        build_pack: z.enum(['dockerfile', 'docker-compose']),
        repository: z.string()
      }),
    [t]
  );

  const form = useForm<z.infer<typeof quickDeploySchema>>({
    resolver: zodResolver(quickDeploySchema),
    defaultValues: {
      application_name: application_name,
      domain: '',
      branch: 'main',
      build_pack: 'dockerfile',
      repository: repository || ''
    }
  });

  const fetchRepositoryBranches = useCallback(async () => {
    if (!repository_full_name) {
      return;
    }

    try {
      const result = await getGithubRepositoryBranches(repository_full_name).unwrap();
      const branchOptions = result.map((branch) => ({
        label: branch.name,
        value: branch.name
      }));
      setAvailableBranches(branchOptions);

      const current = form.getValues('branch');
      const defaultBranch =
        branchOptions.find((b) => b.value === 'main') ||
        branchOptions.find((b) => b.value === 'master') ||
        branchOptions[0];
      if (!current || !branchOptions.some((b) => b.value === current)) {
        if (defaultBranch) {
          form.setValue('branch', defaultBranch.value);
        }
      }
    } catch {
      toast.error('Failed to fetch repository branches');
    }
  }, [getGithubRepositoryBranches, form, repository_full_name]);

  useEffect(() => {
    if (repository_full_name) {
      fetchRepositoryBranches();
    }
  }, [repository_full_name, fetchRepositoryBranches]);

  useEffect(() => {
    if (application_name) {
      form.setValue('application_name', application_name);
    }
    if (repository) {
      form.setValue('repository', repository);
    }
  }, [application_name, repository, form]);

  const handleCreate = async () => {
    const isValid = await form.trigger();
    if (!isValid) {
      toast.warning('Please fix the errors before saving');
      return;
    }

    const values = form.getValues();
    try {
      const result = await createProject({
        name: values.application_name,
        domain: values.domain,
        repository: values.repository,
        branch: values.branch,
        build_pack: values.build_pack as BuildPack
      }).unwrap();

      toast.success(t('selfHost.quickDeploy.toast.draftSaved'));
      router.push('/self-host/application/' + result.id);
    } catch {
      toast.error(t('selfHost.quickDeploy.toast.saveFailed'));
    }
  };

  const formFields = useMemo(
    () => [
      {
        key: 'application_name',
        type: 'input' as const,
        label: t('selfHost.quickDeploy.fields.appName.label'),
        name: 'application_name',
        placeholder: t('selfHost.quickDeploy.fields.appName.placeholder'),
        required: true
      },
      {
        key: 'domain',
        type: 'input' as const,
        label: t('selfHost.quickDeploy.fields.domain.label'),
        name: 'domain',
        placeholder: t('selfHost.quickDeploy.fields.domain.placeholder'),
        required: true
      },
      {
        key: 'branch',
        type: 'select' as const,
        label: t('selfHost.quickDeploy.fields.branch.label'),
        name: 'branch',
        placeholder:
          availableBranches.length === 0
            ? 'No branches available'
            : t('selfHost.quickDeploy.fields.branch.placeholder'),
        selectOptions: availableBranches,
        required: true,
        isLoading: isLoadingBranches
      },
      {
        key: 'build_pack',
        type: 'select' as const,
        label: t('selfHost.quickDeploy.fields.buildPack.label'),
        name: 'build_pack',
        placeholder: t('selfHost.quickDeploy.fields.buildPack.placeholder'),
        selectOptions: [
          { label: 'Dockerfile', value: 'dockerfile' },
          { label: 'Docker Compose', value: 'docker-compose' }
        ],
        required: false
      }
    ],
    [t, availableBranches, isLoadingBranches]
  );

  const headerContent = useMemo(
    () => (
      <div className="text-center pb-2 w-full ">
        <div className="flex items-center justify-center gap-2 mb-2">
          <Github className="h-6 w-6 text-muted-foreground" />
          <span className="text-sm text-muted-foreground">{repository_full_name}</span>
        </div>
        <h2 className="text-2xl font-bold">{application_name || 'New Project'}</h2>
        <p className="text-sm text-muted-foreground mt-1">
          {t('selfHost.quickDeploy.description')}
        </p>
      </div>
    ),
    [repository_full_name, application_name, t]
  );

  const buttonLabel = useMemo(
    () =>
      isCreatingProject
        ? t('selfHost.quickDeploy.actions.creating')
        : t('selfHost.quickDeploy.actions.create'),
    [isCreatingProject, t]
  );

  return {
    form,
    formFields,
    headerContent,
    handleCreate,
    buttonLabel,
    isLoadingBranches,
    isCreatingProject
  };
}
