'use client';

import { useCallback, useMemo, useState, useEffect } from 'react';
import type { FormEvent, ChangeEvent } from 'react';
import { formatDistanceToNow } from 'date-fns';
import { toast } from 'sonner';
import type { ApiKey } from '@/redux/services/api-keys/apiKeysApi';

export type { ApiKey };

export function getExpiryStatus(expiresAt: string | null) {
  if (!expiresAt) return { label: 'Never expires', variant: 'secondary' as const };
  const expiry = new Date(expiresAt);
  if (expiry < new Date()) return { label: 'Expired', variant: 'destructive' as const };
  return {
    label: `Expires ${formatDistanceToNow(expiry, { addSuffix: true })}`,
    variant: 'outline' as const,
  };
}

export function useApiKeyCreatedDialog(apiKey: string | null, onClose: () => void) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    if (!apiKey) return;
    try {
      await navigator.clipboard.writeText(apiKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error('Failed to copy to clipboard');
    }
  }, [apiKey]);

  const actions = useMemo(() => [{ label: 'Done', onClick: onClose }], [onClose]);

  return { copied, handleCopy, actions };
}

export function useCreateApiKeyDialog(
  open: boolean,
  onOpenChange: (open: boolean) => void,
  onCreate: (data: { name: string; expiresIn?: number }) => Promise<unknown>,
  isCreating: boolean,
) {
  const [name, setName] = useState('');
  const [expiresIn, setExpiresIn] = useState('0');

  useEffect(() => {
    if (!open) {
      setName('');
      setExpiresIn('0');
    }
  }, [open]);

  const handleSubmit = useCallback(
    async (e: FormEvent) => {
      e.preventDefault();
      if (!name.trim()) return;
      const expiry = parseInt(expiresIn, 10);
      const result = await onCreate({
        name: name.trim(),
        ...(expiry > 0 ? { expiresIn: expiry } : {}),
      });
      if (result) onOpenChange(false);
    },
    [name, expiresIn, onCreate, onOpenChange],
  );

  const submitForm = useCallback(() => {
    const form = document.getElementById('create-api-key-form') as HTMLFormElement;
    form?.requestSubmit();
  }, []);

  const actions = useMemo(
    () => [
      { label: 'Cancel', onClick: () => onOpenChange(false), variant: 'outline' as const },
      {
        label: isCreating ? 'Creating...' : 'Create',
        onClick: submitForm,
        disabled: !name.trim() || isCreating,
        loading: isCreating,
      },
    ],
    [onOpenChange, isCreating, name, submitForm],
  );

  const handleNameChange = useCallback((e: ChangeEvent<HTMLInputElement>) => {
    setName(e.target.value);
  }, []);

  return { name, expiresIn, setExpiresIn, handleSubmit, actions, handleNameChange };
}

export function useRenameApiKeyDialog(
  open: boolean,
  onOpenChange: (open: boolean) => void,
  currentName: string,
  onRename: (name: string) => Promise<boolean>,
  isUpdating: boolean,
) {
  const [name, setName] = useState(currentName);

  useEffect(() => {
    if (open) setName(currentName);
  }, [open, currentName]);

  const handleSubmit = useCallback(
    async (e: FormEvent) => {
      e.preventDefault();
      if (!name.trim() || name.trim() === currentName) return;
      const ok = await onRename(name.trim());
      if (ok) onOpenChange(false);
    },
    [name, currentName, onRename, onOpenChange],
  );

  const submitForm = useCallback(() => {
    const form = document.getElementById('rename-api-key-form') as HTMLFormElement;
    form?.requestSubmit();
  }, []);

  const handleNameChange = useCallback((e: ChangeEvent<HTMLInputElement>) => {
    setName(e.target.value);
  }, []);

  const actions = useMemo(
    () => [
      { label: 'Cancel', onClick: () => onOpenChange(false), variant: 'outline' as const },
      {
        label: isUpdating ? 'Renaming...' : 'Rename',
        onClick: submitForm,
        disabled: !name.trim() || name.trim() === currentName || isUpdating,
        loading: isUpdating,
      },
    ],
    [onOpenChange, isUpdating, name, currentName, submitForm],
  );

  return { name, handleSubmit, actions, handleNameChange };
}

export function useDeleteApiKeyDialog(
  onOpenChange: (open: boolean) => void,
  onDelete: () => void,
  isDeleting: boolean,
) {
  const actions = useMemo(
    () => [
      { label: 'Cancel', onClick: () => onOpenChange(false), variant: 'outline' as const },
      {
        label: isDeleting ? 'Deleting...' : 'Delete',
        onClick: onDelete,
        variant: 'destructive' as const,
        disabled: isDeleting,
        loading: isDeleting,
      },
    ],
    [onOpenChange, isDeleting, onDelete],
  );

  return { actions };
}

export function useApiKeysListDialogs(
  onRename: (keyId: string, name: string) => Promise<boolean>,
  onDelete: (keyId: string) => Promise<void>,
) {
  const [renameTarget, setRenameTarget] = useState<ApiKey | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ApiKey | null>(null);

  const handleCloseRename = useCallback((open: boolean) => {
    if (!open) setRenameTarget(null);
  }, []);

  const handleCloseDelete = useCallback((open: boolean) => {
    if (!open) setDeleteTarget(null);
  }, []);

  const handleRenameKey = useCallback(
    async (name: string): Promise<boolean> => {
      if (renameTarget) return onRename(renameTarget.id, name);
      return false;
    },
    [renameTarget, onRename],
  );

  const handleDeleteKey = useCallback(() => {
    if (deleteTarget) {
      onDelete(deleteTarget.id);
      setDeleteTarget(null);
    }
  }, [deleteTarget, onDelete]);

  return {
    renameTarget,
    setRenameTarget,
    deleteTarget,
    setDeleteTarget,
    handleCloseRename,
    handleCloseDelete,
    handleRenameKey,
    handleDeleteKey,
  };
}
