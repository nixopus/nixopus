'use client';

import React, { useState, useRef } from 'react';
import {
  Layers,
  Calendar,
  HardDrive,
  Tag,
  Copy,
  Check,
  ChevronDown,
  Package,
  Clock,
  Search,
  ChevronsUpDown,
  RefreshCw,
  X,
  ChevronRight,
  Rows3,
  Rows4,
  Loader2,
  Download,
  Box,
  Cpu,
  MemoryStick,
  Network,
  Terminal as TerminalIcon
} from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { formatBytes } from '@/lib/utils';
import { useGetImagesQuery } from '@/redux/services/container/imagesApi';
import { formatDistanceToNow, format } from 'date-fns';
import { cn } from '@/lib/utils';
import { useContainerImages } from '@/packages/hooks/containers/use-container-images';
import {
  ContainerImage,
  ImagesProps,
  StatItemProps,
  ImageCardProps,
  DetailRowProps,
  LogsTabProps,
  LogEntryProps,
  TerminalProps,
  OverviewTabProps,
  StatBlockProps,
  ResourceGaugeProps,
  InfoLineProps
} from '@/packages/types/containers';
import {
  CopyButton,
  EmptyState,
  PortDisplay,
  StatusIndicator,
  ResourceLimitsForm
} from '@/packages/components/container';
import { ImagesSectionSkeleton } from '@/packages/components/container-skeletons';
import { useContainerOverview } from '@/packages/hooks/containers/use-container-overview';
import { textColorClasses, resourceGaugeColors } from '@/packages/utils/container-styles';
import { Container } from '@/redux/services/container/containerApi';
import {
  useContainerLogs,
  levelOptions,
  levelColors,
  ParsedLogEntry
} from '@/packages/hooks/containers/use-container-logs';
import '@xterm/xterm/css/xterm.css';
import { useContainerTerminal } from '@/packages/hooks/terminal/use-container-terminal';

// ============================================================================
// Images Section
// ============================================================================

export function Images({ containerId, imagePrefix }: ImagesProps) {
  const { data: images = [], isLoading } = useGetImagesQuery({ containerId, imagePrefix });
  const { t } = useTranslation();

  if (isLoading) {
    return <ImagesSectionSkeleton />;
  }

  if (images.length === 0) {
    return <EmptyState icon={Layers} message={t('containers.images.none')} />;
  }

  const totalSize = images.reduce((acc, img) => acc + img.size, 0);
  const totalLayers = images.length;

  return (
    <div className="space-y-8">
      <div className="flex items-center gap-8">
        <StatItem icon={Layers} value={totalLayers} label="Image Layers" />
        <StatItem icon={HardDrive} value={formatBytes(totalSize)} label="Total Size" />
      </div>

      <div className="space-y-4">
        {images.map((image, index) => (
          <ImageCard key={image.id} image={image} isFirst={index === 0} />
        ))}
      </div>
    </div>
  );
}

function StatItem({ icon: Icon, value, label }: StatItemProps) {
  return (
    <div className="flex items-center gap-3">
      <div className="p-2 rounded-lg bg-muted/50">
        <Icon className="h-4 w-4 text-muted-foreground" />
      </div>
      <div>
        <p className="text-xl font-bold">{value}</p>
        <p className="text-xs text-muted-foreground">{label}</p>
      </div>
    </div>
  );
}

