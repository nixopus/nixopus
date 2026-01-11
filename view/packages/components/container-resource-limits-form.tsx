'use client';

import { useState } from 'react';
import { Settings2, Zap, Info } from 'lucide-react';
import { UseFormReturn, ControllerRenderProps } from 'react-hook-form';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog';
import { Form, FormControl, FormField, FormItem, FormLabel } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { Container } from '@/redux/services/container/containerApi';
import { useTranslation } from '@/hooks/use-translation';
import {
  useUpdateContainerResources,
  presetConfig,
  fieldConfigs,
  formatPresetValue,
  PresetType,
  FieldConfig,
  ResourceLimitsFormValues
} from '../hooks/containers/use-update-container-resources';
import { cn } from '@/lib/utils';
import { ResourceGuard } from '@/packages/components/rbac';
import { Skeleton } from '@/components/ui/skeleton';

interface ResourceLimitsFormProps {
  container: Container;
}

interface PresetButtonProps {
  presetKey: PresetType;
  memory: number;
  isActive: boolean;
  onSelect: (key: PresetType) => void;
}

function PresetButton({ presetKey, memory, isActive, onSelect }: PresetButtonProps) {
  return (
    <button
      type="button"
      onClick={() => onSelect(presetKey)}
      className={cn(
        'flex flex-col items-center gap-1 p-3 rounded-lg border-2 transition-all text-xs',
        isActive
          ? 'border-primary bg-primary/5 text-primary'
          : 'border-muted hover:border-muted-foreground/20 hover:bg-muted/50'
      )}
    >
      <span className="font-medium">{formatPresetValue(presetKey, memory)}</span>
      <span className={cn('capitalize', isActive ? 'text-primary/70' : 'text-muted-foreground')}>
        {presetKey}
      </span>
    </button>
  );
}

interface PresetGridProps {
  currentMemory: number;
  onPresetSelect: (key: PresetType) => void;
}

function PresetGrid({ currentMemory, onPresetSelect }: PresetGridProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-3">
      <label className="text-sm font-medium">{t('containers.resourceLimits.presets.label')}</label>
      <div className="grid grid-cols-5 gap-2">
        {presetConfig.map(({ key, memory }) => (
          <PresetButton
            key={key}
            presetKey={key}
            memory={memory}
            isActive={currentMemory === memory}
            onSelect={onPresetSelect}
          />
        ))}
      </div>
    </div>
  );
}

interface ResourceFieldProps {
  config: FieldConfig;
  field: ControllerRenderProps<ResourceLimitsFormValues, FieldConfig['name']>;
}

function ResourceField({ config, field }: ResourceFieldProps) {
  const { t } = useTranslation();
  const {
    icon: Icon,
    labelKey,
    placeholderKey,
    unitKey,
    descriptionKey,
    unlimitedDescKey,
    min,
    isUnlimited
  } = config;
  const description =
    isUnlimited(field.value) && unlimitedDescKey ? t(unlimitedDescKey) : t(descriptionKey);

  return (
    <FormItem>
      <FormLabel className="flex items-center gap-2">
        <Icon className="h-4 w-4 text-muted-foreground" />
        {t(labelKey)}
        <Tooltip>
          <TooltipTrigger asChild>
            <Info className="h-3.5 w-3.5 text-muted-foreground cursor-help" />
          </TooltipTrigger>
          <TooltipContent side="top" className="max-w-xs bg-popover text-popover-foreground border">
            {description}
          </TooltipContent>
        </Tooltip>
      </FormLabel>
      <div className={cn(unitKey && 'flex gap-2')}>
        <FormControl>
          <Input
            type="number"
            min={min}
            placeholder={t(placeholderKey)}
            {...field}
            onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
            className={cn(unitKey && 'flex-1')}
          />
        </FormControl>
        {unitKey && (
          <span className="flex items-center px-3 bg-muted rounded-md text-sm text-muted-foreground">
            {t(unitKey)}
          </span>
        )}
      </div>
    </FormItem>
  );
}

interface FormActionsProps {
  isLoading: boolean;
  isDirty: boolean;
  onReset: () => void;
  onCancel: () => void;
}

function FormActions({ isLoading, isDirty, onReset, onCancel }: FormActionsProps) {
  const { t } = useTranslation();

  return (
    <div className="flex justify-between pt-4">
      {isDirty ? (
        <Button type="button" variant="ghost" onClick={onReset} disabled={isLoading}>
          {t('containers.resourceLimits.buttons.reset')}
        </Button>
      ) : (
        <div />
      )}
      <div className="flex gap-2">
        <Button type="button" variant="outline" onClick={onCancel} disabled={isLoading}>
          {t('containers.resourceLimits.buttons.cancel')}
        </Button>
        <Button type="submit" disabled={isLoading}>
          {isLoading
            ? t('containers.resourceLimits.buttons.saving')
            : t('containers.resourceLimits.buttons.save')}
        </Button>
      </div>
    </div>
  );
}

interface ResourceFieldsProps {
  form: UseFormReturn<ResourceLimitsFormValues>;
}

function ResourceFields({ form }: ResourceFieldsProps) {
  return (
    <>
      {fieldConfigs.map((config) => (
        <FormField
          key={config.name}
          control={form.control}
          name={config.name}
          render={({ field }) => <ResourceField config={config} field={field} />}
        />
      ))}
    </>
  );
}

export function ResourceLimitsForm({ container }: ResourceLimitsFormProps) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  const { form, onSubmit, isLoading, resetToCurrentValues, applyPreset } =
    useUpdateContainerResources({
      containerId: container.id,
      currentMemory: container.host_config.memory,
      currentMemorySwap: container.host_config.memory_swap,
      currentCpuShares: container.host_config.cpu_shares,
      onSuccess: () => setOpen(false)
    });

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen);
    if (newOpen) resetToCurrentValues();
  };

  return (
    <ResourceGuard
      resource="container"
      action="update"
      loadingFallback={<Skeleton className="h-9 w-28" />}
    >
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogTrigger asChild>
          <Button variant="outline" size="sm" className="gap-2">
            <Settings2 className="h-4 w-4" />
            {t('containers.resourceLimits.editButton')}
          </Button>
        </DialogTrigger>

        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Zap className="h-5 w-5 text-primary" />
              {t('containers.resourceLimits.title')}
            </DialogTitle>
            <DialogDescription>{t('containers.resourceLimits.description')}</DialogDescription>
          </DialogHeader>

          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <PresetGrid currentMemory={form.watch('memoryMB')} onPresetSelect={applyPreset} />
              <ResourceFields form={form} />
              <FormActions
                isLoading={isLoading}
                isDirty={form.formState.isDirty}
                onReset={resetToCurrentValues}
                onCancel={() => setOpen(false)}
              />
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </ResourceGuard>
  );
}
