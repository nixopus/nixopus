'use client';

import { Box, Clock, Network, Globe, Lock, ArrowRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ContainerActions } from './actions';
import { formatDistanceToNow } from 'date-fns';
import { Container } from '@/redux/services/container/containerApi';

interface ContainerCardProps {
  container: Container;
  onClick: () => void;
  getGradientFromName: (name: string) => string;
  onAction: (id: string, action: Action) => void;
}

export enum Action {
  START = 'start',
  STOP = 'stop',
  REMOVE = 'remove'
}

export const ContainerCard = ({ container, onClick, onAction }: ContainerCardProps) => {
  const isRunning = container.status === 'running';
  const hasPorts = container.ports && container.ports.length > 0;

  return (
    <div
      onClick={onClick}
      className={cn(
        'group relative rounded-xl p-5 cursor-pointer transition-all duration-200 bg-muted/50',
        'hover:bg-muted/70',
        'border border-transparent hover:border-border/50'
      )}
    >
      {/* Status Indicator */}
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-center gap-3 min-w-0 flex-1">
          <div
            className={cn(
              'p-2.5 rounded-xl flex-shrink-0',
              isRunning ? 'bg-emerald-500/10' : 'bg-zinc-500/10'
            )}
          >
            <Box className={cn('h-5 w-5', isRunning ? 'text-emerald-500' : 'text-zinc-500')} />
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <h3 className="font-semibold truncate">{container.name}</h3>
              {isRunning && (
                <span className="relative flex h-2 w-2 flex-shrink-0">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
                </span>
              )}
            </div>
            <p className="text-xs text-muted-foreground truncate mt-0.5 font-mono">
              {container.id.slice(0, 12)}
            </p>
          </div>
        </div>

        <div
          className="flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={(e) => e.stopPropagation()}
        >
          <ContainerActions container={container} onAction={onAction} />
        </div>
      </div>

      {/* Image */}
      <div className="mt-4 flex items-center gap-2 text-sm text-muted-foreground">
        <span className="truncate" title={container.image}>
          {container.image}
        </span>
      </div>

      {/* Bottom Row - Ports & Time */}
      <div className="mt-4 flex items-center justify-between gap-4">
        {/* Ports */}
        <div className="flex items-center gap-2 min-w-0 flex-1">
          {hasPorts ? (
            <div className="flex items-center gap-1.5 overflow-hidden">
              <Network className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
              <div className="flex items-center gap-1 overflow-hidden">
                {container.ports.slice(0, 2).map((port, idx) => (
                  <PortPill key={idx} port={port} />
                ))}
                {container.ports.length > 2 && (
                  <span className="text-xs text-muted-foreground">
                    +{container.ports.length - 2}
                  </span>
                )}
              </div>
            </div>
          ) : (
            <span className="text-xs text-muted-foreground/50">No ports</span>
          )}
        </div>

        {/* Created */}
        <div className="flex items-center gap-1.5 text-xs text-muted-foreground flex-shrink-0">
          <Clock className="h-3 w-3" />
          <span>{formatDistanceToNow(new Date(container.created), { addSuffix: true })}</span>
        </div>
      </div>
    </div>
  );
};

function PortPill({ port }: { port: { private_port: number; public_port: number; type: string } }) {
  const hasPublic = port.public_port > 0;

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-mono',
        hasPublic
          ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
          : 'bg-muted text-muted-foreground'
      )}
    >
      {hasPublic ? (
        <>
          <span>{port.public_port}</span>
          <ArrowRight className="h-2.5 w-2.5" />
          <span>{port.private_port}</span>
        </>
      ) : (
        <span>{port.private_port}</span>
      )}
    </span>
  );
}
