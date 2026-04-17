'use client';

import { useState, useMemo, useEffect, useCallback } from 'react';
import { authClient } from '@/packages/lib/auth-client';
import { toast } from 'sonner';

export interface Passkey {
  id: string;
  name: string | null;
  credentialID: string;
  deviceType: string;
  backedUp: boolean;
  createdAt: Date | string | null;
}

async function fetchPasskeyList(): Promise<{ data: Passkey[]; error: Error | null }> {
  try {
    const res = await fetch('/api/auth/passkey/list-user-passkeys', {
      method: 'GET',
      credentials: 'include',
    });
    if (!res.ok) {
      return { data: [], error: new Error(`Failed to load passkeys (${res.status})`) };
    }
    const data = await res.json();
    return { data: (data as Passkey[]) ?? [], error: null };
  } catch (e) {
    return { data: [], error: e instanceof Error ? e : new Error('Failed to load passkeys') };
  }
}

async function apiDeletePasskey(id: string): Promise<boolean> {
  const res = await fetch('/api/auth/passkey/delete-passkey', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id }),
  });
  return res.ok;
}

async function apiUpdatePasskey(id: string, name: string): Promise<boolean> {
  const res = await fetch('/api/auth/passkey/update-passkey', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, name }),
  });
  return res.ok;
}

export function usePasskeys() {
  const [passkeys, setPasskeys] = useState<Passkey[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isAdding, setIsAdding] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [isFetching, setIsFetching] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [addDialogOpen, setAddDialogOpen] = useState(false);

  const refetch = useCallback(async () => {
    setIsFetching(true);
    const result = await fetchPasskeyList();
    setPasskeys(result.data);
    setError(result.error);
    setIsLoading(false);
    setIsFetching(false);
  }, []);

  useEffect(() => {
    refetch();
  }, [refetch]);

  const addPasskey = useCallback(
    async (name?: string) => {
      setIsAdding(true);
      try {
        const res = await authClient.passkey.addPasskey({ name });
        if (res?.error) {
          toast.error(res.error.message || 'Failed to register passkey');
          return false;
        }
        toast.success('Passkey registered successfully');
        await refetch();
        return true;
      } catch (error: unknown) {
        const msg = error instanceof Error ? error.message : 'Failed to register passkey';
        if (!msg.includes('cancelled') && !msg.includes('abort')) {
          toast.error(msg);
        }
        return false;
      } finally {
        setIsAdding(false);
      }
    },
    [refetch],
  );

  const deletePasskey = useCallback(
    async (id: string) => {
      setIsDeleting(true);
      try {
        const ok = await apiDeletePasskey(id);
        if (!ok) {
          toast.error('Failed to delete passkey');
          return;
        }
        toast.success('Passkey deleted');
        await refetch();
      } catch (error: unknown) {
        toast.error(error instanceof Error ? error.message : 'Failed to delete passkey');
      } finally {
        setIsDeleting(false);
      }
    },
    [refetch],
  );

  const updatePasskey = useCallback(
    async (id: string, name: string): Promise<boolean> => {
      setIsUpdating(true);
      try {
        const ok = await apiUpdatePasskey(id, name);
        if (!ok) {
          toast.error('Failed to update passkey');
          return false;
        }
        toast.success('Passkey renamed');
        await refetch();
        return true;
      } catch (error: unknown) {
        toast.error(error instanceof Error ? error.message : 'Failed to update passkey');
        return false;
      } finally {
        setIsUpdating(false);
      }
    },
    [refetch],
  );

  const filteredPasskeys = useMemo(() => {
    if (!searchTerm.trim()) return passkeys;
    const q = searchTerm.toLowerCase();
    return passkeys.filter((pk) => (pk.name || '').toLowerCase().includes(q));
  }, [passkeys, searchTerm]);

  return {
    passkeys: filteredPasskeys,
    hasPasskeys: passkeys.length > 0,
    isLoading,
    isFetching,
    error,
    isAdding,
    isDeleting,
    isUpdating,
    addPasskey,
    deletePasskey,
    updatePasskey,
    refetch,
    searchTerm,
    setSearchTerm,
    addDialogOpen,
    setAddDialogOpen,
  };
}
