'use client';

import { useState } from 'react';
import {
  Copy,
  Check,
  ChevronDown,
  ArrowRight,
  Globe,
  Lock,
  Clock,
  Box,
  Cpu,
  MemoryStick,
  HardDrive,
  Network,
  Terminal as TerminalIcon
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { formatDistanceToNow, format } from 'date-fns';
import { Container } from '@/redux/services/container/containerApi';
import { cn } from '@/lib/utils';
import { ResourceLimitsForm } from './ResourceLimitsForm';

interface OverviewTabProps {
  container: Container;
}

export function OverviewTab({ container }: OverviewTabProps) {
  const [copied, setCopied] = useState<string | null>(null);
  const [showRaw, setShowRaw] = useState(false);

  const copyToClipboard = (text: string, key: string) => {
    navigator.clipboard.writeText(text);
    setCopied(key);
    setTimeout(() => setCopied(null), 2000);
  };

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
              <PortFlow key={idx} port={port} />
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
            <Button
              variant="ghost"
              size="sm"
              className="absolute top-2 right-2 h-7 px-2 opacity-0 group-hover:opacity-100 transition-opacity text-zinc-400 hover:text-zinc-200"
              onClick={() => copyToClipboard(container.command, 'cmd')}
            >
              {copied === 'cmd' ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
            </Button>
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
            <Button
              variant="ghost"
              size="sm"
              className="absolute top-3 right-3 z-10 h-7 text-zinc-400"
              onClick={() => copyToClipboard(JSON.stringify(container, null, 2), 'raw')}
            >
              {copied === 'raw' ? (
                <Check className="h-3 w-3 mr-1" />
              ) : (
                <Copy className="h-3 w-3 mr-1" />
              )}
              {copied === 'raw' ? 'Copied' : 'Copy'}
            </Button>
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

function StatBlock({
  value,
  label,
  sublabel,
  color,
  pulse
}: {
  value: string | number;
  label: string;
  sublabel?: string;
  color?: 'emerald' | 'red' | 'amber' | 'blue';
  pulse?: boolean;
}) {
  const colorClasses = {
    emerald: 'text-emerald-500',
    red: 'text-red-500',
    amber: 'text-amber-500',
    blue: 'text-blue-500'
  };

  return (
    <div className="relative">
      <div className="space-y-1">
        <div className="flex items-center gap-2">
          {pulse && (
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
            </span>
          )}
          <span
            className={cn(
              'text-2xl font-bold tracking-tight capitalize',
              color && colorClasses[color]
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
}: {
  icon: React.ElementType;
  label: string;
  value: number;
  maxLabel: string;
  color: 'blue' | 'purple' | 'amber' | 'emerald';
  unlimited?: boolean;
  showBar?: boolean;
}) {
  const colorClasses = {
    blue: { bg: 'bg-blue-500', track: 'bg-blue-500/20', text: 'text-blue-500' },
    purple: { bg: 'bg-purple-500', track: 'bg-purple-500/20', text: 'text-purple-500' },
    amber: { bg: 'bg-amber-500', track: 'bg-amber-500/20', text: 'text-amber-500' },
    emerald: { bg: 'bg-emerald-500', track: 'bg-emerald-500/20', text: 'text-emerald-500' }
  };

  const colors = colorClasses[color];

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

function PortFlow({ port }: { port: { private_port: number; public_port: number; type: string } }) {
  const hasPublic = port.public_port > 0;

  return (
    <div
      className={cn(
        'flex items-center gap-3 px-4 py-3 rounded-xl transition-colors',
        hasPublic ? 'bg-emerald-500/5' : 'bg-zinc-500/5'
      )}
    >
      {hasPublic ? (
        <>
          <div className="flex items-center gap-2">
            <Globe className="h-4 w-4 text-emerald-500" />
            <span className="font-mono text-lg font-semibold text-emerald-600 dark:text-emerald-400">
              {port.public_port}
            </span>
          </div>
          <ArrowRight className="h-4 w-4 text-muted-foreground" />
          <div className="flex items-center gap-2">
            <Lock className="h-4 w-4 text-muted-foreground" />
            <span className="font-mono text-lg">{port.private_port}</span>
          </div>
        </>
      ) : (
        <div className="flex items-center gap-2">
          <Lock className="h-4 w-4 text-muted-foreground" />
          <span className="font-mono text-lg">{port.private_port}</span>
        </div>
      )}
      <span className="text-xs text-muted-foreground uppercase ml-1">/{port.type}</span>
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
}: {
  icon: React.ElementType;
  label: string;
  value: string;
  displayValue?: string;
  sublabel?: string;
  mono?: boolean;
  copyable?: boolean;
  onCopy?: () => void;
  copied?: boolean;
}) {
  return (
    <div className="flex items-start gap-3 py-2">
      <Icon className="h-4 w-4 mt-1 text-muted-foreground flex-shrink-0" />
      <div className="flex-1 min-w-0">
        <p className="text-xs text-muted-foreground uppercase tracking-wide mb-0.5">{label}</p>
        <div className="flex items-center gap-2">
          <span className={cn('text-sm truncate', mono && 'font-mono')} title={value}>
            {displayValue || value}
          </span>
          {copyable && onCopy && (
            <button
              onClick={onCopy}
              className="text-muted-foreground hover:text-foreground transition-colors flex-shrink-0"
            >
              {copied ? (
                <Check className="h-3 w-3 text-emerald-500" />
              ) : (
                <Copy className="h-3 w-3" />
              )}
            </button>
          )}
        </div>
        {sublabel && <p className="text-xs text-muted-foreground/60 mt-0.5">{sublabel}</p>}
      </div>
    </div>
  );
}
