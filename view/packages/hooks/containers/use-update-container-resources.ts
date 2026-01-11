'use client';

import { useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { toast } from 'sonner';
import { useUpdateContainerResourcesMutation } from '@/redux/services/container/containerApi';
import { useTranslation, translationKey } from '@/packages/hooks/shared/use-translation';
import { Cpu, MemoryStick, HardDrive, LucideIcon } from 'lucide-react';

interface UseUpdateContainerResourcesProps {
  containerId: string;
  currentMemory: number;
  currentMemorySwap: number;
  currentCpuShares: number;
  onSuccess?: () => void;
}

export const bytesToMB = (bytes: number): number => {
  if (bytes === -1) return -1;
  if (bytes <= 0) return 0;
  return Math.round(bytes / (1024 * 1024));
};

export const mbToBytes = (mb: number): number => {
  if (mb === -1) return -1;
  if (mb <= 0) return 0;
  return mb * 1024 * 1024;
};

export interface ResourceLimitsFormValues {
  memoryMB: number;
  memorySwapMB: number;
  cpuShares: number;
}

export type PresetType = 'small' | 'medium' | 'large' | 'xlarge' | 'unlimited';
export type FieldName = 'memoryMB' | 'memorySwapMB' | 'cpuShares';

export const presetConfig: { key: PresetType; memory: number }[] = [
  { key: 'small', memory: 256 },
  { key: 'medium', memory: 512 },
  { key: 'large', memory: 1024 },
  { key: 'xlarge', memory: 2048 },
  { key: 'unlimited', memory: 0 }
];

export interface FieldConfig {
  name: FieldName;
  icon: LucideIcon;
  labelKey: translationKey;
  placeholderKey: translationKey;
  unitKey?: translationKey;
  descriptionKey: translationKey;
  unlimitedDescKey?: translationKey;
  min: number;
  isUnlimited: (value: number) => boolean;
}

export const fieldConfigs: FieldConfig[] = [
  {
    name: 'memoryMB',
    icon: MemoryStick,
    labelKey: 'containers.resourceLimits.memory.label' as translationKey,
    placeholderKey: 'containers.resourceLimits.memory.placeholder' as translationKey,
    unitKey: 'containers.resourceLimits.memory.unit' as translationKey,
    descriptionKey: 'containers.resourceLimits.memory.description' as translationKey,
    unlimitedDescKey: 'containers.resourceLimits.memory.unlimited' as translationKey,
    min: 0,
    isUnlimited: (value) => value === 0
  },
  {
    name: 'memorySwapMB',
    icon: HardDrive,
    labelKey: 'containers.resourceLimits.memorySwap.label' as translationKey,
    placeholderKey: 'containers.resourceLimits.memorySwap.placeholder' as translationKey,
    unitKey: 'containers.resourceLimits.memorySwap.unit' as translationKey,
    descriptionKey: 'containers.resourceLimits.memorySwap.description' as translationKey,
    unlimitedDescKey: 'containers.resourceLimits.memorySwap.unlimited' as translationKey,
    min: -1,
    isUnlimited: (value) => value === 0 || value === -1
  },
  {
    name: 'cpuShares',
    icon: Cpu,
    labelKey: 'containers.resourceLimits.cpuShares.label' as translationKey,
    placeholderKey: 'containers.resourceLimits.cpuShares.placeholder' as translationKey,
    descriptionKey: 'containers.resourceLimits.cpuShares.description' as translationKey,
    min: 0,
    isUnlimited: () => false
  }
];

export const formatPresetValue = (key: PresetType, memory: number): string => {
  if (key === 'unlimited') return '∞';
  return memory >= 1024 ? `${memory / 1024}G` : `${memory}M`;
};

export function useUpdateContainerResources({
  containerId,
  currentMemory,
  currentMemorySwap,
  currentCpuShares,
  onSuccess
}: UseUpdateContainerResourcesProps) {
  const { t } = useTranslation();
  const [updateResources, { isLoading }] = useUpdateContainerResourcesMutation();

  const resourceLimitsSchema = z.object({
    memoryMB: z
      .number()
      .min(0, { message: t('containers.resourceLimits.validation.memoryMin' as translationKey) })
      .refine((val) => val === 0 || val >= 6, {
        message: t('containers.resourceLimits.validation.memoryMin' as translationKey)
      }),
    memorySwapMB: z.number().refine((val) => val === -1 || val === 0 || val >= 0, {
      message: t('containers.resourceLimits.validation.swapMin' as translationKey)
    }),
    cpuShares: z
      .number()
      .min(0, { message: t('containers.resourceLimits.validation.cpuSharesMin' as translationKey) })
  });

  const form = useForm<ResourceLimitsFormValues>({
    resolver: zodResolver(resourceLimitsSchema),
    defaultValues: {
      memoryMB: bytesToMB(currentMemory),
      memorySwapMB: bytesToMB(currentMemorySwap),
      cpuShares: currentCpuShares || 0
    }
  });

  const resetToCurrentValues = useCallback(() => {
    form.reset({
      memoryMB: bytesToMB(currentMemory),
      memorySwapMB: bytesToMB(currentMemorySwap),
      cpuShares: currentCpuShares || 0
    });
  }, [form, currentMemory, currentMemorySwap, currentCpuShares]);

  const applyPreset = useCallback(
    (preset: PresetType) => {
      const presets = {
        small: { memoryMB: 256, memorySwapMB: 512, cpuShares: 512 },
        medium: { memoryMB: 512, memorySwapMB: 1024, cpuShares: 1024 },
        large: { memoryMB: 1024, memorySwapMB: 2048, cpuShares: 1024 },
        xlarge: { memoryMB: 2048, memorySwapMB: 4096, cpuShares: 2048 },
        unlimited: { memoryMB: 0, memorySwapMB: 0, cpuShares: 0 }
      };

      const presetValues = presets[preset];
      form.setValue('memoryMB', presetValues.memoryMB, { shouldDirty: true });
      form.setValue('memorySwapMB', presetValues.memorySwapMB, { shouldDirty: true });
      form.setValue('cpuShares', presetValues.cpuShares, { shouldDirty: true });
    },
    [form]
  );

  const onSubmit = useCallback(
    async (values: ResourceLimitsFormValues) => {
      try {
        const result = await updateResources({
          containerId,
          memory: mbToBytes(values.memoryMB),
          memory_swap: mbToBytes(values.memorySwapMB),
          cpu_shares: values.cpuShares
        }).unwrap();

        if (result.warnings && result.warnings.length > 0) {
          toast.warning(t('containers.resourceLimits.warnings' as translationKey), {
            description: result.warnings.join(', ')
          });
        } else {
          toast.success(t('containers.resourceLimits.success' as translationKey));
        }

        onSuccess?.();
      } catch (error) {
        toast.error(t('containers.resourceLimits.error' as translationKey));
      }
    },
    [containerId, updateResources, t, onSuccess]
  );

  return {
    form,
    onSubmit,
    isLoading,
    resetToCurrentValues,
    applyPreset
  };
}
