import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { Copy } from 'lucide-react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Application } from '@/redux/types/applications';
import {
  useDuplicateProjectMutation,
  useGetFamilyEnvironmentsQuery
} from '@/redux/services/deploy/applicationsApi';
import { useGetGithubRepositoryBranchesMutation } from '@/redux/services/connector/githubConnectorApi';
import { DropdownMenuItem } from '@nixopus/ui';
import { DialogAction } from '@nixopus/ui';
import { formatEnvironmentName, isValidEnvironmentName } from '@/packages/utils/environment';

interface UseDuplicateProjectProps {
  application: Application;
}

export function useDuplicateProject({ application }: UseDuplicateProjectProps) {
  const { t } = useTranslation();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [domain, setDomain] = useState('');
  const [environment, setEnvironment] = useState('');
  const [branch, setBranch] = useState('');
  const [environmentError, setEnvironmentError] = useState('');
  const [availableBranches, setAvailableBranches] = useState<{ label: string; value: string }[]>(
    []
  );

  const [duplicateProject, { isLoading }] = useDuplicateProjectMutation();
  const [getGithubRepositoryBranches, { isLoading: isLoadingBranches }] =
    useGetGithubRepositoryBranchesMutation();

  const { data: existingEnvironments = [] } = useGetFamilyEnvironmentsQuery(
    { familyId: application.family_id || '' },
    { skip: !application.family_id }
  );

  const fetchRepositoryBranches = useCallback(async () => {
    if (!application.repository) {
      return;
    }

    try {
      const result = await getGithubRepositoryBranches(application.repository).unwrap();
      const branchOptions = result.map((branch) => ({
        label: branch.name,
        value: branch.name
      }));
      setAvailableBranches(branchOptions);

      // Set default branch if available
      const defaultBranch =
        branchOptions.find((b) => b.value === 'main') ||
        branchOptions.find((b) => b.value === 'master') ||
        branchOptions[0];
      if (defaultBranch) {
        setBranch(defaultBranch.value);
      }
    } catch {
      toast.error('Failed to fetch repository branches');
    }
  }, [getGithubRepositoryBranches, application.repository]);

  useEffect(() => {
    if (open && application.repository) {
      fetchRepositoryBranches();
    }
  }, [open, application.repository, fetchRepositoryBranches]);

  const handleDuplicate = async () => {
    const formatted = formatEnvironmentName(environment);
    if (!formatted || !branch) {
      toast.error(t('selfHost.applicationDetails.header.duplicate.validation'));
      return;
    }

    if (!isValidEnvironmentName(formatted)) {
      setEnvironmentError('Invalid environment name. Use lowercase letters, numbers, and hyphens.');
      return;
    }

    if (formatted === application.environment) {
      setEnvironmentError('Cannot duplicate with the same environment.');
      return;
    }

    if (existingEnvironments.includes(formatted)) {
      setEnvironmentError('This environment already exists in the project family.');
      return;
    }

    try {
      const duplicateData: any = {
        source_project_id: application.id,
        environment: formatted,
        branch
      };

      if (domain && domain.trim() !== '') {
        duplicateData.domains = [domain.trim()];
      }

      const result = await duplicateProject(duplicateData).unwrap();

      toast.success(t('selfHost.applicationDetails.header.duplicate.success'));
      setOpen(false);
      resetForm();
      router.push(`/apps/application/${result.id}`);
    } catch (error) {
      toast.error(t('selfHost.applicationDetails.header.duplicate.error'));
    }
  };

  const resetForm = () => {
    setDomain('');
    setEnvironment('');
    setBranch('');
    setEnvironmentError('');
  };

  const handleOpenChange = (isOpen: boolean) => {
    setOpen(isOpen);
    if (!isOpen) {
      resetForm();
    }
  };

  const isFormValid = environment && branch; // Domain is now optional

  const formFields = useMemo(
    () => [
      {
        id: 'environment',
        label: t('selfHost.applicationDetails.header.duplicate.dialog.environment'),
        type: 'input' as const,
        value: environment,
        onChange: (e: React.ChangeEvent<HTMLInputElement>) => {
          setEnvironmentError('');
          setEnvironment(formatEnvironmentName(e.target.value));
        },
        placeholder: 'e.g. staging, qa, preview',
        hint: environmentError || 'Lowercase letters, numbers, and hyphens only'
      },
      {
        id: 'branch',
        label: 'Branch',
        type: 'select' as const,
        value: branch,
        onChange: setBranch,
        options: availableBranches,
        placeholder: isLoadingBranches ? 'Loading branches...' : 'Select a branch',
        disabled: isLoadingBranches,
        loading: isLoadingBranches,
        hint: 'Select the branch to use for the duplicate project'
      },
      {
        id: 'domain',
        label: t('selfHost.applicationDetails.header.duplicate.dialog.domain'),
        type: 'input' as const,
        value: domain,
        onChange: (e: React.ChangeEvent<HTMLInputElement>) => setDomain(e.target.value),
        placeholder: 'staging.example.com',
        hint: t('selfHost.applicationDetails.header.duplicate.dialog.domainHint')
      }
    ],
    [
      t,
      environment,
      setEnvironment,
      environmentError,
      branch,
      setBranch,
      availableBranches,
      isLoadingBranches,
      domain,
      setDomain
    ]
  );

  const dialogActions = useMemo<DialogAction[]>(
    () => [
      {
        label: t('common.cancel'),
        onClick: () => handleOpenChange(false),
        variant: 'outline',
        disabled: isLoading
      },
      {
        label: isLoading
          ? t('selfHost.applicationDetails.header.duplicate.dialog.duplicating')
          : t('selfHost.applicationDetails.header.duplicate.dialog.duplicateButton'),
        onClick: handleDuplicate,
        disabled: isLoading || !isFormValid,
        loading: isLoading
      }
    ],
    [t, isLoading, isFormValid, handleDuplicate, handleOpenChange]
  );

  const trigger = useMemo(
    () => (
      <DropdownMenuItem onSelect={(e) => e.preventDefault()} className="gap-2">
        <Copy className="h-4 w-4" />
        {t('selfHost.applicationDetails.header.duplicate.button')}
      </DropdownMenuItem>
    ),
    [t]
  );

  const dialogTitle = t('selfHost.applicationDetails.header.duplicate.dialog.title');
  const dialogDescription = t('selfHost.applicationDetails.header.duplicate.dialog.description');

  return {
    open,
    setOpen: handleOpenChange,
    formFields,
    dialogActions,
    trigger,
    dialogTitle,
    dialogDescription,
    isLoading,
    isFormValid
  };
}
