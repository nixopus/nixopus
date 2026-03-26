'use client';

import React, { useEffect, useState } from 'react';
import type { MCPProvider } from '@/redux/types/mcp';
import { getBaseUrl } from '@/redux/conf';

interface Props {
  providers: MCPProvider[];
  onSelect: (provider: MCPProvider) => void;
}

export function MCPProviderPicker({ providers, onSelect }: Props) {
  const [baseUrl, setBaseUrl] = useState('');

  useEffect(() => {
    getBaseUrl()
      .then(setBaseUrl)
      .catch(() => {});
  }, []);

  const resolveIconUrl = (logoUrl: string) => {
    if (!logoUrl) return '';
    if (logoUrl.startsWith('http')) return logoUrl;
    return `${baseUrl}${logoUrl}`;
  };

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 p-4">
      {providers.map((provider) => (
        <button
          key={provider.id}
          onClick={() => onSelect(provider)}
          className="flex flex-col items-center gap-2 p-4 rounded-lg border border-border hover:border-primary hover:bg-accent transition-colors cursor-pointer"
        >
          <div className="w-10 h-10 flex items-center justify-center">
            {baseUrl && provider.logo_url ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={resolveIconUrl(provider.logo_url)}
                alt={provider.name}
                width={40}
                height={40}
                className="object-contain"
              />
            ) : (
              <div className="w-10 h-10 rounded bg-muted" />
            )}
          </div>
          <span className="text-sm font-medium text-center">{provider.name}</span>
          {provider.description && (
            <span className="text-xs text-muted-foreground text-center line-clamp-2">
              {provider.description}
            </span>
          )}
        </button>
      ))}
    </div>
  );
}
