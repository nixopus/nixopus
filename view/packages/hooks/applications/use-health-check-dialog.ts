'use client';

import { useState, useEffect, useMemo } from 'react';
import {
  useCreateHealthCheckMutation,
  useUpdateHealthCheckMutation,
  useDeleteHealthCheckMutation,
  useToggleHealthCheckMutation
} from '@/redux/services/deploy/healthcheckApi';
import { Application } from '@/redux/types/applications';
import { HealthCheck } from '@/redux/types/healthcheck';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { toast } from 'sonner';
import { SelectOption } from '@/components/ui/select-wrapper';
import { DialogAction } from '@/components/ui/dialog-wrapper';

interface UseHealthCheckDialogProps {
  application: Application;
  healthCheck?: HealthCheck;
  onSuccess?: () => void;
}

export function useHealthCheckDialog({
  application,
  healthCheck,
  onSuccess
}: UseHealthCheckDialogProps) {
  const { t } = useTranslation();
  const [endpoint, setEndpoint] = useState('/');
  const [method, setMethod] = useState<'GET' | 'POST' | 'HEAD'>('GET');
  const [intervalSeconds, setIntervalSeconds] = useState<string>('60');
  const [timeoutSeconds, setTimeoutSeconds] = useState<string>('30');
  const [enabled, setEnabled] = useState(true);

  const [createHealthCheck, { isLoading: isCreating }] = useCreateHealthCheckMutation();
  const [updateHealthCheck, { isLoading: isUpdating }] = useUpdateHealthCheckMutation();
  const [deleteHealthCheck] = useDeleteHealthCheckMutation();
  const [toggleHealthCheck] = useToggleHealthCheckMutation();

  useEffect(() => {
    if (healthCheck) {
      setEndpoint(healthCheck.endpoint);
      setMethod(healthCheck.method);
      setIntervalSeconds(healthCheck.interval_seconds.toString());
      setTimeoutSeconds(healthCheck.timeout_seconds.toString());
      setEnabled(healthCheck.enabled);
    } else {
      setEndpoint('/');
      setMethod('GET');
      setIntervalSeconds('60');
      setTimeoutSeconds('30');
      setEnabled(true);
    }
  }, [healthCheck]);

  const handleSubmit = async () => {
    try {
      const intervalSecondsNum = parseInt(intervalSeconds, 10) || 60;
      const timeoutSecondsNum = parseInt(timeoutSeconds, 10) || 30;

      if (healthCheck) {
        await updateHealthCheck({
          application_id: application.id,
          endpoint,
          method,
          interval_seconds: intervalSecondsNum,
          timeout_seconds: timeoutSecondsNum
        }).unwrap();
        toast.success(t('selfHost.monitoring.healthCheck.updated' as any));
      } else {
        await createHealthCheck({
          application_id: application.id,
          endpoint,
          method,
          interval_seconds: intervalSecondsNum,
          timeout_seconds: timeoutSecondsNum
        }).unwrap();
        toast.success(t('selfHost.monitoring.healthCheck.created' as any));
      }
      onSuccess?.();
    } catch (error: any) {
      const errorMessage =
        error?.data?.error ||
        error?.data?.message ||
        error?.message ||
        t('selfHost.monitoring.healthCheck.error' as any);
      toast.error(errorMessage);
    }
  };

  const handleDelete = async () => {
    if (!healthCheck) return;
    try {
      await deleteHealthCheck(application.id).unwrap();
      toast.success(t('selfHost.monitoring.healthCheck.deleted' as any));
      onSuccess?.();
    } catch (error: any) {
      const errorMessage =
        error?.data?.error ||
        error?.data?.message ||
        error?.message ||
        t('selfHost.monitoring.healthCheck.error' as any);
      toast.error(errorMessage);
    }
  };

  const handleToggle = async () => {
    if (!healthCheck) return;
    try {
      await toggleHealthCheck({
        application_id: application.id,
        enabled: !enabled
      }).unwrap();
      setEnabled(!enabled);
      toast.success(
        enabled
          ? t('selfHost.monitoring.healthCheck.disabled' as any)
          : t('selfHost.monitoring.healthCheck.enabled' as any)
      );
    } catch (error: any) {
      const errorMessage =
        error?.data?.error ||
        error?.data?.message ||
        error?.message ||
        t('selfHost.monitoring.healthCheck.error' as any);
      toast.error(errorMessage);
    }
  };

  const handleIntervalSecondsBlur = (e: React.FocusEvent<HTMLInputElement>) => {
    const inputValue = e.target.value.trim();
    if (inputValue === '') {
      setIntervalSeconds('60');
      return;
    }
    const value = parseInt(inputValue, 10);
    if (isNaN(value)) {
      setIntervalSeconds('60');
    } else if (value < 30) {
      setIntervalSeconds('30');
    } else if (value > 3600) {
      setIntervalSeconds('3600');
    }
    // If value is valid and in range, don't update state to avoid resetting user input
  };

  const handleTimeoutSecondsBlur = (e: React.FocusEvent<HTMLInputElement>) => {
    const inputValue = e.target.value.trim();
    if (inputValue === '') {
      setTimeoutSeconds('30');
      return;
    }
    const value = parseInt(inputValue, 10);
    if (isNaN(value)) {
      setTimeoutSeconds('30');
    } else if (value < 5) {
      setTimeoutSeconds('5');
    } else if (value > 120) {
      setTimeoutSeconds('120');
    }
    // If value is valid and in range, don't update state to avoid resetting user input
  };

  const methodOptions: SelectOption[] = useMemo(
    () => [
      { value: 'GET', label: 'GET' },
      { value: 'POST', label: 'POST' },
      { value: 'HEAD', label: 'HEAD' }
    ],
    []
  );

  const dialogActions: DialogAction[] = useMemo(
    () => [
      ...(healthCheck
        ? [
            {
              label: t('selfHost.monitoring.healthCheck.delete' as any) || 'Delete',
              onClick: handleDelete,
              variant: 'destructive' as const,
              disabled: isCreating || isUpdating,
              loading: false
            }
          ]
        : []),
      {
        label: healthCheck
          ? t('selfHost.monitoring.healthCheck.update' as any) || 'Update'
          : t('selfHost.monitoring.healthCheck.create' as any) || 'Create',
        onClick: handleSubmit,
        disabled: isCreating || isUpdating,
        loading: isCreating || isUpdating
      }
    ],
    [healthCheck, isCreating, isUpdating, t, handleSubmit, handleDelete]
  );

  return {
    endpoint,
    setEndpoint,
    method,
    setMethod,
    methodOptions,
    intervalSeconds,
    setIntervalSeconds,
    handleIntervalSecondsBlur,
    timeoutSeconds,
    setTimeoutSeconds,
    handleTimeoutSecondsBlur,
    enabled,
    setEnabled,
    handleSubmit,
    handleDelete,
    handleToggle,
    dialogActions,
    isLoading: isCreating || isUpdating
  };
}
