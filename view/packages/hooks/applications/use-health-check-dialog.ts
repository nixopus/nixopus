'use client';

import { useState, useEffect } from 'react';
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
  const [intervalSeconds, setIntervalSeconds] = useState(60);
  const [timeoutSeconds, setTimeoutSeconds] = useState(30);
  const [enabled, setEnabled] = useState(true);

  const [createHealthCheck, { isLoading: isCreating }] = useCreateHealthCheckMutation();
  const [updateHealthCheck, { isLoading: isUpdating }] = useUpdateHealthCheckMutation();
  const [deleteHealthCheck] = useDeleteHealthCheckMutation();
  const [toggleHealthCheck] = useToggleHealthCheckMutation();

  useEffect(() => {
    if (healthCheck) {
      setEndpoint(healthCheck.endpoint);
      setMethod(healthCheck.method);
      setIntervalSeconds(healthCheck.interval_seconds);
      setTimeoutSeconds(healthCheck.timeout_seconds);
      setEnabled(healthCheck.enabled);
    } else {
      setEndpoint('/');
      setMethod('GET');
      setIntervalSeconds(60);
      setTimeoutSeconds(30);
      setEnabled(true);
    }
  }, [healthCheck]);

  const handleSubmit = async () => {
    try {
      if (healthCheck) {
        await updateHealthCheck({
          application_id: application.id,
          endpoint,
          method,
          interval_seconds: intervalSeconds,
          timeout_seconds: timeoutSeconds
        }).unwrap();
        toast.success(t('selfHost.monitoring.healthCheck.updated' as any));
      } else {
        await createHealthCheck({
          application_id: application.id,
          endpoint,
          method,
          interval_seconds: intervalSeconds,
          timeout_seconds: timeoutSeconds
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

  return {
    endpoint,
    setEndpoint,
    method,
    setMethod,
    intervalSeconds,
    setIntervalSeconds,
    timeoutSeconds,
    setTimeoutSeconds,
    enabled,
    setEnabled,
    handleSubmit,
    handleDelete,
    handleToggle,
    isLoading: isCreating || isUpdating
  };
}
