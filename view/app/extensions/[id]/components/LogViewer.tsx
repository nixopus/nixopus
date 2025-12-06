'use client';

import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useTranslation } from '@/hooks/use-translation';
import { LogEntry } from './LogEntry';
import type { ExtensionLog } from '@/redux/types/extension';
import { useLogViewer } from '../hooks/use-log-viewer';

interface LogViewerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  executionId: string | null;
  logs: ExtensionLog[];
}

export function LogViewer({ open, onOpenChange, executionId, logs }: LogViewerProps) {
  const { t } = useTranslation();

  const { formattedLogs, collapsedLogs, toggleCollapse, logsEndRef, isEmpty } = useLogViewer({
    open,
    executionId,
    logs
  });

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="sm:max-w-3xl">
        <SheetHeader>
          <SheetTitle>{t('extensions.logs') || 'Execution Logs'}</SheetTitle>
        </SheetHeader>
        <div className="flex flex-col h-[calc(100vh-120px)] mt-4">
          <div className="mb-3 text-xs text-muted-foreground px-1">
            {t('extensions.executionId') || 'Execution ID'}:{' '}
            <span className="font-mono">{executionId}</span>
          </div>
          <div className="flex-1 overflow-y-auto space-y-1 pr-2 min-h-0">
            {isEmpty ? (
              <div className="text-sm text-muted-foreground text-center py-8">No logs yet...</div>
            ) : (
              formattedLogs.map((log) => (
                <LogEntry
                  key={log.id}
                  log={log}
                  isCollapsed={collapsedLogs.has(log.id)}
                  onToggleCollapse={() => toggleCollapse(log.id)}
                />
              ))
            )}
            <div ref={logsEndRef} />
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
