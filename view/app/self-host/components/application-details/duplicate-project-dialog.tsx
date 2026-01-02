'use client';
import React, { useState, useEffect, useCallback } from 'react';
import { Copy } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { DropdownMenuItem } from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { Application, Environment } from '@/redux/types/applications';
import {
  useDuplicateProjectMutation,
  useGetFamilyEnvironmentsQuery
} from '@/redux/services/deploy/applicationsApi';
import { useGetGithubRepositoryBranchesMutation } from '@/redux/services/connector/githubConnectorApi';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';

interface DuplicateProjectDialogProps {
  application: Application;
}

const ENVIRONMENTS: Environment[] = ['development', 'staging', 'production'];

export function DuplicateProjectDialog({ application }: DuplicateProjectDialogProps) {
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
      setDomain('');
      setEnvironment('');
      setBranch('');
      router.push(`/self-host/application/${result.id}`);
    } catch (error) {
      toast.error(t('selfHost.applicationDetails.header.duplicate.error'));
    }
  };

  const handleOpenChange = (isOpen: boolean) => {
    setOpen(isOpen);
    if (!isOpen) {
      setDomain('');
      setEnvironment('');
      setBranch('');
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      {isDisabled ? (
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
      ) : (
        <DialogTrigger asChild>
          <DropdownMenuItem onSelect={(e) => e.preventDefault()} className="gap-2">
            <Copy className="h-4 w-4" />
            {t('selfHost.applicationDetails.header.duplicate.button')}
          </DropdownMenuItem>
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>
            {t('selfHost.applicationDetails.header.duplicate.dialog.title')}
          </DialogTitle>
          <DialogDescription>
            {t('selfHost.applicationDetails.header.duplicate.dialog.description')}
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="environment">
              {t('selfHost.applicationDetails.header.duplicate.dialog.environment')}
            </Label>
            <Select
              value={environment}
              onValueChange={(value) => setEnvironment(value as Environment)}
            >
              <SelectTrigger id="environment">
                <SelectValue
                  placeholder={t(
                    'selfHost.applicationDetails.header.duplicate.dialog.selectEnvironment'
                  )}
                />
              </SelectTrigger>
              <SelectContent>
                {availableEnvironments.map((env) => (
                  <SelectItem key={env} value={env}>
                    <span className="capitalize">{env}</span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              {t('selfHost.applicationDetails.header.duplicate.dialog.environmentHint')}
            </p>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="branch">Branch</Label>
            <Select
              value={branch}
              onValueChange={(value) => setBranch(value)}
              disabled={isLoadingBranches}
            >
              <SelectTrigger id="branch">
                <SelectValue
                  placeholder={isLoadingBranches ? 'Loading branches...' : 'Select a branch'}
                />
              </SelectTrigger>
              <SelectContent>
                {availableBranches.map((branchOption) => (
                  <SelectItem key={branchOption.value} value={branchOption.value}>
                    {branchOption.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              Select the branch to use for the duplicate project
            </p>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="domain">
              {t('selfHost.applicationDetails.header.duplicate.dialog.domain')}
            </Label>
            <Input
              id="domain"
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
              placeholder="staging.example.com"
            />
            <p className="text-xs text-muted-foreground">
              {t('selfHost.applicationDetails.header.duplicate.dialog.domainHint')}
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)} disabled={isLoading}>
            {t('common.cancel')}
          </Button>
          <Button
            onClick={handleDuplicate}
            disabled={isLoading || !domain || !environment || !branch}
          >
            {isLoading
              ? t('selfHost.applicationDetails.header.duplicate.dialog.duplicating')
              : t('selfHost.applicationDetails.header.duplicate.dialog.duplicateButton')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default DuplicateProjectDialog;
