import React from 'react';
import { ApplicationDomain } from '@/redux/types/applications';
import { SHARED_DOMAIN_REGEX } from '@/packages/utils/util';

export const defaultValidator = (value: string): boolean => {
  if (!value || value.trim() === '') return true; // Empty is allowed
  return SHARED_DOMAIN_REGEX.test(value.trim());
};

interface UseMultipleDomainsProps {
  form: any;
  name: string;
  maxDomains?: number;
  validator?: (value: string) => boolean;
  maxVisible?: number;
}

export function useMultipleDomains({
  form,
  name,
  maxDomains = 5,
  validator = defaultValidator,
  maxVisible = 2
}: UseMultipleDomainsProps) {
  const domains = form.watch(name) || [];
  const [showAll, setShowAll] = React.useState(false);

  const addDomain = React.useCallback(() => {
    if (domains.length < maxDomains) {
      form.setValue(name, [...domains, ''], { shouldValidate: true });
    }
  }, [domains, maxDomains, form, name]);

  const removeDomain = React.useCallback(
    (index: number) => {
      const newDomains = domains.filter((_: string, i: number) => i !== index);
      form.setValue(name, newDomains, { shouldValidate: true });
      // If we remove a domain and now have <= maxVisible, hide the "show all" state
      if (newDomains.length <= maxVisible) {
        setShowAll(false);
      }
    },
    [domains, form, name, maxVisible]
  );

  const updateDomain = React.useCallback(
    (index: number, value: string) => {
      const newDomains = [...domains];
      newDomains[index] = value;
      form.setValue(name, newDomains, { shouldValidate: true });
    },
    [domains, form, name]
  );

  // Initialize with one empty domain if none exist
  React.useEffect(() => {
    if (!domains || domains.length === 0) {
      form.setValue(name, [''], { shouldValidate: false });
    }
  }, [domains, form, name]);

  // When collapsed, show the last maxVisible domains (most recent)
  // When expanded, show all domains
  const visibleDomains = React.useMemo(
    () => (showAll ? domains : domains.slice(-maxVisible)),
    [showAll, domains, maxVisible]
  );

  const remainingCount = React.useMemo(
    () => Math.max(0, domains.length - maxVisible),
    [domains.length, maxVisible]
  );

  // Calculate the starting index for visible domains
  const startIndex = React.useMemo(
    () => (showAll ? 0 : Math.max(0, domains.length - maxVisible)),
    [showAll, domains.length, maxVisible]
  );

  const validationRules = React.useMemo(
    () => ({
      validate: (value: any) => {
        if (!value || !Array.isArray(value)) return true;
        const nonEmpty = value.filter((d: string) => d && d.trim() !== '');
        if (nonEmpty.length > maxDomains) {
          return `Maximum ${maxDomains} domains allowed`;
        }
        const unique = new Set(nonEmpty.map((d: string) => d.trim().toLowerCase()));
        if (unique.size !== nonEmpty.length) {
          return 'Domains must be unique';
        }
        const allValid = nonEmpty.every((domain: string) => {
          if (!domain || domain.trim() === '') return true;
          return validator(domain.trim());
        });
        if (!allValid) {
          return 'Invalid domain format';
        }
        return true;
      }
    }),
    [maxDomains, validator]
  );

  return {
    domains,
    showAll,
    setShowAll,
    addDomain,
    removeDomain,
    updateDomain,
    visibleDomains,
    remainingCount,
    startIndex,
    validationRules,
    maxVisible
  };
}

interface UseDomainDropdownProps {
  domains?: ApplicationDomain[];
  onDomainClick?: (domain: string) => void;
}

export function useDomainDropdown({ domains, onDomainClick }: UseDomainDropdownProps) {
  // Get all domains from array
  const allDomains = React.useMemo(() => {
    if (domains && domains.length > 0) {
      return domains.map((d) => d.domain);
    }
    return [];
  }, [domains]);

  const handleDomainClick = React.useCallback(
    (domainValue: string, e?: React.MouseEvent) => {
      if (e) {
        e.stopPropagation();
      }
      if (onDomainClick) {
        onDomainClick(domainValue);
      } else {
        window.open(`https://${domainValue}`, '_blank', 'noopener,noreferrer');
      }
    },
    [onDomainClick]
  );

  return {
    allDomains,
    handleDomainClick
  };
}
