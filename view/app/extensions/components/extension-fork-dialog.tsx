'use client';

import React, { useEffect, useMemo, useState } from 'react';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { useTranslation } from '@/hooks/use-translation';
import { Extension } from '@/redux/types/extension';
import { useForkExtensionMutation } from '@/redux/services/extensions/extensionsApi';
import { toast } from 'sonner';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import AceEditor from '@/components/ui/ace-editor';
import YAML from 'yaml';

interface ExtensionForkDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  extension: Extension;
}

export default function ExtensionForkDialog({ open, onOpenChange, extension }: ExtensionForkDialogProps) {
  const { t } = useTranslation();
  const [forkYaml, setForkYaml] = useState<string>('');
  const [forkExtension, { isLoading }] = useForkExtensionMutation();

  const preview = useMemo(() => {
    try {
      const y = YAML.parse(forkYaml || '');
      return {
        variables: y?.variables || {},
        execution: y?.execution || {},
        metadata: y?.metadata || {}
      } as any;
    } catch {
      return undefined;
    }
  }, [forkYaml]);

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
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="sm:max-w-3xl md:max-w-4xl w-full p-0">
        <SheetHeader>
          <SheetTitle>{t('extensions.fork') || 'Fork Extension'}</SheetTitle>
          <SheetDescription>{extension.description}</SheetDescription>
        </SheetHeader>
        <div className="px-4 pb-4 flex-1 flex flex-col overflow-hidden">
        <Tabs defaultValue="edit" className="w-full flex-1 flex flex-col overflow-hidden">
          <TabsList>
            <TabsTrigger value="edit">{t('common.edit') || 'Edit'}</TabsTrigger>
            <TabsTrigger value="preview">{t('common.preview') || 'Preview'}</TabsTrigger>
          </TabsList>
          <TabsContent value="edit" className="mt-2 flex-1 overflow-hidden">
            <div className="flex flex-col gap-2 h-full min-h-0">
              <Label>{t('extensions.forkYaml') || 'YAML (optional)'}</Label>
              <div className="flex-1 min-h-0">
                <AceEditor
                  mode="yaml"
                  name="fork-yaml-editor"
                  value={forkYaml}
                  onChange={setForkYaml}
                  height="100%"
                />
              </div>
            </div>
          </TabsContent>
          <TabsContent value="preview" className="mt-2">
            <div className="space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="rounded-md border p-3">
                  <div className="text-xs text-muted-foreground">{t('extensions.type') || 'Type'}</div>
                  <div className="text-sm font-medium">{preview?.metadata?.type || extension.extension_type}</div>
                </div>
                <div className="rounded-md border p-3">
                  <div className="text-xs text-muted-foreground">{t('extensions.category') || 'Category'}</div>
                  <div className="text-sm font-medium">{preview?.metadata?.category || extension.category}</div>
                </div>
              </div>
              <div className="rounded-md border p-3">
                <div className="text-sm font-semibold mb-2">{t('extensions.variables') || 'Variables'}</div>
                <div className="space-y-2">
                  {(preview?.variables && Object.keys(preview.variables).length > 0) ? (
                    Object.entries(preview.variables).map(([key, val]: any) => (
                      <div key={key} className="flex items-center justify-between text-sm">
                        <div className="font-medium">{key}</div>
                        <div className="text-muted-foreground">{val?.type}</div>
                      </div>
                    ))
                  ) : (
                    <div className="text-sm text-muted-foreground">{t('extensions.noVariables') || 'No variables defined'}</div>
                  )}
                </div>
              </div>
              <div className="rounded-md border p-3">
                <div className="text-sm font-semibold mb-2">{t('extensions.execution') || 'Execution'}</div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <div className="text-xs text-muted-foreground mb-1">{t('extensions.runSteps') || 'Run steps'}</div>
                    <ul className="list-disc pl-5 space-y-1 text-sm">
                      {(preview?.execution?.run || []).map((s: any, i: number) => (
                        <li key={i}>{s?.name || s?.type || 'Step'}</li>
                      ))}
                    </ul>
                  </div>
                  <div>
                    <div className="text-xs text-muted-foreground mb-1">{t('extensions.validateSteps') || 'Validate steps'}</div>
                    <ul className="list-disc pl-5 space-y-1 text-sm">
                      {(preview?.execution?.validate || []).map((s: any, i: number) => (
                        <li key={i}>{s?.name || s?.type || 'Step'}</li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          </TabsContent>
        </Tabs>
        </div>
        <div className="flex justify-end gap-2 p-4">
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={doFork} disabled={isLoading}>{t('extensions.fork') || 'Fork'}</Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}


