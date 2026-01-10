'use client';

import React, { useEffect, useMemo, useState } from 'react';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription
} from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { useTranslation } from '@/hooks/use-translation';
import { Extension } from '@/redux/types/extension';
import { useForkExtensionMutation } from '@/redux/services/extensions/extensionsApi';
import { toast } from 'sonner';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import AceEditor from '@/components/ui/ace-editor';
import YAML from 'yaml';
import { StepsSection } from '@/app/extensions/[id]/components/OverviewTab';

interface ExtensionForkDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  extension: Extension;
}

export default function ExtensionForkDialog({
  open,
  onOpenChange,
  extension
}: ExtensionForkDialogProps) {
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
      await forkExtension({
        extensionId: extension.extension_id,
        yaml_content: forkYaml || undefined
      }).unwrap();
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
                    <div className="text-xs text-muted-foreground">
                      {t('extensions.type') || 'Type'}
                    </div>
                    <div className="text-sm font-medium">
                      {preview?.metadata?.type || extension.extension_type}
                    </div>
                  </div>
                  <div className="rounded-md border p-3">
                    <div className="text-xs text-muted-foreground">
                      {t('extensions.category') || 'Category'}
                    </div>
                    <div className="text-sm font-medium">
                      {preview?.metadata?.category || extension.category}
                    </div>
                  </div>
                </div>
                {preview?.variables && Object.keys(preview.variables).length > 0 && (
                  <div className="rounded-md border overflow-hidden">
                    <div className="grid grid-cols-12 bg-muted/50 px-3 py-2 text-xs font-medium text-muted-foreground">
                      <div className="col-span-3">Name</div>
                      <div className="col-span-2">Type</div>
                      <div className="col-span-2">Required</div>
                      <div className="col-span-2">Default</div>
                      <div className="col-span-3">Description</div>
                    </div>
                    <div className="divide-y">
                      {Object.entries(preview.variables).map(([key, val]: any) => (
                        <div key={key} className="grid grid-cols-12 px-3 py-3 text-sm">
                          <div className="col-span-3 font-medium">{key}</div>
                          <div className="col-span-2 text-muted-foreground">
                            {val?.variable_type || val?.type}
                          </div>
                          <div className="col-span-2 text-muted-foreground">
                            {val?.is_required ? 'Yes' : 'No'}
                          </div>
                          <div className="col-span-2 text-muted-foreground truncate">
                            {String(val?.default_value ?? '')}
                          </div>
                          <div className="col-span-3 text-muted-foreground">{val?.description}</div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
                {(preview?.execution?.run?.length || preview?.execution?.validate?.length) && (
                  <StepsSection
                    tRunLabel={t('extensions.runSteps') || 'Run steps'}
                    tValidateLabel={t('extensions.validateSteps') || 'Validate steps'}
                    title={t('extensions.execution') || 'Execution'}
                    runSteps={preview?.execution?.run || []}
                    validateSteps={preview?.execution?.validate || []}
                    openRunIndex={null}
                    setOpenRunIndex={() => {}}
                    openValidateIndex={null}
                    setOpenValidateIndex={() => {}}
                  />
                )}
              </div>
            </TabsContent>
          </Tabs>
        </div>
        <div className="flex justify-end gap-2 p-4">
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={doFork} disabled={isLoading}>
            {t('extensions.fork') || 'Fork'}
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}
