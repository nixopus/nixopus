import React from 'react';
import { CheckCircle2, Clock, Loader2, XCircle } from 'lucide-react';
import type { ExtensionLog } from '@/redux/types/extension';

export interface FormattedLog {
  id: string;
  timestamp: string;
  level: string;
  message: string;
  icon?: React.ReactNode;
  color: string;
  data?: unknown;
  isVerbose?: boolean;
  progressInfo?: {
    progress?: string;
    status?: string;
    id?: string;
  };
}

function extractDockerProgress(dataStr: string): FormattedLog['progressInfo'] | undefined {
  try {
    const lines = dataStr.split(/[\r\n]+/).filter((l) => l.trim());
    let lastProgress: string | undefined;
    let lastStatus: string | undefined;

    for (let i = lines.length - 1; i >= 0; i--) {
      const line = lines[i].trim();
      if (!line) continue;

      try {
        const parsed = JSON.parse(line);
        if (
          parsed.status &&
          (parsed.status.includes('Downloading') ||
            parsed.status.includes('Extracting') ||
            parsed.status.includes('Pulling'))
        ) {
          if (parsed.progress) {
            lastProgress = parsed.progress;
          }
          if (parsed.status) {
            lastStatus = parsed.status;
          }
          if (lastProgress && lastStatus) break;
        }
      } catch {
        continue;
      }
    }

    if (lastProgress || lastStatus) {
      return {
        progress: lastProgress || '',
        status: lastStatus || 'Processing'
      };
    }
  } catch {}

  return undefined;
}

export function formatLog(
  log: ExtensionLog,
  isStepCompleted?: boolean,
  isStepFailed?: boolean
): FormattedLog {
  const timestamp = new Date(log.created_at).toLocaleTimeString();
  const level = log.level.toUpperCase();

  let icon: React.ReactNode | undefined;
  let color = 'text-muted-foreground';
  let isVerbose = false;
  let progressInfo: FormattedLog['progressInfo'] | undefined;

  if (log.message === 'execution_started') {
    icon = <Clock className="h-4 w-4" />;
    color = 'text-blue-600 dark:text-blue-400';
  } else if (log.message === 'execution_completed') {
    icon = <CheckCircle2 className="h-4 w-4" />;
    color = 'text-green-600 dark:text-green-400';
  } else if (log.message.startsWith('step_started')) {
    if (isStepFailed) {
      icon = <XCircle className="h-4 w-4" />;
      color = 'text-red-600 dark:text-red-400';
    } else if (isStepCompleted) {
      icon = <CheckCircle2 className="h-4 w-4" />;
      color = 'text-green-600 dark:text-green-400';
    } else {
      icon = <Loader2 className="h-4 w-4 animate-spin" />;
      color = 'text-blue-600 dark:text-blue-400';
    }
  } else if (log.message.startsWith('step_completed')) {
    icon = <CheckCircle2 className="h-4 w-4" />;
    color = 'text-green-600 dark:text-green-400';
  } else if (log.message.startsWith('step_failed')) {
    icon = <XCircle className="h-4 w-4" />;
    color = 'text-red-600 dark:text-red-400';
  } else if (log.level === 'error' || log.level === 'ERROR') {
    icon = <XCircle className="h-4 w-4" />;
    color = 'text-red-600 dark:text-red-400';
  }

  if (log.data) {
    const dataStr = typeof log.data === 'string' ? log.data : JSON.stringify(log.data);
    const isLargeData = dataStr.length > 5000;
    const isDockerProgress =
      typeof log.data === 'string' &&
      (log.data.includes('{"status":"Downloading"') ||
        log.data.includes('"status":"Pulling') ||
        log.data.includes('"status":"Extracting') ||
        log.data.includes('"status":"Verifying'));

    if (isLargeData || isDockerProgress) {
      isVerbose = true;

      if (isDockerProgress) {
        progressInfo = extractDockerProgress(dataStr);
      }
    }
  }

  return {
    id: log.id,
    timestamp,
    level,
    message: log.message,
    icon,
    color,
    data: log.data,
    isVerbose,
    progressInfo
  };
}

export function formatLogMessage(message: string, data?: unknown): string {
  if (!data) {
    if (message === 'execution_started') return 'Execution started';
    if (message === 'execution_completed') return 'Execution completed';
    return message;
  }

  try {
    const parsed = typeof data === 'string' ? JSON.parse(data) : data;

    if (message === 'step_started' && parsed.step_name) {
      const phase = parsed.phase ? ` (${parsed.phase})` : '';
      const order = parsed.order ? ` #${parsed.order}` : '';
      return `Starting: ${parsed.step_name}${phase}${order}`;
    }

    if (message === 'step_completed' && parsed.step_name) {
      return `Completed: ${parsed.step_name}`;
    }

    if (message === 'step_failed' && parsed.step_name) {
      return `Failed: ${parsed.step_name}`;
    }

    if (message.includes('Check') && parsed.output) {
      const output = typeof parsed.output === 'string' ? parsed.output.trim() : '';
      if (output) {
        return `${message}: ${output.split('\n')[0].substring(0, 80)}`;
      }
    }

    return message;
  } catch {
    return message;
  }
}

export function formatDataPreview(data: unknown): string {
  if (!data) return '';

  const dataStr = typeof data === 'string' ? data : JSON.stringify(data, null, 2);

  if (dataStr.length > 200) {
    return dataStr.substring(0, 200) + '...';
  }

  return dataStr;
}

export function formatVerboseData(data: unknown): string {
  if (!data) return '';

  return typeof data === 'string' ? data : JSON.stringify(data ?? null, null, 2);
}
