import { useState, useMemo, useCallback } from 'react';
import {
  useListApiKeysQuery,
  useCreateApiKeyMutation,
  useUpdateApiKeyMutation,
  useDeleteApiKeyMutation,
  type CreateApiKeyRequest,
} from '@/redux/services/api-keys/apiKeysApi';
import { toast } from 'sonner';

const EXPIRY_OPTIONS = [
  { label: 'No expiration', value: 0 },
  { label: '7 days', value: 60 * 60 * 24 * 7 },
  { label: '30 days', value: 60 * 60 * 24 * 30 },
  { label: '60 days', value: 60 * 60 * 24 * 60 },
  { label: '90 days', value: 60 * 60 * 24 * 90 },
  { label: '1 year', value: 60 * 60 * 24 * 365 },
] as const;

export function useApiKeys() {
  const { data: apiKeys, isLoading, isFetching, error, refetch } = useListApiKeysQuery();
  const [createApiKey, { isLoading: isCreating }] = useCreateApiKeyMutation();
  const [updateApiKey, { isLoading: isUpdating }] = useUpdateApiKeyMutation();
  const [deleteApiKey, { isLoading: isDeleting }] = useDeleteApiKeyMutation();

  const [newlyCreatedKey, setNewlyCreatedKey] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');

  const sortedKeys = useMemo(() => {
    if (!apiKeys) return [];
    return [...apiKeys].sort(
      (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
    );
  }, [apiKeys]);

  const filteredKeys = useMemo(() => {
    if (!searchTerm.trim()) return sortedKeys;
    const q = searchTerm.toLowerCase();
    return sortedKeys.filter(
      (k) =>
        (k.name || '').toLowerCase().includes(q) ||
        (k.prefix || '').toLowerCase().includes(q),
    );
  }, [sortedKeys, searchTerm]);

  const handleCreate = useCallback(
    async (data: CreateApiKeyRequest) => {
      try {
        const result = await createApiKey(data).unwrap();
        setNewlyCreatedKey(result.key);
        toast.success('API key created', {
          description: "Copy your key now — it won't be shown again.",
        });
        return result;
      } catch (error: unknown) {
        const msg =
          error instanceof Error
            ? error.message
            : (error as { data?: { message?: string } })?.data?.message ||
              'Failed to create API key';
        toast.error('Failed to create API key', { description: msg });
        return null;
      }
    },
    [createApiKey],
  );

  const handleRename = useCallback(
    async (keyId: string, name: string): Promise<boolean> => {
      try {
        await updateApiKey({ keyId, name }).unwrap();
        toast.success('API key renamed');
        return true;
      } catch (error: unknown) {
        const msg =
          error instanceof Error
            ? error.message
            : (error as { data?: { message?: string } })?.data?.message ||
              'Failed to rename API key';
        toast.error('Failed to rename', { description: msg });
        return false;
      }
    },
    [updateApiKey],
  );

  const handleDelete = useCallback(
    async (keyId: string) => {
      try {
        await deleteApiKey({ keyId }).unwrap();
        toast.success('API key deleted');
      } catch (error: unknown) {
        const msg =
          error instanceof Error
            ? error.message
            : (error as { data?: { message?: string } })?.data?.message ||
              'Failed to delete API key';
        toast.error('Failed to delete', { description: msg });
      }
    },
    [deleteApiKey],
  );

  const clearNewKey = useCallback(() => setNewlyCreatedKey(null), []);

  return {
    apiKeys: filteredKeys,
    isLoading,
    isFetching,
    error,
    isCreating,
    isUpdating,
    isDeleting,
    newlyCreatedKey,
    clearNewKey,
    handleCreate,
    handleRename,
    handleDelete,
    refetch,
    expiryOptions: EXPIRY_OPTIONS,
    searchTerm,
    setSearchTerm,
    createDialogOpen,
    setCreateDialogOpen,
  };
}
