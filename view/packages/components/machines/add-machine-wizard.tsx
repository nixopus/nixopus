'use client';

import { useState } from 'react';
import { useAddMachine } from '@/packages/hooks/machines/use-add-machine';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  Button,
  Input,
  Label,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription
} from '@nixopus/ui';
import { Copy, Check, Loader2, AlertCircle, CheckCircle2 } from 'lucide-react';

interface AddMachineWizardProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AddMachineWizard({ open, onOpenChange }: AddMachineWizardProps) {
  const [copied, setCopied] = useState(false);
  const { t } = useTranslation();

  const {
    step,
    form,
    publicKey,
    verificationStatus,
    error,
    isCreating,
    isVerifying,
    canProceedStep1,
    updateForm,
    handleCreateMachine,
    handleKeyConfirmed,
    handleRetryVerification,
    reset
  } = useAddMachine(() => {});

  const handleClose = () => {
    onOpenChange(false);
    setTimeout(reset, 300);
  };

  const handleCopyKey = async () => {
    if (!publicKey) return;
    await navigator.clipboard.writeText(publicKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{t('machines.addMachine')}</DialogTitle>
          <DialogDescription>
            {step === 'copy-key' && t('machines.copyKey')}
            {step === 'verify-connection' && t('machines.verifyConnection')}
          </DialogDescription>
        </DialogHeader>

        {error && (
          <div className="flex items-center gap-2 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {step === 'enter-details' && (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="machine-name">{t('machines.nameLabel')}</Label>
              <Input
                id="machine-name"
                placeholder={t('machines.namePlaceholder')}
                value={form.name}
                onChange={(e) => updateForm('name', e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="machine-host">{t('machines.hostLabel')}</Label>
              <Input
                id="machine-host"
                placeholder={t('machines.hostPlaceholder')}
                value={form.host}
                onChange={(e) => updateForm('host', e.target.value)}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="machine-port">{t('machines.portLabel')}</Label>
                <Input
                  id="machine-port"
                  type="number"
                  placeholder="22"
                  value={form.port}
                  onChange={(e) => updateForm('port', e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="machine-user">{t('machines.userLabel')}</Label>
                <Input
                  id="machine-user"
                  placeholder="root"
                  value={form.user}
                  onChange={(e) => updateForm('user', e.target.value)}
                />
              </div>
            </div>
            <Button
              className="w-full"
              onClick={handleCreateMachine}
              disabled={!canProceedStep1 || isCreating}
            >
              {isCreating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  {t('machines.next')}
                </>
              ) : (
                t('machines.next')
              )}
            </Button>
          </div>
        )}

        {step === 'copy-key' && publicKey && (
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">{t('machines.copyKeyInstructions')}</p>
            <div className="relative">
              <pre className="max-h-32 overflow-auto rounded-md bg-muted p-3 text-xs break-all whitespace-pre-wrap">
                {publicKey}
              </pre>
              <Button
                variant="ghost"
                size="sm"
                className="absolute top-2 right-2"
                onClick={handleCopyKey}
              >
                {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
              </Button>
            </div>
            <Button className="w-full" onClick={handleKeyConfirmed} disabled={isVerifying}>
              {isVerifying ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  {t('machines.verifying')}
                </>
              ) : (
                t('machines.keyAdded')
              )}
            </Button>
          </div>
        )}

        {step === 'verify-connection' && (
          <div className="flex flex-col items-center gap-4 py-6">
            {verificationStatus === 'polling' && (
              <>
                <Loader2 className="h-10 w-10 animate-spin text-primary" />
                <p className="text-sm text-muted-foreground">{t('machines.verifying')}</p>
              </>
            )}
            {verificationStatus === 'success' && (
              <>
                <CheckCircle2 className="h-10 w-10 text-green-500" />
                <p className="text-sm font-medium">{t('machines.verificationSuccess')}</p>
                <Button onClick={handleClose}>{t('machines.done')}</Button>
              </>
            )}
            {verificationStatus === 'failed' && (
              <>
                <AlertCircle className="h-10 w-10 text-destructive" />
                <p className="text-sm text-destructive">{t('machines.verificationFailed')}</p>
                <div className="flex gap-2">
                  <Button variant="outline" onClick={handleRetryVerification}>
                    {t('machines.retry')}
                  </Button>
                  <Button variant="ghost" onClick={handleClose}>
                    {t('machines.close')}
                  </Button>
                </div>
              </>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
