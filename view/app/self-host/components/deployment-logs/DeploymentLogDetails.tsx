'use client';

import React from 'react';
import { cn } from '@/lib/utils';
import { FormattedLogEntry } from '../../hooks/use_deployment_logs_viewer';

interface DeploymentLogDetailsProps {
  log: FormattedLogEntry;
  isDense?: boolean;
}

export function DeploymentLogDetails({ log, isDense }: DeploymentLogDetailsProps) {
  return (
    <div
      className={cn(
        'px-4 bg-muted/20 border-b border-border/50 min-w-0 overflow-x-hidden',
        isDense ? 'pb-2 pt-1' : 'pb-4 pt-2'
      )}
    >
      <div className="ml-7 min-w-0">
        <div
          className={cn(
            'font-mono bg-muted/50 rounded border break-words whitespace-pre-wrap overflow-wrap-anywhere',
            isDense ? 'text-xs p-2' : 'text-sm p-3'
          )}
        >
          {log.message}
        </div>
      </div>
    </div>
  );
}

export default DeploymentLogDetails;
