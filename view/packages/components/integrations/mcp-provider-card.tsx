'use client';

import React from 'react';
import { Button, Badge, Skeleton, Card, CardContent, CardHeader } from '@nixopus/ui';
import type { MCPProvider, MCPServer } from '@/redux/types/mcp';

interface MCPProviderCardProps {
  provider: MCPProvider;
  server: MCPServer | null;
  isLoading: boolean;
  onConfigure: (provider: MCPProvider) => void;
  canConfigure: boolean;
  iconBaseUrl: string;
}

function resolveIconUrl(logoUrl: string, baseUrl: string): string {
  if (!logoUrl) return '';
  if (logoUrl.startsWith('http')) return logoUrl;
  return `${baseUrl}${logoUrl}`;
}

export function MCPProviderCard({
  provider,
  server,
  isLoading,
  onConfigure,
  canConfigure,
  iconBaseUrl
}: MCPProviderCardProps) {
  const isConnected = server?.enabled === true;
  const isConfigured = !!server;

  const badge = isConnected
    ? { label: 'Connected', variant: 'default' as const }
    : isConfigured
      ? { label: 'Configured', variant: 'outline' as const }
      : { label: 'Not Set Up', variant: 'secondary' as const };

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center gap-3 pb-2">
          <Skeleton className="h-8 w-8 rounded-md" />
          <div className="space-y-1 flex-1">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-3 w-16" />
          </div>
        </CardHeader>
        <CardContent className="flex items-center justify-between pt-2">
          <Skeleton className="h-5 w-20 rounded-full" />
          <Skeleton className="h-8 w-24 rounded-md" />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center gap-3 pb-2">
        <div className="flex h-8 w-8 items-center justify-center rounded-md bg-muted">
          {iconBaseUrl && provider.logo_url ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={resolveIconUrl(provider.logo_url, iconBaseUrl)}
              alt={provider.name}
              width={20}
              height={20}
              className="object-contain"
            />
          ) : (
            <div className="h-4 w-4 rounded bg-muted-foreground/20" />
          )}
        </div>
        <div>
          <p className="text-sm font-medium leading-none">{provider.name}</p>
          <p className="text-xs text-muted-foreground mt-0.5">MCP Server</p>
        </div>
      </CardHeader>
      <CardContent className="flex items-center justify-between pt-2">
        <Badge variant={badge.variant}>{badge.label}</Badge>
        {canConfigure && (
          <Button variant="outline" size="sm" onClick={() => onConfigure(provider)}>
            Configure
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
