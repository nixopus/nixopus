'use client';

import React from 'react';
import { Server } from '@/redux/types/servers';
import { RoutingStrategy } from '@/redux/types/applications';
import { Label } from '@nixopus/ui';
import { Badge } from '@nixopus/ui';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@nixopus/ui';
import { cn } from '@/lib/utils';
import { Server as ServerIcon } from 'lucide-react';

interface ServerSelectorProps {
  servers: Server[];
  selectedServerIds: string[];
  onToggleServer: (id: string) => void;
  routingStrategy: RoutingStrategy | null;
  onRoutingStrategyChange: (strategy: RoutingStrategy) => void;
  primaryServerId: string | null;
  onPrimaryServerIdChange: (id: string) => void;
}

export function ServerSelector({
  servers,
  selectedServerIds,
  onToggleServer,
  routingStrategy,
  onRoutingStrategyChange,
  primaryServerId,
  onPrimaryServerIdChange
}: ServerSelectorProps) {
  if (servers.length < 2) return null;

  const selectedServers = servers.filter((s) => selectedServerIds.includes(s.id));
  const showStrategyControls = selectedServerIds.length >= 2;

  return (
    <div className="space-y-3">
      <Label className="text-sm font-medium flex items-center gap-2">
        <ServerIcon className="h-4 w-4" />
        Deploy to Servers
      </Label>

      <div className="flex flex-wrap gap-2">
        {servers.map((server) => {
          const isSelected = selectedServerIds.includes(server.id);
          return (
            <Badge
              key={server.id}
              variant={isSelected ? 'default' : 'outline'}
              className={cn(
                'cursor-pointer transition-colors px-3 py-1.5',
                isSelected && 'bg-primary text-primary-foreground'
              )}
              onClick={() => onToggleServer(server.id)}
            >
              {server.name?.trim() || server.host?.trim() || server.id}
              {server.is_default && <span className="ml-1 text-xs opacity-70">(default)</span>}
            </Badge>
          );
        })}
      </div>

      {showStrategyControls && (
        <div className="space-y-3 pt-2 border-t">
          <div className="space-y-1.5">
            <Label className="text-sm">Routing Strategy</Label>
            <Select
              value={routingStrategy ?? ''}
              onValueChange={(v) => onRoutingStrategyChange(v as RoutingStrategy)}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Select strategy" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="round_robin">Round Robin — distribute traffic evenly</SelectItem>
                <SelectItem value="primary_failover">
                  Primary Failover — one primary, others as backup
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {routingStrategy === 'primary_failover' && (
            <div className="space-y-1.5">
              <Label className="text-sm">Primary Server</Label>
              <Select value={primaryServerId ?? ''} onValueChange={onPrimaryServerIdChange}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Select primary server" />
                </SelectTrigger>
                <SelectContent>
                  {selectedServers.map((s) => (
                    <SelectItem key={s.id} value={s.id}>
                      {s.name?.trim() || s.host?.trim() || s.id}
                      {s.is_default && ' (default)'}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
