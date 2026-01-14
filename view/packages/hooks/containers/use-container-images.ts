import { useState, useCallback } from 'react';

export const useContainerImages = () => {
  const [copied, setCopied] = useState<string | null>(null);

  const copyToClipboard = useCallback((text: string, key: string) => {
    navigator.clipboard.writeText(text);
    setCopied(key);
    setTimeout(() => setCopied(null), 2000);
  }, []);

  return {
    copied,
    copyToClipboard
  };
};
