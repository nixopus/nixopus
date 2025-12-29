'use client';
import React, { useState } from 'react';
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
import { Application, Environment } from '@/redux/types/applications';
import { useDuplicateProjectMutation } from '@/redux/services/deploy/applicationsApi';
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
  const [duplicateProject, { isLoading }] = useDuplicateProjectMutation();

  const availableEnvironments = ENVIRONMENTS.filter((env) => env !== application.environment);

  const handleDuplicate = async () => {
    if (!domain || !environment) {
      toast.error(t('selfHost.applicationDetails.header.duplicate.validation'));
      return;
    }

    try {
      const result = await duplicateProject({
        source_project_id: application.id,
        domain,
        environment
      }).unwrap();

      toast.success(t('selfHost.applicationDetails.header.duplicate.success'));
      setOpen(false);
      setDomain('');
      setEnvironment('');
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
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <DropdownMenuItem onSelect={(e) => e.preventDefault()} className="gap-2">
          <Copy className="h-4 w-4" />
          {t('selfHost.applicationDetails.header.duplicate.button')}
        </DropdownMenuItem>
      </DialogTrigger>
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
            <Label htmlFor="source-project">
              {t('selfHost.applicationDetails.header.duplicate.dialog.sourceProject')}
            </Label>
            <Input id="source-project" value={application.name} disabled className="bg-muted" />
          </div>
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
          <Button onClick={handleDuplicate} disabled={isLoading || !domain || !environment}>
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
