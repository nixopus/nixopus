import { useEffect } from 'react';

type UseTerminalKeyboardShortcutsProps = {
  isTerminalOpen: boolean;
  activeSessionId: string | null;
  activePaneId: string | null;
  splitPanesCount: number;
  sessionsCount: number;
  onCloseSplitPane: (paneId: string) => void;
  onCloseSession: (sessionId: string) => void;
  onToggleTerminal: () => void;
};

export const useTerminalKeyboardShortcuts = ({
  isTerminalOpen,
  activeSessionId,
  activePaneId,
  splitPanesCount,
  sessionsCount,
  onCloseSplitPane,
  onCloseSession,
  onToggleTerminal
}: UseTerminalKeyboardShortcutsProps) => {
  useEffect(() => {
    if (!isTerminalOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // Check for CMD+D (Mac) or CTRL+D (Windows/Linux)
      const isModifierKey = e.ctrlKey || e.metaKey;
      const key = e.key?.toLowerCase();
      const isDPressed = key === 'd';

      if (isModifierKey && isDPressed && activeSessionId && sessionsCount > 0) {
        // dont trigger if user is typing in an input field or textarea
        const target = e.target as HTMLElement;
        if (target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement) {
          return;
        }

        e.preventDefault();
        e.stopPropagation();
        e.stopImmediatePropagation();

        // close split pane, session, or session + terminal panel for last session
        if (splitPanesCount > 1 && activePaneId) {
          onCloseSplitPane(activePaneId);
        } else if (activeSessionId) {
          onCloseSession(activeSessionId);
        } else {
          onToggleTerminal();
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown, true);
    return () => document.removeEventListener('keydown', handleKeyDown, true);
  }, [
    isTerminalOpen,
    activeSessionId,
    activePaneId,
    splitPanesCount,
    sessionsCount,
    onCloseSplitPane,
    onCloseSession,
    onToggleTerminal
  ]);
};
