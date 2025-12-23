'use client';

import React from 'react';
import {
  Cpu,
  MemoryStick,
  HardDriveIcon,
  Box,
  Wifi,
  WifiOff,
  AlertTriangle,
  MoreVertical,
  Plug
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import {
  STATUS_COLORS,
  STATUS_TEXT_COLORS,
  HEALTH_COLORS,
  ROLE_ICONS,
  ROLE_BADGE_COLORS,
  ROLE_BG_COLORS
} from '../constants';
import { useServerCard } from '../hooks/use-server-card';
import type {
  ServerCardProps,
  CardStatusIndicatorProps,
  CardServerIconProps,
  ServerLabelsProps,
  MetricsGridProps,
  CardStatusMessageProps,
  CardActionsProps
} from '../types';

function StatusIndicator({ status }: CardStatusIndicatorProps) {
  return (
    <div className="absolute top-4 right-4">
      <span className={cn('relative flex h-3 w-3', STATUS_COLORS[status], 'rounded-full')}>
        {status === 'online' && (
          <span
            className={cn(
              'animate-ping absolute inline-flex h-full w-full rounded-full opacity-75',
              STATUS_COLORS[status]
            )}
          />
        )}
      </span>
    </div>
  );
}

function ServerIcon({ status, health, role }: CardServerIconProps) {
  return (
    <div className={cn('p-2.5 rounded-lg', ROLE_BG_COLORS[role])}>
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

function ServerLabels({
  visibleLabels,
  remainingCount,
  role,
  status,
  statusText
}: ServerLabelsProps) {
  return (
    <div className="flex items-center gap-2 flex-wrap min-w-0 w-full overflow-hidden">
      <Badge
        variant="outline"
        className={cn('gap-1 min-w-0 overflow-hidden', ROLE_BADGE_COLORS[role])}
        style={{ maxWidth: 'calc(50% - 4px)', width: 'max-content' }}
      >
        {ROLE_ICONS[role]}
        <span className="truncate inline-block max-w-full min-w-0">
          {role.charAt(0).toUpperCase() + role.slice(1)}
        </span>
      </Badge>
      <Badge
        variant="secondary"
        className={cn('text-xs min-w-0 overflow-hidden', STATUS_TEXT_COLORS[status])}
        style={{ maxWidth: 'calc(50% - 4px)', width: 'max-content' }}
      >
        <span className="truncate inline-block max-w-full min-w-0">{statusText}</span>
      </Badge>
      {visibleLabels.map((label) => (
        <Badge
          key={label}
          variant="outline"
          className="text-xs min-w-0 overflow-hidden"
          style={{ maxWidth: 'calc(50% - 4px)', width: 'max-content' }}
        >
          <span className="truncate inline-block max-w-full min-w-0">{label}</span>
        </Badge>
      ))}
      {remainingCount > 0 && (
        <Badge variant="outline" className="text-xs shrink-0">
          +{remainingCount}
        </Badge>
      )}
    </div>
  );
}

function MetricBox({
  icon,
  label,
  value,
  colorClass
}: {
  icon: React.ReactNode;
  label: string;
  value: number;
  colorClass: string;
}) {
  return (
    <div className="bg-muted/50 rounded-lg p-2 text-center">
      <div className="flex items-center justify-center gap-1 text-muted-foreground mb-1">
        {icon}
        <span className="text-xs">{label}</span>
      </div>
      <span className={cn('text-lg font-bold', colorClass)}>{value}%</span>
    </div>
  );
}

function MetricsGrid({ metrics, getMetricColor }: MetricsGridProps) {
  const metricItems = [
    { key: 'cpu', icon: <Cpu className="h-4 w-4" />, label: 'CPU', value: metrics.cpu_usage },
    {
      key: 'ram',
      icon: <MemoryStick className="h-4 w-4" />,
      label: 'RAM',
      value: metrics.memory_usage
    },
    {
      key: 'disk',
      icon: <HardDriveIcon className="h-4 w-4" />,
      label: 'Disk',
      value: metrics.disk_usage
    }
  ];

  return (
    <div className="grid grid-cols-3 gap-3">
      {metricItems.map((metric) => (
        <MetricBox
          key={metric.key}
          icon={metric.icon}
          label={metric.label}
          value={metric.value}
          colorClass={getMetricColor(metric.value)}
        />
      ))}
    </div>
  );
}

function StatusMessage({ status, statusText }: CardStatusMessageProps) {
  if (status === 'offline') {
    return (
      <div className="text-sm text-muted-foreground text-center py-2 bg-muted/50 rounded-lg">
        {statusText}
      </div>
    );
  }

  if (status === 'connecting') {
    return (
      <div className="text-sm text-amber-500 text-center py-2 bg-amber-500/10 rounded-lg flex items-center justify-center gap-2">
        <span className="animate-spin h-4 w-4 border-2 border-amber-500 border-t-transparent rounded-full" />
        {statusText}...
      </div>
    );
  }

  return null;
}

function CardActions({
  serverId,
  menuItems,
  onTestConnection,
  isTestingConnection,
  testConnectionLabel
}: CardActionsProps) {
  return (
    <div className="flex items-center gap-2 pt-2">
      <Button
        variant="outline"
        size="sm"
        className="flex-1"
        onClick={(e) => {
          e.stopPropagation();
          onTestConnection(serverId);
        }}
        disabled={isTestingConnection}
      >
        <Plug className="h-4 w-4 mr-1.5" />
        {testConnectionLabel}
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="h-8 w-8 p-0"
            onClick={(e) => e.stopPropagation()}
          >
            <MoreVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {menuItems.map((item) => (
            <React.Fragment key={item.key}>
              {item.separator && <DropdownMenuSeparator />}
              <DropdownMenuItem
                onClick={item.onClick}
                disabled={item.disabled}
                className={item.destructive ? 'text-destructive' : undefined}
              >
                <item.icon className="h-4 w-4 mr-2" />
                {item.label}
              </DropdownMenuItem>
            </React.Fragment>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

export function ServerCard({
  server,
  isSelected,
  onSelect,
  onEdit,
  onDelete,
  onTestConnection,
  isTestingConnection
}: ServerCardProps) {
  const { t, menuItems, visibleLabels, remainingLabelsCount, getMetricColor } = useServerCard({
    server,
    onEdit,
    onDelete
  });

  return (
    <Card
      className={cn(
        'relative cursor-pointer transition-all duration-200 hover:shadow-md w-full overflow-hidden',
        isSelected && 'ring-2 ring-primary shadow-md',
        server.status === 'offline' && 'opacity-70'
      )}
      onClick={() => onSelect(server)}
    >
      <StatusIndicator status={server.status} />

      <CardHeader className="pb-2 overflow-hidden">
        <div className="flex items-start gap-3 min-w-0">
          <ServerIcon status={server.status} health={server.health} role={server.role} />
          <div className="flex-1 min-w-0 pr-6">
            <h3 className="font-semibold text-base truncate">{server.name}</h3>
            <p className="text-sm text-muted-foreground truncate">
              {server.host}:{server.port}
            </p>
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-4 overflow-hidden">
        <div className="min-w-0 w-full overflow-hidden">
          <ServerLabels
            visibleLabels={visibleLabels}
            remainingCount={remainingLabelsCount}
            role={server.role}
            status={server.status}
            statusText={t(`servers.status.${server.status}`)}
          />
        </div>

        {server.metrics && server.status === 'online' && (
          <>
            <MetricsGrid metrics={server.metrics} getMetricColor={getMetricColor} />
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Box className="h-4 w-4" />
              <span>
                {server.metrics.container_count} {t('servers.details.containers').toLowerCase()}
              </span>
            </div>
          </>
        )}

        <StatusMessage status={server.status} statusText={t(`servers.status.${server.status}`)} />

        <CardActions
          serverId={server.id}
          menuItems={menuItems}
          onTestConnection={onTestConnection}
          isTestingConnection={isTestingConnection}
          testConnectionLabel={t('servers.testConnection')}
        />
      </CardContent>
    </Card>
  );
}
