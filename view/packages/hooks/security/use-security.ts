'use client';

import { useCallback, useMemo, useState } from 'react';
import { toast } from 'sonner';
import { authClient } from '@/packages/lib/auth-client';
import type { Passkey } from '@/packages/hooks/security/use-passkeys';

export type { Passkey };

interface PasskeyMutations {
  addPasskey: (name?: string) => Promise<boolean>;
  deletePasskey: (id: string) => Promise<void>;
  updatePasskey: (id: string, name: string) => Promise<boolean>;
  isAdding: boolean;
  isDeleting: boolean;
  isUpdating: boolean;
}

export function usePasskeyList(
  setAddDialogOpen: (open: boolean) => void,
  mutations: PasskeyMutations,
) {
  const { addPasskey, deletePasskey, updatePasskey, isAdding, isDeleting, isUpdating } = mutations;

  const [renameTarget, setRenameTarget] = useState<Passkey | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Passkey | null>(null);
  const [newName, setNewName] = useState('');
  const [passkeyName, setPasskeyName] = useState('');

  const handleAdd = useCallback(async () => {
    const success = await addPasskey(passkeyName || undefined);
    if (success) {
      setAddDialogOpen(false);
      setPasskeyName('');
      (window as Window & { __refreshSudoPasskeys?: () => void }).__refreshSudoPasskeys?.();
    }
  }, [addPasskey, passkeyName, setAddDialogOpen]);

  const handleRename = useCallback(async () => {
    if (!renameTarget || !newName.trim()) return;
    const ok = await updatePasskey(renameTarget.id, newName.trim());
    if (ok) {
      setRenameTarget(null);
      setNewName('');
    }
  }, [renameTarget, newName, updatePasskey]);

  const handleDelete = useCallback(async () => {
    if (!deleteTarget) return;
    await deletePasskey(deleteTarget.id);
    setDeleteTarget(null);
    (window as Window & { __refreshSudoPasskeys?: () => void }).__refreshSudoPasskeys?.();
  }, [deleteTarget, deletePasskey]);

  const addActions = useMemo(
    () => [
      { label: 'Cancel', onClick: () => setAddDialogOpen(false), variant: 'outline' as const },
      {
        label: isAdding ? 'Registering...' : 'Register',
        onClick: handleAdd,
        disabled: isAdding,
        loading: isAdding,
      },
    ],
    [isAdding, handleAdd, setAddDialogOpen],
  );

  const renameActions = useMemo(
    () => [
      { label: 'Cancel', onClick: () => setRenameTarget(null), variant: 'outline' as const },
      {
        label: isUpdating ? 'Saving...' : 'Save',
        onClick: handleRename,
        disabled: isUpdating || !newName.trim(),
        loading: isUpdating,
      },
    ],
    [isUpdating, handleRename, newName],
  );

  const deleteActions = useMemo(
    () => [
      { label: 'Cancel', onClick: () => setDeleteTarget(null), variant: 'outline' as const },
      {
        label: isDeleting ? 'Deleting...' : 'Delete',
        onClick: handleDelete,
        variant: 'destructive' as const,
        disabled: isDeleting,
        loading: isDeleting,
      },
    ],
    [isDeleting, handleDelete],
  );

  return {
    renameTarget,
    setRenameTarget,
    deleteTarget,
    setDeleteTarget,
    newName,
    setNewName,
    passkeyName,
    setPasskeyName,
    handleAdd,
    handleRename,
    handleDelete,
    addActions,
    renameActions,
    deleteActions,
  };
}

export function usePasskeyVerificationDialog(onVerified: () => void, onCancel: () => void) {
  const [isVerifying, setIsVerifying] = useState(false);

  const handleVerify = useCallback(async () => {
    setIsVerifying(true);
    try {
      const res = await authClient.signIn.passkey();
      if (res?.error) {
        toast.error(res.error.message || 'Verification failed');
        return;
      }
      onVerified();
    } catch (error: unknown) {
      const msg = error instanceof Error ? error.message : 'Verification failed';
      if (!msg.includes('cancelled') && !msg.includes('abort')) {
        toast.error(msg);
      }
    } finally {
      setIsVerifying(false);
    }
  }, [onVerified]);

  const handleOpenChange = useCallback(
    (val: boolean) => {
      if (!val) onCancel();
    },
    [onCancel],
  );

  const actions = useMemo(
    () => [
      {
        label: 'Cancel',
        onClick: onCancel,
        variant: 'outline' as const,
        disabled: isVerifying,
      },
      {
        label: isVerifying ? 'Verifying...' : 'Verify with Passkey',
        onClick: handleVerify,
        disabled: isVerifying,
        loading: isVerifying,
      },
    ],
    [onCancel, isVerifying, handleVerify],
  );

  return { isVerifying, handleVerify, handleOpenChange, actions };
}
