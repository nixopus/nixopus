'use client';

import { ChevronDown, ChevronRight, Download } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { cn } from '@/lib/utils';
import {
  formatLogMessage,
  formatDataPreview,
  formatVerboseData,
  type FormattedLog
} from './utils/log-formatter';

interface LogEntryProps {
  log: FormattedLog;
  isCollapsed: boolean;
  onToggleCollapse: () => void;
}

export function LogEntry({ log, isCollapsed, onToggleCollapse }: LogEntryProps) {
  const hasVerboseData = log.isVerbose && log.data != null;

  return (
    <div
      className={cn(
        'text-sm border-l-2 pl-3 py-2 transition-colors',
        log.color,
        'bg-muted/30 dark:bg-muted/10 hover:bg-muted/50'
      )}
    >
      <div className="flex items-start gap-2">
        <div className="flex-shrink-0 mt-0.5">{log.icon}</div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-xs text-muted-foreground font-mono">{log.timestamp}</span>
            <Badge variant="outline" className="text-xs px-1.5 py-0">
              {log.level}
            </Badge>
            <span className="font-medium">{formatLogMessage(log.message, log.data)}</span>
          </div>

          {log.progressInfo && <ProgressInfo progressInfo={log.progressInfo} />}

          {log.data != null && !hasVerboseData && (
            <div className="mt-2 text-xs font-mono bg-muted/50 p-2 rounded border break-all">
              {formatDataPreview(log.data)}
            </div>
          )}

          {hasVerboseData && log.data != null && (
            <VerboseDataSection
              data={log.data}
              isCollapsed={isCollapsed}
              onToggle={onToggleCollapse}
            />
          )}
        </div>
      </div>
    </div>
  );
}

interface ProgressInfoProps {
  progressInfo: FormattedLog['progressInfo'];
}

function ProgressInfo({ progressInfo }: ProgressInfoProps) {
  if (!progressInfo) return null;

  return (
    <div className="mt-2 flex items-center gap-2 text-xs bg-blue-50 dark:bg-blue-950/20 p-2 rounded border border-blue-200 dark:border-blue-800">
      <Download className="h-3.5 w-3.5 text-blue-600 dark:text-blue-400" />
      <div className="flex-1">
        <div className="font-medium text-blue-900 dark:text-blue-300">{progressInfo.status}</div>
        {progressInfo.progress && (
          <div className="text-blue-700 dark:text-blue-400 mt-0.5">{progressInfo.progress}</div>
        )}
      </div>
    </div>
  );
}

interface VerboseDataSectionProps {
  data: unknown;
  isCollapsed: boolean;
  onToggle: () => void;
}

function VerboseDataSection({ data, isCollapsed, onToggle }: VerboseDataSectionProps) {
  return (
    <Collapsible open={!isCollapsed} onOpenChange={onToggle}>
      <CollapsibleTrigger className="mt-2 flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground">
        {isCollapsed ? <ChevronRight className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
        <span>{isCollapsed ? 'Show verbose output' : 'Hide verbose output'}</span>
      </CollapsibleTrigger>
      <CollapsibleContent className="mt-2">
        <div className="text-xs font-mono bg-muted/50 p-3 rounded border overflow-x-auto max-h-96 overflow-y-auto">
          <pre className="whitespace-pre-wrap break-all">{formatVerboseData(data)}</pre>
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}
