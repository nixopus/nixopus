'use client';

import React, { useState, useEffect } from 'react';
import { Keyboard } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { DialogWrapper } from '@/components/ui/dialog-wrapper';
import { Separator } from '@/components/ui/separator';

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

export function KeyboardShortcuts() {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 's' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setIsOpen((prev) => !prev);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  const halfLength = Math.ceil(shortcuts.length / 2);
  const leftColumn = shortcuts.slice(0, halfLength);
  const rightColumn = shortcuts.slice(halfLength);

  const trigger = (
    <Button variant="outline" size="icon" className="h-9 w-9" data-slot="keyboard-shortcuts">
      <Keyboard className="h-4 w-4" />
    </Button>
  );

  return (
    <DialogWrapper
      open={isOpen}
      onOpenChange={setIsOpen}
      title="Keyboard Shortcuts"
      trigger={trigger}
      size="lg"
    >
      <div className="grid grid-cols-2 gap-4 py-4">
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
    </DialogWrapper>
  );
}
