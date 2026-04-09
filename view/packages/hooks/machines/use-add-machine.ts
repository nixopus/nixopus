'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import {
  useCreateMachineMutation,
  useVerifyMachineMutation,
  useLazyGetMachineSshStatusQuery
} from '@/redux/services/servers/serversApi';

type WizardStep = 'enter-details' | 'copy-key' | 'verify-connection';

interface MachineForm {
  name: string;
  host: string;
  port: string;
  user: string;
}

export function useAddMachine(onSuccess?: () => void) {
  const [step, setStep] = useState<WizardStep>('enter-details');
  const [form, setForm] = useState<MachineForm>({
    name: '',
    host: '',
    port: '22',
    user: 'root'
  });
  const [machineId, setMachineId] = useState<string | null>(null);
  const [publicKey, setPublicKey] = useState<string | null>(null);
  const [verificationStatus, setVerificationStatus] = useState<
    'idle' | 'polling' | 'success' | 'failed'
  >('idle');
  const [error, setError] = useState<string | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const [createMachine, { isLoading: isCreating }] = useCreateMachineMutation();
  const [verifyMachine, { isLoading: isVerifying }] = useVerifyMachineMutation();
  const [triggerSshStatus] = useLazyGetMachineSshStatusQuery();

  const updateForm = useCallback((field: keyof MachineForm, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    setError(null);
  }, []);

  const handleCreateMachine = useCallback(async () => {
    setError(null);
    try {
      const result = await createMachine({
        name: form.name,
        host: form.host,
        port: form.port ? parseInt(form.port, 10) : undefined,
        user: form.user || undefined
      }).unwrap();
      setMachineId(result.id);
      setPublicKey(result.public_key);
      setStep('copy-key');
    } catch (err: any) {
      setError(err?.data?.detail || err?.message || 'Failed to create machine');
    }
  }, [form, createMachine]);

  const handleKeyConfirmed = useCallback(async () => {
    if (!machineId) return;
    setError(null);
    try {
      await verifyMachine(machineId).unwrap();
      setVerificationStatus('polling');
      setStep('verify-connection');
    } catch (err: any) {
      setError(err?.data?.detail || err?.message || 'Failed to start verification');
    }
  }, [machineId, verifyMachine]);

  useEffect(() => {
    if (verificationStatus !== 'polling' || !machineId) return;

    let attempts = 0;
    const maxAttempts = 10;

    pollRef.current = setInterval(async () => {
      attempts++;
      try {
        const result = await triggerSshStatus(machineId).unwrap();
        if (result.is_active) {
          setVerificationStatus('success');
          if (pollRef.current) clearInterval(pollRef.current);
          onSuccess?.();
        }
      } catch {
        // continue polling
      }

      if (attempts >= maxAttempts) {
        setVerificationStatus('failed');
        if (pollRef.current) clearInterval(pollRef.current);
      }
    }, 2000);

    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, [verificationStatus, machineId, triggerSshStatus, onSuccess]);

  const handleRetryVerification = useCallback(async () => {
    if (!machineId) return;
    setVerificationStatus('idle');
    setError(null);
    try {
      await verifyMachine(machineId).unwrap();
      setVerificationStatus('polling');
    } catch (err: any) {
      setError(err?.data?.detail || err?.message || 'Failed to start verification');
    }
  }, [machineId, verifyMachine]);

  const reset = useCallback(() => {
    setStep('enter-details');
    setForm({ name: '', host: '', port: '22', user: 'root' });
    setMachineId(null);
    setPublicKey(null);
    setVerificationStatus('idle');
    setError(null);
    if (pollRef.current) clearInterval(pollRef.current);
  }, []);

  const canProceedStep1 = form.name.trim() !== '' && form.host.trim() !== '';

  return {
    step,
    form,
    machineId,
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
  };
}
