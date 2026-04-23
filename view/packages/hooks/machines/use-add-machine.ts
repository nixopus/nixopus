'use client';

import { useState, useCallback } from 'react';
import {
  useCreateMachineMutation,
  useVerifyMachineMutation
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

  const [createMachine, { isLoading: isCreating }] = useCreateMachineMutation();
  const [verifyMachine, { isLoading: isVerifying }] = useVerifyMachineMutation();

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
    setVerificationStatus('polling');
    setStep('verify-connection');
    try {
      const result = await verifyMachine(machineId).unwrap();
      if (result.is_active) {
        setVerificationStatus('success');
        onSuccess?.();
      } else {
        setVerificationStatus('failed');
      }
    } catch (err: any) {
      setVerificationStatus('failed');
      setError(err?.data?.detail || err?.message || 'Verification failed');
    }
  }, [machineId, verifyMachine, onSuccess]);

  const handleRetryVerification = useCallback(async () => {
    if (!machineId) return;
    setVerificationStatus('polling');
    setError(null);
    try {
      const result = await verifyMachine(machineId).unwrap();
      if (result.is_active) {
        setVerificationStatus('success');
        onSuccess?.();
      } else {
        setVerificationStatus('failed');
      }
    } catch (err: any) {
      setVerificationStatus('failed');
      setError(err?.data?.detail || err?.message || 'Verification failed');
    }
  }, [machineId, verifyMachine, onSuccess]);

  const reset = useCallback(() => {
    setStep('enter-details');
    setForm({ name: '', host: '', port: '22', user: 'root' });
    setMachineId(null);
    setPublicKey(null);
    setVerificationStatus('idle');
    setError(null);
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
