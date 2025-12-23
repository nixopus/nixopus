'use client';

import React from 'react';
import { X, Wifi, WifiOff, AlertTriangle, Box, Clock, Trash2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';
import { formatDistanceToNow } from 'date-fns';
import {
  STATUS_COLORS,
  STATUS_TEXT_COLORS,
  HEALTH_COLORS,
  ROLE_ICONS,
  ROLE_BADGE_COLORS,
  ROLE_BG_COLORS
} from '../constants';
import { useServerDetailsPanel } from '../hooks/use-server-details-panel';
import type {
  ServerDetailsPanelProps,
  PanelHeaderProps,
  ServerBadgesProps,
  MetricRowProps,
  PanelMetricsSectionProps,
  ClusterActionsSectionProps,
  QuickActionsSectionProps,
  DeleteFooterProps
} from '../types';

function PanelHeader({ server, onClose }: PanelHeaderProps) {
  return (
    <div className="p-4 border-b flex items-start justify-between">
      <div className="flex items-center gap-3">
        <div className={cn('p-2.5 rounded-lg', ROLE_BG_COLORS[server.role])}>
          {server.status === 'online' ? (
            <Wifi className={cn('h-6 w-6', HEALTH_COLORS[server.health])} />
          ) : server.status === 'error' ? (
            <AlertTriangle className="h-6 w-6 text-red-500" />
          ) : (
            <WifiOff className="h-6 w-6 text-zinc-400" />
          )}
        </div>
        <div>
          <h2 className="font-semibold text-lg">{server.name}</h2>
          <p className="text-sm text-muted-foreground">
            {server.host}:{server.port}
          </p>
        </div>
      </div>
      <Button variant="ghost" size="icon" onClick={onClose}>
        <X className="h-4 w-4" />
      </Button>
    </div>
  );
}

function ServerBadges({ server, statusText, healthText }: ServerBadgesProps) {
  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2 flex-wrap">
        <Badge variant="outline" className={cn('gap-1', ROLE_BADGE_COLORS[server.role])}>
          {ROLE_ICONS[server.role]}
          {server.role.charAt(0).toUpperCase() + server.role.slice(1)}
        </Badge>
        <Badge variant="secondary" className={cn('gap-1.5', STATUS_TEXT_COLORS[server.status])}>
          <span className={cn('w-2 h-2 rounded-full', STATUS_COLORS[server.status])} />
          {statusText}
        </Badge>
        <Badge variant="outline" className={cn(HEALTH_COLORS[server.health])}>
          {healthText}
        </Badge>
      </div>

      {server.labels.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {server.labels.map((label) => (
            <Badge key={label} variant="outline" className="text-xs">
              {label}
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
}

function MetricRow({ config, getBarColor, getTextColor }: MetricRowProps) {
  const Icon = config.icon;
  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Icon className="h-4 w-4" />
          <span>{config.label}</span>
        </div>
        <span className={cn('font-semibold', getTextColor(config.value))}>{config.value}%</span>
      </div>
      <div className="h-2 bg-secondary rounded-full overflow-hidden">
        <div
          className={cn('h-full transition-all', getBarColor(config.value))}
          style={{ width: `${config.value}%` }}
        />
      </div>
    </div>
  );
}

function MetricsSection({
  metrics,
  metricRows,
  getBarColor,
  getTextColor,
  containerLabel,
  lastCheckedLabel
}: PanelMetricsSectionProps) {
  return (
    <div className="space-y-4">
      <div className="space-y-4">
        {metricRows.map((row) => (
          <MetricRow
            key={row.key}
            config={row}
            getBarColor={getBarColor}
            getTextColor={getTextColor}
          />
        ))}
      </div>

      <div className="flex items-center justify-between p-3 bg-muted/50 rounded-lg">
        <div className="flex items-center gap-2 text-sm">
          <Box className="h-4 w-4 text-muted-foreground" />
          <span>{containerLabel}</span>
        </div>
        <span className="font-semibold">{metrics.container_count}</span>
      </div>

      {metrics.last_checked_at && (
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Clock className="h-3.5 w-3.5" />
          <span>
            {lastCheckedLabel}:{' '}
            {formatDistanceToNow(new Date(metrics.last_checked_at), { addSuffix: true })}
          </span>
        </div>
      )}
    </div>
  );
}

function OfflineMessage({ message }: { message: string }) {
  return (
    <div className="text-center py-8 text-muted-foreground">
      <WifiOff className="h-12 w-12 mx-auto mb-3 opacity-40" />
      <p>{message}</p>
    </div>
  );
}

function ClusterActionsSection({
  actions,
  showNoClusterMessage,
  noClusterMessage,
  title
}: ClusterActionsSectionProps) {
  return (
    <div className="space-y-3">
      <h3 className="text-sm font-medium">{title}</h3>

      {actions.length > 0 && (
        <div className="space-y-2">
          {actions.map((action) => (
            <Button
              key={action.key}
              variant="outline"
              size="sm"
              className={cn(
                'w-full justify-start',
                action.destructive && 'text-destructive hover:text-destructive'
              )}
              onClick={action.onClick}
              disabled={action.disabled}
            >
              <action.icon className="h-4 w-4 mr-2" />
              {action.label}
            </Button>
          ))}
        </div>
      )}

      {showNoClusterMessage && <p className="text-sm text-muted-foreground">{noClusterMessage}</p>}
    </div>
  );
}

function QuickActionsSection({ actions, title }: QuickActionsSectionProps) {
  return (
    <div className="space-y-3">
      <h3 className="text-sm font-medium">{title}</h3>
      <div className="grid grid-cols-2 gap-2">
        {actions.map((action) => (
          <Button
            key={action.key}
            variant="outline"
            size="sm"
            className="justify-start"
            onClick={action.onClick}
            disabled={action.disabled}
          >
            <action.icon className="h-4 w-4 mr-2" />
            {action.label}
          </Button>
        ))}
      </div>
    </div>
  );
}

function DeleteFooter({ onDelete, label }: DeleteFooterProps) {
  return (
    <div className="p-4 border-t">
      <Button variant="destructive" size="sm" className="w-full" onClick={onDelete}>
        <Trash2 className="h-4 w-4 mr-2" />
        {label}
      </Button>
    </div>
  );
}

export function ServerDetailsPanel({
  server,
  onClose,
  onEdit,
  onDelete,
  onTestConnection,
  onJoinCluster,
  onLeaveCluster,
  onPromote,
  onDemote,
  isTestingConnection,
  hasCluster
}: ServerDetailsPanelProps) {
  const {
    t,
    metricRows,
    clusterActions,
    quickActions,
    showNoClusterMessage,
    getMetricBarColor,
    getMetricTextColor,
    isOnline
  } = useServerDetailsPanel({
    server,
    onEdit,
    onTestConnection,
    onJoinCluster,
    onLeaveCluster,
    onPromote,
    onDemote,
    isTestingConnection,
    hasCluster
  });

  return (
    <div className="h-full flex flex-col bg-card border-l overflow-hidden shadow-lg">
      <PanelHeader server={server} onClose={onClose} />

      <ScrollArea className="flex-1 overflow-hidden">
        <div className="p-4 space-y-6">
          <ServerBadges
            server={server}
            statusText={t(`servers.status.${server.status}`)}
            healthText={t(`servers.health.${server.health}`)}
          />

          <Separator />

          {server.metrics && isOnline && (
            <>
              <h3 className="text-sm font-medium">{t('servers.details.metrics')}</h3>
              <MetricsSection
                metrics={server.metrics}
                metricRows={metricRows}
                getBarColor={getMetricBarColor}
                getTextColor={getMetricTextColor}
                containerLabel={t('servers.details.containerCount')}
                lastCheckedLabel={t('servers.details.lastChecked')}
              />
            </>
          )}

          {server.status === 'offline' && <OfflineMessage message={t('servers.status.offline')} />}

          <Separator />

          <ClusterActionsSection
            actions={clusterActions}
            showNoClusterMessage={showNoClusterMessage}
            noClusterMessage={t('servers.cluster.noCluster.description')}
            title={t('servers.actions.clusterActions')}
          />

          <Separator />

          <QuickActionsSection actions={quickActions} title={t('servers.actions.quickActions')} />
        </div>
      </ScrollArea>

      <DeleteFooter onDelete={() => onDelete(server)} label={t('servers.deleteServer')} />
    </div>
  );
}
