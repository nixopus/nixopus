'use client';

import React from 'react';
import { ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { FormattedLogEntry, LogLevel } from '../../hooks/use_deployment_logs_viewer';

interface DeploymentLogRowProps {
  log: FormattedLogEntry;
  isExpanded: boolean;
  onToggle: () => void;
  isDense?: boolean;
}

export function DeploymentLogRow({ log, isExpanded, onToggle, isDense }: DeploymentLogRowProps) {
  return (
    <div
      className={cn(
        'flex items-start gap-3 px-4 cursor-pointer transition-colors min-w-0',
        'hover:bg-muted/50 border-b border-border/50',
        isExpanded && 'bg-muted/30',
        isDense ? 'py-1' : 'py-3'
      )}
      onClick={onToggle}
    >
      <ChevronIcon isExpanded={isExpanded} isDense={isDense} />
      <LevelBadge level={log.level} isDense={isDense} />
      <TimestampCell timestamp={log.formattedTime} isDense={isDense} />
      <MessageCell message={log.message} isExpanded={isExpanded} isDense={isDense} />
    </div>
  );
}

function ChevronIcon({ isExpanded, isDense }: { isExpanded: boolean; isDense?: boolean }) {
  return (
    <div className="flex-shrink-0 mt-0.5">
      <ChevronRight
        className={cn(
          'text-muted-foreground transition-transform duration-200',
          isExpanded && 'rotate-90',
          isDense ? 'h-3 w-3' : 'h-4 w-4'
        )}
      />
    </div>
  );
}

const levelStyles: Record<LogLevel, string> = {
  error: 'bg-red-500/10 text-red-500 border-red-500/20',
  warn: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
  info: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
  debug: 'bg-gray-500/10 text-gray-500 border-gray-500/20'
};

function LevelBadge({ level, isDense }: { level: LogLevel; isDense?: boolean }) {
  return (
    <Badge
      variant="outline"
      className={cn(
        'justify-center flex-shrink-0',
        levelStyles[level],
        isDense ? 'text-[10px] px-1 py-0 w-12' : 'text-xs px-2 py-0 w-14'
      )}
    >
      {level.toUpperCase()}
    </Badge>
  );
}

function TimestampCell({ timestamp, isDense }: { timestamp: string; isDense?: boolean }) {
  return (
    <div className={cn('flex-shrink-0', isDense ? 'w-40' : 'w-44')}>
      <span className={cn('font-mono text-muted-foreground', isDense ? 'text-xs' : 'text-sm')}>
        {timestamp}
      </span>
    </div>
  );
}

function MessageCell({
  message,
  isExpanded,
  isDense
}: {
  message: string;
  isExpanded: boolean;
  isDense?: boolean;
}) {
  return (
    <div className="flex-1 min-w-0">
      <span
        className={cn(
          isDense ? 'text-xs' : 'text-sm',
          isExpanded
            ? 'whitespace-pre-wrap break-words'
            : 'break-words line-clamp-1 overflow-hidden'
        )}
      >
        {message}
      </span>
    </div>
  );
}

export default DeploymentLogRow;