function ImageCard({ image, isFirst }: ImageCardProps) {
  const [expanded, setExpanded] = useState(isFirst);
  const { copied, copyToClipboard } = useContainerImages();

  const createdDate = new Date(image.created * 1000);
  const primaryTag = image.repo_tags?.[0] || '<none>';
  const hasLabels = image.labels && Object.keys(image.labels).length > 0;

  return (
    <div className="group">
      <button
        onClick={() => setExpanded(!expanded)}
        className={cn(
          'w-full flex items-center gap-4 p-4 rounded-xl transition-colors text-left',
          'hover:bg-muted/30',
          expanded && 'bg-muted/20'
        )}
      >
        <div
          className={cn(
            'p-3 rounded-xl transition-colors',
            isFirst ? 'bg-emerald-500/10' : 'bg-muted/50'
          )}
        >
          <Package
            className={cn('h-5 w-5', isFirst ? 'text-emerald-500' : 'text-muted-foreground')}
          />
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold truncate">{primaryTag}</span>
            {isFirst && (
              <span className="px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                Current
              </span>
            )}
          </div>
          <p className="text-xs text-muted-foreground mt-0.5 font-mono">
            {image.id.replace('sha256:', '').slice(0, 12)}
          </p>
        </div>

        <div className="hidden sm:flex items-center gap-6 text-sm text-muted-foreground">
          <span className="flex items-center gap-1.5">
            <HardDrive className="h-3.5 w-3.5" />
            {formatBytes(image.size)}
          </span>
          <span className="flex items-center gap-1.5">
            <Clock className="h-3.5 w-3.5" />
            {formatDistanceToNow(createdDate, { addSuffix: true })}
          </span>
        </div>

        <ChevronDown
          className={cn(
            'h-4 w-4 text-muted-foreground transition-transform',
            expanded && 'rotate-180'
          )}
        />
      </button>

      {expanded && (
        <div className="px-4 pb-4 pt-2 space-y-4 ml-[60px]">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <DetailRow
              icon={Tag}
              label="Image ID"
              value={image.id.replace('sha256:', '')}
              displayValue={image.id.replace('sha256:', '').slice(0, 24) + '...'}
              mono
              copyable
              onCopy={() => copyToClipboard(image.id, 'id')}
              copied={copied === 'id'}
            />
            <DetailRow icon={Calendar} label="Created" value={format(createdDate, 'PPpp')} />
            <DetailRow
              icon={HardDrive}
              label="Size"
              value={formatBytes(image.size)}
              sublabel={`Virtual: ${formatBytes(image.virtual_size)}`}
            />
            {image.shared_size > 0 && (
              <DetailRow icon={Layers} label="Shared Size" value={formatBytes(image.shared_size)} />
            )}
          </div>

          {image.repo_tags && image.repo_tags.length > 0 && (
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                Tags
              </p>
              <div className="flex flex-wrap gap-2">
                {image.repo_tags.map((tag, idx) => (
                  <div
                    key={idx}
                    className="group/tag flex items-center gap-2 px-3 py-1.5 rounded-lg bg-muted/30 text-sm"
                  >
                    <Tag className="h-3 w-3 text-muted-foreground" />
                    <span className="font-mono">{tag}</span>
                    <div
                      className="opacity-0 group-hover/tag:opacity-100 transition-opacity"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <CopyButton
                        copied={copied === `tag-${idx}`}
                        onCopy={() => copyToClipboard(tag, `tag-${idx}`)}
                        size="sm"
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {image.repo_digests && image.repo_digests.length > 0 && (
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                Digests
              </p>
              <div className="space-y-1">
                {image.repo_digests.map((digest, idx) => (
                  <div
                    key={idx}
                    className="group/digest flex items-center gap-2 p-2 rounded-lg bg-zinc-950 text-zinc-400"
                  >
                    <code className="text-xs font-mono truncate flex-1">{digest}</code>
                    <div
                      className="opacity-0 group-hover/digest:opacity-100 transition-opacity flex-shrink-0"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <CopyButton
                        copied={copied === `digest-${idx}`}
                        onCopy={() => copyToClipboard(digest, `digest-${idx}`)}
                        size="sm"
                        className="text-zinc-500 hover:text-zinc-300"
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {hasLabels && (
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                Labels
              </p>
              <div className="grid grid-cols-1 gap-1">
                {Object.entries(image.labels).map(([key, value]) => (
                  <div key={key} className="flex items-start gap-2 py-1.5 text-sm">
                    <span className="text-muted-foreground font-mono text-xs">{key}:</span>
                    <span className="font-mono text-xs break-all">{value}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function DetailRow({
  icon: Icon,
  label,
  value,
  displayValue,
  sublabel,
  mono,
  copyable,
  onCopy,
  copied
}: DetailRowProps) {
  return (
    <div className="flex items-start gap-2">
      <Icon className="h-4 w-4 mt-0.5 text-muted-foreground flex-shrink-0" />
      <div className="min-w-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        <div className="flex items-center gap-2">
          <span className={cn('text-sm truncate', mono && 'font-mono')} title={value}>
            {displayValue || value}
          </span>
          {copyable && onCopy && (
            <div onClick={(e) => e.stopPropagation()}>
              <CopyButton copied={!!copied} onCopy={onCopy} size="sm" />
            </div>
          )}
        </div>
        {sublabel && <p className="text-xs text-muted-foreground/60">{sublabel}</p>}
      </div>
    </div>
  );
}

// ============================================================================
// Overview Tab
// ============================================================================

export function OverviewTab({ container }: OverviewTabProps) {
  const { copied, showRaw, setShowRaw, copyToClipboard } = useContainerOverview();

  const memory = container.host_config.memory;
  const memoryMB = memory > 0 ? Math.round(memory / (1024 * 1024)) : 0;
  const memorySwap = container.host_config.memory_swap;
  const swapMB = memorySwap > 0 ? Math.round(memorySwap / (1024 * 1024)) : 0;

  return (
    <div className="space-y-10">
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
        <StatBlock
          value={container.status}
          label="Status"
          color={container.status === 'running' ? 'emerald' : 'red'}
          pulse={container.status === 'running'}
        />
        <StatBlock
          value={memoryMB > 0 ? `${memoryMB} MB` : '∞'}
          label="Memory Limit"
          sublabel={memoryMB === 0 ? 'Unlimited' : undefined}
        />
        <StatBlock value={container.host_config.cpu_shares.toString()} label="CPU Shares" />
        <StatBlock
          value={container.ports?.length || 0}
          label="Exposed Ports"
          sublabel={container.ports?.length ? 'configured' : 'none'}
        />
      </div>

      <section className="space-y-8">
        <div className="flex items-center gap-4">
          <SectionLabel>Resource Allocation</SectionLabel>
          <ResourceLimitsForm container={container} />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <ResourceGauge
            icon={MemoryStick}
            label="Memory"
            value={memoryMB}
            maxLabel={memoryMB > 0 ? `${memoryMB} MB` : 'No Limit'}
            color="blue"
            unlimited={memoryMB === 0}
          />
          <ResourceGauge
            icon={HardDrive}
            label="Swap"
            value={swapMB}
            maxLabel={swapMB > 0 ? `${swapMB} MB` : 'No Limit'}
            color="purple"
            unlimited={swapMB === 0}
          />
          <ResourceGauge
            icon={Cpu}
            label="CPU Shares"
            value={container.host_config.cpu_shares}
            maxLabel={`${container.host_config.cpu_shares} shares`}
            color="amber"
            showBar={false}
          />
        </div>
      </section>

      {container.ports && container.ports.length > 0 && (
        <section className="space-y-4">
          <SectionLabel>Network Configuration</SectionLabel>
          <div className="flex flex-wrap gap-4">
            {container.ports.map((port, idx) => (
              <PortDisplay key={idx} port={port} variant="flow" />
            ))}
          </div>
        </section>
      )}

      <section className="space-y-8">
        <SectionLabel>Container Identity</SectionLabel>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-y-4 gap-x-12">
          <InfoLine
            icon={Box}
            label="Image"
            value={container.image}
            copyable
            onCopy={() => copyToClipboard(container.image, 'image')}
            copied={copied === 'image'}
          />
          <InfoLine
            icon={HardDrive}
            label="Container ID"
            value={container.id}
            displayValue={container.id.slice(0, 12) + '...'}
            mono
            copyable
            onCopy={() => copyToClipboard(container.id, 'id')}
            copied={copied === 'id'}
          />
          <InfoLine
            icon={Network}
            label="IP Address"
            value={container.ip_address || 'Not assigned'}
            mono
          />
          <InfoLine
            icon={Clock}
            label="Created"
            value={formatDistanceToNow(new Date(container.created), { addSuffix: true })}
            sublabel={format(new Date(container.created), 'PPpp')}
          />
        </div>
      </section>

      {container.command && (
        <section className="space-y-8">
          <SectionLabel>Entrypoint</SectionLabel>
          <div className="relative group">
            <div className="flex items-start gap-3 p-4 rounded-xl bg-zinc-950 text-zinc-300">
              <TerminalIcon className="h-4 w-4 mt-0.5 text-zinc-500 flex-shrink-0" />
              <code className="text-sm font-mono break-all">{container.command}</code>
            </div>
            <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
              <CopyButton
                copied={copied === 'cmd'}
                onCopy={() => copyToClipboard(container.command, 'cmd')}
                size="sm"
                className="text-zinc-400 hover:text-zinc-200"
              />
            </div>
          </div>
        </section>
      )}

      <section className="pt-4">
        <button
          onClick={() => setShowRaw(!showRaw)}
          className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          <ChevronDown className={cn('h-4 w-4 transition-transform', showRaw && 'rotate-180')} />
          <span>Raw inspection data</span>
        </button>
        {showRaw && (
          <div className="mt-4 relative group">
            <div className="absolute top-3 right-3 z-10">
              <CopyButton
                copied={copied === 'raw'}
                onCopy={() => copyToClipboard(JSON.stringify(container, null, 2), 'raw')}
                size="sm"
                showText
                className="text-zinc-400"
              />
            </div>
            <pre className="p-4 rounded-xl bg-zinc-950 text-zinc-400 text-xs font-mono overflow-auto max-h-[400px] no-scrollbar">
              {JSON.stringify(container, null, 2)}
            </pre>
          </div>
        )}
      </section>
    </div>
  );
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
      {children}
    </h3>
  );
}

function StatBlock({ value, label, sublabel, color, pulse }: StatBlockProps) {
  return (
    <div className="relative">
      <div className="space-y-1">
        <div className="flex items-center gap-2">
          {pulse && <StatusIndicator isRunning={true} size="md" />}
          <span
            className={cn(
              'text-2xl font-bold tracking-tight capitalize',
              color && textColorClasses[color]
            )}
          >
            {value}
          </span>
        </div>
        <p className="text-sm text-muted-foreground">{label}</p>
        {sublabel && <p className="text-xs text-muted-foreground/60">{sublabel}</p>}
      </div>
    </div>
  );
}

function ResourceGauge({
  icon: Icon,
  label,
  value,
  maxLabel,
  color,
  unlimited,
  showBar = true
}: ResourceGaugeProps) {
  const colors = resourceGaugeColors[color];

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <div className={cn('p-2 rounded-lg', colors.track)}>
          <Icon className={cn('h-4 w-4', colors.text)} />
        </div>
        <span className="text-sm font-medium">{label}</span>
      </div>

      {showBar && !unlimited && (
        <div className={cn('h-2 rounded-full overflow-hidden', colors.track)}>
          <div className={cn('h-full rounded-full', colors.bg)} style={{ width: '70%' }} />
        </div>
      )}

      {unlimited ? (
        <div className="flex items-center gap-2">
          <span className="text-2xl font-bold">∞</span>
          <span className="text-sm text-muted-foreground">Unlimited</span>
        </div>
      ) : (
        <p className="text-lg font-semibold">{maxLabel}</p>
      )}
    </div>
  );
}

function InfoLine({
  icon: Icon,
  label,
  value,
  displayValue,
  sublabel,
  mono,
  copyable,
  onCopy,
  copied
}: InfoLineProps) {
  return (
    <div className="flex items-start gap-3 py-2">
      <Icon className="h-4 w-4 mt-1 text-muted-foreground flex-shrink-0" />
      <div className="flex-1 min-w-0">
        <p className="text-xs text-muted-foreground uppercase tracking-wide mb-0.5">{label}</p>
        <div className="flex items-center gap-2">
          <span className={cn('text-sm truncate', mono && 'font-mono')} title={value}>
            {displayValue || value}
          </span>
          {copyable && onCopy && <CopyButton copied={!!copied} onCopy={onCopy} size="sm" />}
        </div>
        {sublabel && <p className="text-xs text-muted-foreground/60 mt-0.5">{sublabel}</p>}
      </div>
    </div>
  );
}

// ============================================================================
// Logs Tab
// ============================================================================

export function LogsTab({ container, logs, onLoadMore, onRefresh }: LogsTabProps) {
  const { t } = useTranslation();
  const {
    parsedLogs,
    searchTerm,
    setSearchTerm,
    levelFilter,
    setLevelFilter,
    isDense,
    setIsDense,
    isLoadingMore,
    setIsLoadingMore,
    isRefreshing,
    setIsRefreshing,
    isCopied,
    allExpanded,
    handleCopyLogs,
    handleDownloadLogs,
    toggleLogExpansion,
    isLogExpanded,
    handleExpandCollapseToggle,
    clearFilters,
    hasActiveFilters
  } = useContainerLogs(logs, container.name);

  const handleLoadMore = async () => {
    setIsLoadingMore(true);
    await onLoadMore();
    setIsLoadingMore(false);
  };

  const handleRefresh = async () => {
    if (!onRefresh) return;
    setIsRefreshing(true);
    try {
      await onRefresh();
    } finally {
      setIsRefreshing(false);
    }
  };

  return (
    <div className="space-y-4 overflow-x-hidden">
      <div className="flex flex-col gap-3 min-w-0">
        <div className="flex items-center gap-3 min-w-0">
          <div className="relative flex-1 max-w-md min-w-0">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search logs..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10 bg-transparent"
            />
          </div>
          <div className="flex items-center gap-2 ml-auto flex-shrink-0">
            {hasActiveFilters && (
              <Button
                variant="ghost"
                size="sm"
                onClick={clearFilters}
                className="text-muted-foreground h-9"
              >
                <X className="h-4 w-4 mr-1" />
                Clear
              </Button>
            )}
            {onRefresh && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleRefresh}
                disabled={isRefreshing}
                className="h-9"
                title="Refresh logs"
              >
                {isRefreshing ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <RefreshCw className="h-4 w-4 mr-2" />
                )}
                Refresh
              </Button>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={handleLoadMore}
              disabled={isLoadingMore}
              className="h-9"
            >
              {isLoadingMore ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : (
                <RefreshCw className="h-4 w-4 mr-2" />
              )}
              Load more
            </Button>
          </div>
        </div>

        <div className="flex items-center justify-between min-w-0 gap-2">
          <div className="flex items-center gap-1.5 flex-wrap min-w-0">
            {levelOptions.map((option) => (
              <button
                key={option.value}
                onClick={() => setLevelFilter(option.value)}
                className={cn(
                  'px-3 py-1 text-xs font-medium rounded-full transition-colors flex-shrink-0',
                  levelFilter === option.value
                    ? 'bg-foreground text-background'
                    : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                )}
              >
                {option.label}
              </button>
            ))}
          </div>
          <div className="flex items-center gap-1 flex-shrink-0">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleCopyLogs}
              className="h-8 px-2 text-muted-foreground"
              title={t('containers.logs.copy')}
              disabled={parsedLogs.length === 0}
            >
              {isCopied ? (
                <Check className="h-4 w-4 text-green-500" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleDownloadLogs}
              className="h-8 px-2 text-muted-foreground"
              title={t('containers.logs.download')}
              disabled={parsedLogs.length === 0}
            >
              <Download className="h-4 w-4" />
            </Button>
            <div className="w-px h-4 bg-border mx-1" />
            <Button
              variant="ghost"
              size="sm"
              onClick={handleExpandCollapseToggle}
              className="h-8 px-2 text-muted-foreground"
              title={allExpanded ? 'Collapse all' : 'Expand all'}
            >
              <ChevronsUpDown className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsDense(!isDense)}
              className="h-8 px-2 text-muted-foreground"
              title={isDense ? 'Normal view' : 'Dense view'}
            >
              {isDense ? <Rows3 className="h-4 w-4" /> : <Rows4 className="h-4 w-4" />}
            </Button>
          </div>
        </div>
      </div>

      <div className="rounded-lg border overflow-hidden bg-zinc-950 min-w-0">
        <div className="flex items-center gap-3 px-4 py-2 text-xs font-medium text-zinc-500 uppercase tracking-wider border-b border-zinc-800 min-w-0">
          <div className="w-4 flex-shrink-0" />
          <div className="w-14 flex-shrink-0">Level</div>
          <div className="w-40 flex-shrink-0">Time</div>
          <div className="flex-1 min-w-0">Message</div>
        </div>

        {parsedLogs.length === 0 ? (
          <div className="py-16 text-center text-zinc-500">
            <p className="text-sm">{t('containers.no_logs')}</p>
          </div>
        ) : (
          <div className="max-h-[600px] overflow-y-auto overflow-x-hidden">
            {parsedLogs.map((log) => (
              <LogEntry
                key={log.id}
                log={log}
                isExpanded={isLogExpanded(log.id)}
                onToggle={() => toggleLogExpansion(log.id)}
                isDense={isDense}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function LogEntry({ log, isExpanded, onToggle, isDense }: LogEntryProps) {
  const colors = levelColors[log.level];

  return (
    <div className="border-b border-zinc-800/50 last:border-0 min-w-0">
      <div
        className={cn(
          'flex items-start gap-3 px-4 cursor-pointer transition-colors min-w-0',
          'hover:bg-zinc-900/50',
          isExpanded && 'bg-zinc-900/30',
          isDense ? 'py-1' : 'py-2.5'
        )}
        onClick={onToggle}
      >
        <ChevronRight
          className={cn(
            'h-3.5 w-3.5 text-zinc-500 transition-transform duration-150 flex-shrink-0 mt-0.5',
            isExpanded && 'rotate-90'
          )}
        />
        <span
          className={cn(
            'px-2 py-0.5 rounded text-[10px] font-semibold uppercase w-14 text-center flex-shrink-0',
            colors.bg,
            colors.text
          )}
        >
          {log.level}
        </span>
        <span
          className={cn(
            'w-40 font-mono text-zinc-500 flex-shrink-0',
            isDense ? 'text-[11px]' : 'text-xs'
          )}
        >
          {log.formattedTime}
        </span>
        <span
          className={cn(
            'flex-1 text-zinc-300 break-words line-clamp-1 overflow-hidden min-w-0',
            isDense ? 'text-xs' : 'text-sm'
          )}
        >
          {log.message}
        </span>
      </div>

      {isExpanded && (
        <div className="px-4 pb-3 pt-1 min-w-0 overflow-x-hidden">
          <pre
            className={cn(
              'ml-8 p-3 rounded bg-zinc-900 text-zinc-300 font-mono whitespace-pre-wrap break-words overflow-wrap-anywhere',
              isDense ? 'text-[11px]' : 'text-xs'
            )}
          >
            {log.raw}
          </pre>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Terminal
// ============================================================================

export const Terminal: React.FC<TerminalProps> = ({ containerId }) => {
  const terminalRef = useRef<HTMLDivElement | null>(null);
  const { terminalRef: termRef } = useContainerTerminal(containerId);

  return (
    <div
      ref={(el) => {
        terminalRef.current = el;
        // @ts-ignore
        if (termRef) termRef.current = el;
      }}
      className="relative"
      style={{ height: '60vh', minHeight: 300, backgroundColor: '#1e1e1e' }}
    />
  );
};

export default Terminal;
