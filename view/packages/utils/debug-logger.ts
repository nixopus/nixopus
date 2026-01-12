import { getAdvancedSettings } from '@/packages/utils/advanced-settings';

type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface DebugLoggerOptions {
  prefix?: string;
}

class DebugLogger {
  private prefix: string;

  constructor(options: DebugLoggerOptions = {}) {
    this.prefix = options.prefix ? `[${options.prefix}]` : '[Nixopus]';
  }

  private isDebugEnabled(): boolean {
    try {
      return getAdvancedSettings().debugMode;
    } catch {
      return false;
    }
  }

  private log(level: LogLevel, ...args: unknown[]): void {
    if (!this.isDebugEnabled() && level === 'debug') return;

    const timestamp = new Date().toISOString();
    const formattedPrefix = `${this.prefix} ${timestamp}`;

    switch (level) {
      case 'debug':
        console.debug(formattedPrefix, ...args);
        break;
      case 'info':
        console.info(formattedPrefix, ...args);
        break;
      case 'warn':
        console.warn(formattedPrefix, ...args);
        break;
      case 'error':
        console.error(formattedPrefix, ...args);
        break;
    }
  }

  debug(...args: unknown[]): void {
    this.log('debug', ...args);
  }

  info(...args: unknown[]): void {
    this.log('info', ...args);
  }

  warn(...args: unknown[]): void {
    this.log('warn', ...args);
  }

  error(...args: unknown[]): void {
    this.log('error', ...args);
  }
}

export const debugLogger = new DebugLogger();

export function createDebugLogger(prefix: string): DebugLogger {
  return new DebugLogger({ prefix });
}
