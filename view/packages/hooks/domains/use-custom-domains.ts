'use client';

import { useState, useCallback, useMemo } from 'react';
import {
  useListAllDomainsQuery,
  useAddCustomDomainMutation,
  useVerifyCustomDomainMutation,
  useRemoveCustomDomainMutation,
} from '@/redux/services/domains/customDomainsApi';
import { toast } from 'sonner';
import type { DNSInstruction } from '@/redux/types/custom-domains';

export function useCustomDomains() {
  const { data: allDomains, isLoading, isFetching, error, refetch } = useListAllDomainsQuery();
  const [addDomain, { isLoading: isAdding }] = useAddCustomDomainMutation();
  const [verifyDomain, { isLoading: isVerifying }] = useVerifyCustomDomainMutation();
  const [removeDomain, { isLoading: isRemoving }] = useRemoveCustomDomainMutation();

  const [dnsInstructions, setDnsInstructions] = useState<DNSInstruction[]>([]);
  const [dnsProvider, setDnsProvider] = useState<string>('');
  const [showDnsWizard, setShowDnsWizard] = useState(false);
  const [currentDomainId, setCurrentDomainId] = useState<string>('');

  const handleAddDomain = useCallback(
    async (name: string): Promise<boolean> => {
      try {
        const result = await addDomain({ name }).unwrap();
        setDnsInstructions(result.instructions);
        setDnsProvider(result.dns_provider);
        setCurrentDomainId(result.data.id);
        setShowDnsWizard(true);
        toast.success('Domain added', { description: 'Follow the DNS setup instructions' });
        return true;
      } catch (error: unknown) {
        const err = error as { data?: { message?: string }; message?: string };
        toast.error('Failed to add domain', {
          description: err?.data?.message || err?.message || 'Unknown error',
        });
        return false;
      }
    },
    [addDomain],
  );

  const handleVerifyDomain = useCallback(
    async (id: string) => {
      try {
        await verifyDomain({ id }).unwrap();
        toast.success('DNS verified', { description: 'Your domain is being activated' });
        setShowDnsWizard(false);
      } catch (error: unknown) {
        const err = error as { data?: { message?: string } };
        toast.error('DNS verification failed', {
          description:
            err?.data?.message ||
            'DNS records not yet propagated. Please try again in a few minutes.',
        });
      }
    },
    [verifyDomain],
  );

  const handleRemoveDomain = useCallback(
    async (id: string): Promise<boolean> => {
      try {
        await removeDomain({ id }).unwrap();
        toast.success('Domain removed');
        return true;
      } catch (error: unknown) {
        const err = error as { data?: { message?: string }; message?: string };
        toast.error('Failed to remove domain', {
          description: err?.data?.message || err?.message || 'Unknown error',
        });
        return false;
      }
    },
    [removeDomain],
  );

  const domains = useMemo(() => allDomains || [], [allDomains]);

  return {
    domains,
    isLoading,
    isFetching,
    error,
    isAdding,
    isVerifying,
    isRemoving,
    dnsInstructions,
    dnsProvider,
    showDnsWizard,
    currentDomainId,
    setShowDnsWizard,
    handleAddDomain,
    handleVerifyDomain,
    handleRemoveDomain,
    refetch,
  };
}
