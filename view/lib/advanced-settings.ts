export interface AdvancedSettings {
  websocketReconnectAttempts: number;
  websocketReconnectInterval: number;
  apiRetryAttempts: number;
  disableApiCache: boolean;

  debugMode: boolean;
  showApiErrorDetails: boolean;
  terminalScrollback: number;
  terminalFontSize: number;
  terminalCursorStyle: 'bar' | 'block' | 'underline';
  terminalCursorBlink: boolean;
  terminalLineHeight: number;
  terminalCursorWidth: number;
  terminalTabStopWidth: number;
  terminalFontFamily: string;
  terminalFontWeight: 'normal' | 'bold';
  terminalLetterSpacing: number;

  containerLogTailLines: number;
  containerDefaultRestartPolicy: 'no' | 'always' | 'on-failure' | 'unless-stopped';
  containerStopTimeout: number;
  containerAutoPruneDanglingImages: boolean;
  containerAutoPruneBuildCache: boolean;
}

const STORAGE_KEY = 'nixopus-advanced-settings';

export const DEFAULT_SETTINGS: AdvancedSettings = {
  websocketReconnectAttempts: 5,
  websocketReconnectInterval: 3000,
  apiRetryAttempts: 1,
  disableApiCache: false,

  debugMode: false,
  showApiErrorDetails: false,

  terminalScrollback: 5000,
  terminalFontSize: 13,
  terminalCursorStyle: 'bar',
  terminalCursorBlink: true,
  terminalLineHeight: 1.4,
  terminalCursorWidth: 2,
  terminalTabStopWidth: 4,
  terminalFontFamily: 'JetBrains Mono',
  terminalFontWeight: 'normal',
  terminalLetterSpacing: 0,

  containerLogTailLines: 100,
  containerDefaultRestartPolicy: 'unless-stopped',
  containerStopTimeout: 10,
  containerAutoPruneDanglingImages: false,
  containerAutoPruneBuildCache: false
};

function loadSettingsFromStorage(): AdvancedSettings {
  if (typeof window === 'undefined') return DEFAULT_SETTINGS;

  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) return DEFAULT_SETTINGS;

    const parsed = JSON.parse(stored);
    return { ...DEFAULT_SETTINGS, ...parsed };
  } catch {
    return DEFAULT_SETTINGS;
  }
}

export function getAdvancedSettings(): AdvancedSettings {
  return loadSettingsFromStorage();
}

export function saveSettingsToStorage(settings: AdvancedSettings): void {
  if (typeof window === 'undefined') return;

  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
  } catch (error) {
    console.error('Failed to save advanced settings to storage:', error);
  }
}
