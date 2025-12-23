'use client';

import React, { memo } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import { Box, Wifi, WifiOff, AlertTriangle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import {
  STATUS_COLORS,
  HEALTH_COLORS,
  ROLE_ICONS_SMALL,
  ROLE_BADGE_COLORS,
  ROLE_BG_COLORS
} from '../constants';
import { useServerNode } from '../hooks/use-server-node';
import type {
  ServerNodeType,
  NodeStatusIndicatorProps,
  NodeServerIconProps,
  NodeLabelsProps,
  MetricBarProps,
  NodeMetricsSectionProps,
  NodeStatusMessageProps
} from '../types';
import type { Server as ServerType } from '@/redux/types/servers';

function NodeHandles({ role }: { role: ServerType['role'] }) {
  if (role === 'standalone') return null;

  return (
    <>
      <Handle
        type="target"
        position={Position.Top}
        className="!w-3 !h-3 !bg-primary !border-2 !border-background"
      />
      <Handle
        type="source"
        position={Position.Bottom}
        className="!w-3 !h-3 !bg-primary !border-2 !border-background"
      />
    </>
  );
}

function StatusIndicator({ status }: NodeStatusIndicatorProps) {
  return (
    <div className="absolute -top-1.5 -right-1.5">
      <span className={cn('relative flex h-4 w-4', STATUS_COLORS[status], 'rounded-full')}>
        {status === 'online' && (
          <span
            className={cn(
              'animate-ping absolute inline-flex h-full w-full rounded-full opacity-75',
              STATUS_COLORS[status]
            )}
          />
        )}
        <span className={cn('relative inline-flex rounded-full h-4 w-4', STATUS_COLORS[status])} />
      </span>
    </div>
  );
}

function ServerIcon({ status, health, role }: NodeServerIconProps) {
  return (
    <div className={cn('p-2 rounded-lg', ROLE_BG_COLORS[role])}>
      {status === 'online' ? (
        <Wifi className={cn('h-5 w-5', HEALTH_COLORS[health])} />
      ) : status === 'error' ? (
        <AlertTriangle className="h-5 w-5 text-red-500" />
      ) : (
        <WifiOff className="h-5 w-5 text-zinc-400" />
      )}
    </div>
  );
}

function NodeLabels({ role, roleLabel, visibleLabels, remainingCount }: NodeLabelsProps) {
  return (
    <div className="flex items-center gap-2 mb-3">
      <Badge variant="outline" className={cn('text-xs gap-1', ROLE_BADGE_COLORS[role])}>
        {ROLE_ICONS_SMALL[role]}
        {roleLabel}
      </Badge>
      {visibleLabels.map((label) => (
        <Badge key={label} variant="secondary" className="text-xs">
          {label}
        </Badge>
      ))}
      {remainingCount > 0 && (
        <Badge variant="secondary" className="text-xs">
          +{remainingCount}
        </Badge>
      )}
    </div>
  );
}

function MetricBar({ config, getColor }: MetricBarProps) {
  const Icon = config.icon;
  return (
    <div className="flex items-center gap-2">
      <span className="text-muted-foreground">
        <Icon className="h-3 w-3" />
      </span>
      <div className="flex-1 h-1.5 bg-secondary rounded-full overflow-hidden">
        <div
          className={cn('h-full transition-all', getColor(config.value))}
          style={{ width: `${config.value}%` }}
        />
      </div>
      <span className="text-xs text-muted-foreground w-8 text-right">{config.value}%</span>
    </div>
  );
}

function MetricsSection({
  metricBars,
  containerCountText,
  getMetricBarColor
}: NodeMetricsSectionProps) {
  return (
    <div className="space-y-2">
      {metricBars.map((bar) => (
        <MetricBar key={bar.key} config={bar} getColor={getMetricBarColor} />
      ))}
      <div className="flex items-center gap-2 pt-1 text-xs text-muted-foreground">
        <Box className="h-3 w-3" />
        <span>{containerCountText}</span>
      </div>
    </div>
  );
}

function StatusMessage({ status, offlineText, connectingText }: NodeStatusMessageProps) {
  if (status === 'offline') {
    return <div className="text-xs text-muted-foreground text-center py-2">{offlineText}</div>;
  }

  if (status === 'connecting') {
    return (
      <div className="text-xs text-amber-500 text-center py-2 flex items-center justify-center gap-2">
        <span className="animate-spin h-3 w-3 border-2 border-amber-500 border-t-transparent rounded-full" />
        {connectingText}
      </div>
    );
  }

  return null;
}

function ServerNodeComponent({ data }: NodeProps<ServerNodeType>) {
  const { server, isSelected, onSelect } = data;

  const {
    metricBars,
    visibleLabels,
    remainingLabelsCount,
    getMetricBarColor,
    containerCountText,
    offlineText,
    connectingText,
    roleLabel
  } = useServerNode({ server });

  return (
    <div
      className={cn(
        'relative bg-card border-2 rounded-xl p-4 min-w-[220px] max-w-[280px] cursor-pointer transition-all duration-200',
        'hover:shadow-lg hover:shadow-primary/5 hover:border-primary/50',
        isSelected ? 'border-primary shadow-lg shadow-primary/10' : 'border-border',
        server.status === 'offline' && 'opacity-60'
      )}
      onClick={() => onSelect(server)}
    >
      <NodeHandles role={server.role} />
      <StatusIndicator status={server.status} />

      <div className="flex items-start gap-3 mb-3">
        <ServerIcon status={server.status} health={server.health} role={server.role} />
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-sm truncate">{server.name}</h3>
          <p className="text-xs text-muted-foreground truncate">
            {server.host}:{server.port}
          </p>
        </div>
      </div>

      <NodeLabels
        role={server.role}
        roleLabel={roleLabel}
        visibleLabels={visibleLabels}
        remainingCount={remainingLabelsCount}
      />

      {server.metrics && server.status === 'online' && (
        <MetricsSection
          metricBars={metricBars}
          containerCountText={containerCountText}
          getMetricBarColor={getMetricBarColor}
        />
      )}

      <StatusMessage
        status={server.status}
        offlineText={offlineText}
        connectingText={connectingText}
      />
    </div>
  );
}

export const ServerNode = memo(ServerNodeComponent);
