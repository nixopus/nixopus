'use client';

import React from 'react';
import { useTranslation } from '@/hooks/use-translation';

interface Shortcut {
  keys: string[];
  description: string;
}

const shortcuts: Shortcut[] = [
  { keys: ['Ctrl', 'J'], description: 'Toggle terminal' },
  { keys: ['Ctrl', 'T'], description: 'Change terminal position' },
  { keys: ['Ctrl', 'B'], description: 'Toggle sidebar' },
  { keys: ['Ctrl', 'C'], description: 'Copy file' },
  { keys: ['Ctrl', 'X'], description: 'Cut file' },
  { keys: ['Ctrl', 'V'], description: 'Paste file' },
  { keys: ['Ctrl', 'H'], description: 'Toggle hidden files' },
  { keys: ['Ctrl', 'L'], description: 'Toggle layout (grid/list)' },
  { keys: ['Ctrl', 'Shift', 'N'], description: 'Create new folder' },
  { keys: ['F2'], description: 'Rename file' }
];

export function KeyboardShortcutsSettingsContent() {
  const { t } = useTranslation();
  const halfLength = Math.ceil(shortcuts.length / 2);
  const leftColumn = shortcuts.slice(0, halfLength);
  const rightColumn = shortcuts.slice(halfLength);

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-semibold">Keyboard Shortcuts</h2>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-4">
          {leftColumn.map((shortcut, index) => (
            <div key={index} className="flex items-center justify-between">
              <div className="text-sm text-muted-foreground">{shortcut.description}</div>
              <div className="flex items-center gap-1">
                {shortcut.keys.map((key, keyIndex) => (
                  <React.Fragment key={keyIndex}>
                    <kbd className="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">
                      {key}
                    </kbd>
                    {keyIndex < shortcut.keys.length - 1 && (
                      <span className="text-muted-foreground">+</span>
                    )}
                  </React.Fragment>
                ))}
              </div>
            </div>
          ))}
        </div>
        <div className="space-y-4">
          {rightColumn.map((shortcut, index) => (
            <div key={index} className="flex items-center justify-between">
              <div className="text-sm text-muted-foreground">{shortcut.description}</div>
              <div className="flex items-center gap-1">
                {shortcut.keys.map((key, keyIndex) => (
                  <React.Fragment key={keyIndex}>
                    <kbd className="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">
                      {key}
                    </kbd>
                    {keyIndex < shortcut.keys.length - 1 && (
                      <span className="text-muted-foreground">+</span>
                    )}
                  </React.Fragment>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

