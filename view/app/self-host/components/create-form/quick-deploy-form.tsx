'use client';
import React, { useState, useCallback, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import FormInputField from '@/components/ui/form-input-field';
import FormSelectField from '@/components/ui/form-select-field';
import { BuildPack } from '@/redux/types/deploy-form';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { useGetGithubRepositoryBranchesMutation } from '@/redux/services/connector/githubConnectorApi';
import { useCreateProjectMutation } from '@/redux/services/deploy/applicationsApi';
import { useRouter } from 'next/navigation';
import { z } from 'zod';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Github, Plus } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface QuickDeployFormProps {
  repository?: string;
  repository_full_name?: string;
  application_name?: string;
}

export const QuickDeployForm = ({
  repository,
  repository_full_name,
  application_name = ''
}: QuickDeployFormProps) => {
  const { t } = useTranslation();
  const router = useRouter();

  const [getGithubRepositoryBranches, { isLoading: isLoadingBranches }] =
    useGetGithubRepositoryBranchesMutation();
  const [createProject, { isLoading: isCreatingProject }] = useCreateProjectMutation();

  const [availableBranches, setAvailableBranches] = useState<{ label: string; value: string }[]>(
    []
  );

  const quickDeploySchema = z.object({
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
    branch: z.string().min(1, { message: t('selfHost.deployForm.validation.branch.minLength') }),
    build_pack: z.enum(['dockerfile', 'docker-compose']),
    repository: z.string()
  });

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

  return (
    <ResourceGuard
      resource="deploy"
      action="create"
      loadingFallback={<Skeleton className="h-96" />}
    >
      <Card className="w-full max-w-2xl mx-auto border-0 shadow-none bg-transparent">
        <CardHeader className="text-center pb-2">
          <div className="flex items-center justify-center gap-2 mb-2">
            <Github className="h-6 w-6 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">{repository_full_name}</span>
          </div>
          <CardTitle className="text-2xl font-bold">{application_name || 'New Project'}</CardTitle>
          <CardDescription>{t('selfHost.quickDeploy.description')}</CardDescription>
        </CardHeader>

        <CardContent className="pt-6">
          <Form {...form}>
            <form className="space-y-6">
              <div className="grid sm:grid-cols-2 gap-4">
                <FormInputField
                  form={form}
                  label={t('selfHost.quickDeploy.fields.appName.label')}
                  name="application_name"
                  placeholder={t('selfHost.quickDeploy.fields.appName.placeholder')}
                />

                <FormInputField
                  form={form}
                  label={t('selfHost.quickDeploy.fields.domain.label')}
                  name="domain"
                  placeholder={t('selfHost.quickDeploy.fields.domain.placeholder')}
                  required={true}
                />

                {isLoadingBranches ? (
                  <div className="space-y-2">
                    <div className="flex gap-2">
                      <label className="text-sm font-medium">
                        {t('selfHost.quickDeploy.fields.branch.label')}
                      </label>
                      <span className="text-destructive">*</span>
                    </div>
                    <Skeleton className="h-10 w-full" />
                  </div>
                ) : (
                  <FormSelectField
                    form={form}
                    label={t('selfHost.quickDeploy.fields.branch.label')}
                    name="branch"
                    placeholder={
                      availableBranches.length === 0
                        ? 'No branches available'
                        : t('selfHost.quickDeploy.fields.branch.placeholder')
                    }
                    selectOptions={availableBranches}
                    required={true}
                  />
                )}

                <FormSelectField
                  form={form}
                  label={t('selfHost.quickDeploy.fields.buildPack.label')}
                  name="build_pack"
                  placeholder={t('selfHost.quickDeploy.fields.buildPack.placeholder')}
                  selectOptions={[
                    { label: 'Dockerfile', value: 'dockerfile' },
                    { label: 'Docker Compose', value: 'docker-compose' }
                  ]}
                />
              </div>

              <Button
                type="button"
                onClick={handleCreate}
                disabled={isCreatingProject}
                className="w-full gap-2"
              >
                <Plus className="h-4 w-4" />
                {isCreatingProject
                  ? t('selfHost.quickDeploy.actions.creating')
                  : t('selfHost.quickDeploy.actions.create')}
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>
    </ResourceGuard>
  );
};
