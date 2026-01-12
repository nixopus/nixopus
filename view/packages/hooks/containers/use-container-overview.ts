import { useState, useCallback } from 'react';

export const useContainerOverview = () => {
  const [copied, setCopied] = useState<string | null>(null);
  const [showRaw, setShowRaw] = useState(false);

  const copyToClipboard = useCallback((text: string, key: string) => {
    navigator.clipboard.writeText(text);
    setCopied(key);
    setTimeout(() => setCopied(null), 2000);
  }, []);

  return {
    copied,
    showRaw,
    setShowRaw,
    copyToClipboard
  };
};
