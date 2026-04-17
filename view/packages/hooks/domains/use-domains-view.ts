'use client';

import { useState, useCallback, useMemo } from 'react';
import { toast } from 'sonner';
import { useCustomDomains } from '@/packages/hooks/domains/use-custom-domains';
import type { CustomDomain } from '@/redux/types/custom-domains';

export type { CustomDomain };

export const DOMAIN_REGEX =
  /^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)+$/;

export function isValidDomain(domain: string): boolean {
  if (!domain || domain.length < 4 || domain.length > 253) return false;
  return DOMAIN_REGEX.test(domain);
}

export const statusConfig: Record<
  CustomDomain['status'],
  { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }
> = {
  active: { label: 'Active', variant: 'default' },
  pending_dns: { label: 'Pending DNS', variant: 'outline' },
  dns_verified: { label: 'DNS Verified', variant: 'secondary' },
  failed: { label: 'Failed', variant: 'destructive' },
  removing: { label: 'Removing', variant: 'outline' },
};

export const providerNames: Record<string, string> = {
  cloudflare: 'Cloudflare',
  route53: 'Amazon Route 53',
  godaddy: 'GoDaddy',
  namecheap: 'Namecheap',
  digitalocean: 'DigitalOcean',
  google: 'Google Domains',
  netlify: 'Netlify DNS',
  vercel: 'Vercel DNS',
  azure: 'Azure DNS',
  vultr: 'Vultr',
  hetzner: 'Hetzner DNS',
  ovh: 'OVH',
  gandi: 'Gandi',
  porkbun: 'Porkbun',
  ns1: 'NS1',
  dnsimple: 'DNSimple',
  linode: 'Linode',
  hurricane_electric: 'Hurricane Electric',
  hostgator: 'HostGator',
  bluehost: 'Bluehost',
  siteground: 'SiteGround',
  dreamhost: 'DreamHost',
  bunnycdn: 'Bunny CDN',
};

export function useAddDomainDialog(onAdd: (name: string) => void) {
  const [domainName, setDomainName] = useState('');
  const [validationError, setValidationError] = useState('');

  const handleAdd = useCallback(() => {
    const trimmed = domainName.trim();
    if (!isValidDomain(trimmed)) {
      setValidationError('Enter a valid domain (e.g. app.example.com)');
      return;
    }
    setValidationError('');
    onAdd(trimmed);
    setDomainName('');
  }, [domainName, onAdd]);

  const handleInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setDomainName(e.target.value);
    setValidationError('');
  }, []);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') handleAdd();
    },
    [handleAdd],
  );

  const reset = useCallback(() => {
    setDomainName('');
    setValidationError('');
  }, []);

  return { domainName, validationError, handleAdd, handleInputChange, handleKeyDown, reset };
}

export function useRemoveDomainDialog(onRemove: () => void) {
  const handleRemove = useCallback(() => {
    onRemove();
  }, [onRemove]);

  return { handleRemove };
}

export function useDnsSetupWizard(
  onVerify: (id: string) => void,
  domainId: string,
  onClose: () => void,
) {
  const handleVerify = useCallback(() => {
    onVerify(domainId);
  }, [onVerify, domainId]);

  const handleOpenChange = useCallback(
    (val: boolean) => {
      if (!val) onClose();
    },
    [onClose],
  );

  return { handleVerify, handleOpenChange };
}

export function useCopyField(value: string) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      toast.success('Copied to clipboard');
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error('Failed to copy to clipboard');
    }
  }, [value]);

  return { copied, handleCopy };
}

export function useDomainsView() {
  const domainData = useCustomDomains();

  const [showAddDialog, setShowAddDialog] = useState(false);
  const [removeTarget, setRemoveTarget] = useState<CustomDomain | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortConfig, setSortConfig] = useState<{ field: string; order: 'asc' | 'desc' }>({
    field: 'name',
    order: 'asc',
  });

  const handleSort = useCallback((field: string) => {
    setSortConfig((prev) =>
      prev.field === field
        ? { field, order: prev.order === 'asc' ? 'desc' : 'asc' }
        : { field, order: 'asc' },
    );
  }, []);

  const filteredDomains = useMemo(() => {
    let result = domainData.domains;
    if (searchTerm) {
      const q = searchTerm.toLowerCase();
      result = result.filter((d) => d.name.toLowerCase().includes(q));
    }
    const dir = sortConfig.order === 'asc' ? 1 : -1;
    return [...result].sort((a, b) => {
      switch (sortConfig.field) {
        case 'created':
          return dir * (new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
        case 'type':
          return dir * a.type.localeCompare(b.type);
        case 'status':
          return dir * a.status.localeCompare(b.status);
        default:
          return dir * a.name.localeCompare(b.name);
      }
    });
  }, [domainData.domains, searchTerm, sortConfig]);

  const onAdd = useCallback(
    async (name: string) => {
      const ok = await domainData.handleAddDomain(name);
      if (ok) setShowAddDialog(false);
    },
    [domainData.handleAddDomain],
  );

  const onConfirmRemove = useCallback(async () => {
    if (!removeTarget) return;
    const ok = await domainData.handleRemoveDomain(removeTarget.id);
    if (ok) setRemoveTarget(null);
  }, [removeTarget, domainData.handleRemoveDomain]);

  const onCloseRemoveDialog = useCallback((val: boolean) => {
    if (!val) setRemoveTarget(null);
  }, []);

  const onCloseDnsWizard = useCallback(() => {
    domainData.setShowDnsWizard(false);
  }, [domainData.setShowDnsWizard]);

  return {
    ...domainData,
    filteredDomains,
    showAddDialog,
    setShowAddDialog,
    removeTarget,
    setRemoveTarget,
    searchTerm,
    setSearchTerm,
    sortConfig,
    handleSort,
    onAdd,
    onConfirmRemove,
    onCloseRemoveDialog,
    onCloseDnsWizard,
  };
}
