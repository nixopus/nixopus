'use client';

import React, { useEffect, useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useTranslation } from '@/hooks/use-translation';
import { Extension } from '@/redux/types/extension';
import { useForkExtensionMutation } from '@/redux/services/extensions/extensionsApi';
import { toast } from 'sonner';

interface ExtensionForkDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  extension: Extension;
}

export default function ExtensionForkDialog({ open, onOpenChange, extension }: ExtensionForkDialogProps) {
  const { t } = useTranslation();
  const [forkYaml, setForkYaml] = useState<string>('');
  const [forkExtension, { isLoading }] = useForkExtensionMutation();

  useEffect(() => {
    if (open) {
      setForkYaml(extension.yaml_content || '');
    }
  }, [open, extension]);

  const doFork = async () => {
    try {
      await forkExtension({ extensionId: extension.extension_id, yaml_content: forkYaml || undefined }).unwrap();
      toast.success(t('extensions.forkSuccess'));
      onOpenChange(false);
    } catch (e) {
      toast.error(t('extensions.forkFailed'));
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>{t('extensions.fork') || 'Fork Extension'}</DialogTitle>
          <DialogDescription>{extension.description}</DialogDescription>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <div className="grid gap-2">
            <Label>{t('extensions.forkYaml') || 'YAML (optional)'}</Label>
            <textarea
              className="min-h-[160px] rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              value={forkYaml}
              onChange={(e) => setForkYaml(e.target.value)}
              placeholder={t('extensions.forkYaml') || 'Paste YAML to fully customize'}
            />
          </div>
        </div>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={doFork} disabled={isLoading}>{t('extensions.fork') || 'Fork'}</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}


