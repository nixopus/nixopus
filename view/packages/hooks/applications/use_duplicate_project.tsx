import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { Copy } from 'lucide-react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Application, Environment } from '@/redux/types/applications';
import {
  useDuplicateProjectMutation,
  useGetFamilyEnvironmentsQuery
} from '@/redux/services/deploy/applicationsApi';
import { useGetGithubRepositoryBranchesMutation } from '@/redux/services/connector/githubConnectorApi';
import { DropdownMenuItem } from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { DialogAction } from '@/components/ui/dialog-wrapper';

const ENVIRONMENTS: Environment[] = ['development', 'staging', 'production'];

interface UseDuplicateProjectProps {
  application: Application;
}

export function useDuplicateProject({ application }: UseDuplicateProjectProps) {
  const { t } = useTranslation();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [domain, setDomain] = useState('');
  const [environment, setEnvironment] = useState<Environment | ''>('');
  const [branch, setBranch] = useState('');
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

  const availableEnvironments = ENVIRONMENTS.filter(
    (env) => env !== application.environment && !existingEnvironments.includes(env)
  );

  const isDisabled = availableEnvironments.length === 0;

  const environmentOptions = availableEnvironments.map((env) => ({
    value: env,
    label: env.charAt(0).toUpperCase() + env.slice(1)
  }));

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
    if (!domain || !environment || !branch) {
      toast.error(t('selfHost.applicationDetails.header.duplicate.validation'));
      return;
    }

    try {
      const result = await duplicateProject({
        source_project_id: application.id,
        domain,
        environment,
        branch
      }).unwrap();

      toast.success(t('selfHost.applicationDetails.header.duplicate.success'));
      setOpen(false);
      resetForm();
      router.push(`/self-host/application/${result.id}`);
    } catch (error) {
      toast.error(t('selfHost.applicationDetails.header.duplicate.error'));
    }
  };

  const resetForm = () => {
    setDomain('');
    setEnvironment('');
    setBranch('');
  };

  const handleOpenChange = (isOpen: boolean) => {
    setOpen(isOpen);
    if (!isOpen) {
      resetForm();
    }
  };

  const isFormValid = domain && environment && branch;

  const formFields = useMemo(
    () => [
      {
        id: 'environment',
        label: t('selfHost.applicationDetails.header.duplicate.dialog.environment'),
        type: 'select' as const,
        value: environment,
        onChange: (value: string) => setEnvironment(value as Environment),
        options: environmentOptions,
        placeholder: t('selfHost.applicationDetails.header.duplicate.dialog.selectEnvironment'),
        hint: t('selfHost.applicationDetails.header.duplicate.dialog.environmentHint')
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
      environmentOptions,
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

  const trigger = useMemo(() => {
    if (isDisabled) {
      return (
        <Tooltip>
          <TooltipTrigger asChild>
            <div>
              <DropdownMenuItem disabled className="gap-2">
                <Copy className="h-4 w-4" />
                {t('selfHost.applicationDetails.header.duplicate.button')}
              </DropdownMenuItem>
            </div>
          </TooltipTrigger>
          <TooltipContent>
            All available environments have been created. Environment creation limit reached.
          </TooltipContent>
        </Tooltip>
      );
    }

    return (
      <DropdownMenuItem onSelect={(e) => e.preventDefault()} className="gap-2">
        <Copy className="h-4 w-4" />
        {t('selfHost.applicationDetails.header.duplicate.button')}
      </DropdownMenuItem>
    );
  }, [isDisabled, t]);

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
    isDisabled,
    isFormValid
  };
}
