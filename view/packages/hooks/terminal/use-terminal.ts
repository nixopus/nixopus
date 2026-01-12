import { useState, useRef, useCallback, useEffect } from 'react';
import { useWebSocket } from '@/packages/hooks/shared/socket-provider';
import { getAdvancedSettings } from '@/packages/utils/advanced-settings';
import type { ExitHandler, TerminalOutput } from '../../types/terminal';

const CTRL_C = '\x03';

enum OutputType {
  STDOUT = 'stdout',
  STDERR = 'stderr',
  EXIT = 'exit'
}

export const useTerminal = (
  isTerminalOpen: boolean,
  width: number,
  height: number,
  allowInput: boolean = true,
  terminalId: string = 'terminal_id',
  exitHandler?: ExitHandler
) => {
  const terminalRef = useRef<HTMLDivElement | null>(null);
  const fitAddonRef = useRef<any | null>(null);
  const { isStopped, setIsStopped } = StopExecution();
  const { sendJsonMessage, subscribe, isReady } = useWebSocket();
  const [terminalInstance, setTerminalInstance] = useState<any | null>(null);
  const resizeTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
  const terminalInstanceRef = useRef<any | null>(null);
  const pendingOutputRef = useRef<string[]>([]);
  const currentLineRef = useRef<string>('');

  // Keep refs for WebSocket to ensure we always use the latest values
  const isReadyRef = useRef(isReady);
  const sendJsonMessageRef = useRef(sendJsonMessage);

  useEffect(() => {
    isReadyRef.current = isReady;
  }, [isReady]);

  useEffect(() => {
    sendJsonMessageRef.current = sendJsonMessage;
  }, [sendJsonMessage]);

  const safeSendMessage = useCallback((data: any) => {
    if (isReadyRef.current) {
      sendJsonMessageRef.current(data);
    }
  }, []);

  const destroyTerminal = useCallback(() => {
    const instance = terminalInstanceRef.current;
    if (instance) {
      instance.dispose();
      terminalInstanceRef.current = null;
      setTerminalInstance(null);
    }
    pendingOutputRef.current = [];
    // Clear the terminal container to remove any stale input
    if (terminalRef.current) {
      terminalRef.current.innerHTML = '';
    }
    if (resizeTimeoutRef.current) {
      clearTimeout(resizeTimeoutRef.current);
    }
  }, []);

  useEffect(() => {
    if (isStopped && terminalInstance) {
      safeSendMessage({
        action: 'terminal',
        data: {
          value: CTRL_C,
          terminalId
        }
      });
      setIsStopped(false);
    }
  }, [isStopped, safeSendMessage, setIsStopped, terminalInstance, terminalId]);

  const handleTerminalFrame = useCallback(
    (raw: string) => {
      if (!raw) return;

      let parsedMessage: any;
      try {
        parsedMessage = typeof raw === 'string' && raw.startsWith('{') ? JSON.parse(raw) : raw;
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
        return;
      }

      if (parsedMessage?.terminal_id !== terminalId) {
        return;
      }

      if (parsedMessage?.action === 'error') {
        console.error('Terminal error:', parsedMessage.data);
        return;
      }

      if (parsedMessage?.type === OutputType.EXIT) {
        destroyTerminal();
        return;
      }

      if (parsedMessage?.data) {
        const instance = terminalInstanceRef.current;
        if (!instance) {
          // Terminal not ready yet; buffer output so we don't lose early frames.
          pendingOutputRef.current.push(parsedMessage.data);
          return;
        }

        // Write output even when terminal is inactive (hidden) to preserve state.
        // When user switches back, they'll see all output that happened while inactive.
        instance.write(parsedMessage.data);
      }
    },
    [destroyTerminal, terminalId]
  );

  useEffect(() => {
    // Critical: process every WS frame. Using a single `message` state drops frames under load.
    return subscribe(handleTerminalFrame);
  }, [subscribe, handleTerminalFrame]);

  // Cleanup terminal only when component unmounts (session/pane actually closed)
  // Keep terminal alive when panel is hidden or tabs are switched to preserve state
  useEffect(() => {
    return () => {
      // Only cleanup on unmount (when session/pane is actually closed)
      if (terminalInstanceRef.current) {
        destroyTerminal();
      }
    };
  }, [destroyTerminal]);

  const initializeTerminal = useCallback(async () => {
    if (!terminalRef.current || !isReadyRef.current) return;

    if (terminalInstance) return;

    try {
      const { Terminal } = await import('@xterm/xterm');
      const { FitAddon } = await import('@xterm/addon-fit');
      const { WebLinksAddon } = await import('@xterm/addon-web-links');

      // Get terminal settings from advanced settings
      const terminalSettings = getAdvancedSettings();

      // Build font family string with fallbacks
      const fontFamilyMap: Record<string, string> = {
        'JetBrains Mono':
          '"JetBrains Mono", "Fira Code", "Cascadia Code", "SF Mono", Menlo, Monaco, "Courier New", monospace',
        'Fira Code':
          '"Fira Code", "JetBrains Mono", "Cascadia Code", "SF Mono", Menlo, Monaco, "Courier New", monospace',
        'Cascadia Code':
          '"Cascadia Code", "JetBrains Mono", "Fira Code", "SF Mono", Menlo, Monaco, "Courier New", monospace',
        'SF Mono':
          '"SF Mono", "JetBrains Mono", "Fira Code", "Cascadia Code", Menlo, Monaco, "Courier New", monospace',
        Menlo: 'Menlo, "SF Mono", "JetBrains Mono", "Fira Code", Monaco, "Courier New", monospace',
        Monaco: 'Monaco, "SF Mono", Menlo, "JetBrains Mono", "Courier New", monospace',
        'Courier New': '"Courier New", Monaco, Menlo, "SF Mono", monospace'
      };

      const fontFamily =
        fontFamilyMap[terminalSettings.terminalFontFamily] ||
        `"${
          terminalSettings.terminalFontFamily
        }", "JetBrains Mono", "Fira Code", "Cascadia Code", "SF Mono", Menlo, Monaco, "Courier New", monospace`;

      const fontWeight = terminalSettings.terminalFontWeight === 'bold' ? '600' : '400';

      const term = new Terminal({
        cursorBlink: terminalSettings.terminalCursorBlink,
        cursorStyle: terminalSettings.terminalCursorStyle,
        cursorWidth: terminalSettings.terminalCursorWidth,
        fontFamily,
        fontSize: terminalSettings.terminalFontSize,
        fontWeight,
        fontWeightBold: '600',
        letterSpacing: terminalSettings.terminalLetterSpacing,
        lineHeight: terminalSettings.terminalLineHeight,
        theme: {
          // Warp-inspired dark theme with vibrant accents
          foreground: '#e4e4e7',
          background: '#0c0c0f',
          cursor: '#22d3ee',
          cursorAccent: '#0c0c0f',
          selectionBackground: '#3b82f620',
          selectionForeground: '#ffffff',
          selectionInactiveBackground: '#3b82f610',
          // ANSI colors - vibrant and modern
          black: '#18181b',
          red: '#f87171',
          green: '#4ade80',
          yellow: '#facc15',
          blue: '#60a5fa',
          magenta: '#c084fc',
          cyan: '#22d3ee',
          white: '#e4e4e7',
          // Bright variants
          brightBlack: '#52525b',
          brightRed: '#fca5a5',
          brightGreen: '#86efac',
          brightYellow: '#fde047',
          brightBlue: '#93c5fd',
          brightMagenta: '#d8b4fe',
          brightCyan: '#67e8f9',
          brightWhite: '#fafafa'
        },
        allowTransparency: true,
        rightClickSelectsWord: true,
        disableStdin: !allowInput,
        convertEol: true,
        scrollback: terminalSettings.terminalScrollback,
        tabStopWidth: terminalSettings.terminalTabStopWidth,
        macOptionIsMeta: true,
        macOptionClickForcesSelection: true,
        smoothScrollDuration: 100
      });

      const fitAddon = new FitAddon();
      const webLinksAddon = new WebLinksAddon();

      term.loadAddon(fitAddon);
      term.loadAddon(webLinksAddon);
      fitAddonRef.current = fitAddon;

      if (terminalRef.current) {
        terminalRef.current.innerHTML = '';
        term.open(terminalRef.current);
        fitAddon.activate(term);

        requestAnimationFrame(() => {
          fitAddon.fit();
          const dimensions = fitAddon.proposeDimensions();
          if (dimensions) {
            safeSendMessage({
              action: 'terminal_resize',
              data: {
                cols: dimensions.cols,
                rows: dimensions.rows,
                terminalId
              }
            });
          }
        });

        if (allowInput) {
          term.attachCustomKeyEventHandler((event: KeyboardEvent) => {
            const key = event.key.toLowerCase();

            // Handle Ctrl+J or Cmd+J (toggle terminal shortcut)
            if (key === 'j' && (event.ctrlKey || event.metaKey)) {
              return false;
            }

            // Handle Ctrl+D or Cmd+D (close terminal session shortcut)
            if (key === 'd' && (event.ctrlKey || event.metaKey)) {
              return false;
            }

            // Handle Ctrl+C or Cmd+C for copy (when there's a selection)
            if (key === 'c' && (event.ctrlKey || event.metaKey) && !event.shiftKey) {
              if (event.type === 'keydown') {
                try {
                  const selection = term.getSelection();
                  if (selection) {
                    navigator.clipboard.writeText(selection).then(() => {
                      term.clearSelection();
                    });
                    return false;
                  }
                } catch (error) {
                  console.error('Error in Ctrl+C handler:', error);
                }
              }
            }

            // Allow xterm to process all other keys normally
            return true;
          });

          const isEnterKey = (data: string): boolean => {
            return data === '\r' || data === '\n' || data === '\r\n';
          };

          const isBackspace = (data: string): boolean => {
            return data === '\x7f' || data === '\b';
          };

          const isEscapeSequence = (data: string): boolean => {
            return data.startsWith('\x1b');
          };

          const isPrintableAscii = (data: string): boolean => {
            return data.length === 1 && data >= ' ' && data.charCodeAt(0) < 127;
          };

          const handleExitCommand = (): boolean => {
            if (!exitHandler) return false;

            const {
              splitPanesCount,
              sessionsCount,
              activePaneId,
              activeSessionId,
              onCloseSplitPane,
              onCloseSession,
              onToggleTerminal
            } = exitHandler;

            // close split pane > close session > close session + terminal panel (for last session)
            const canCloseSplitPane = splitPanesCount > 1 && activePaneId && onCloseSplitPane;
            if (canCloseSplitPane) {
              onCloseSplitPane(activePaneId);
              return true;
            }

            if (activeSessionId && onCloseSession) {
              onCloseSession(activeSessionId);
              return true;
            }

            if (onToggleTerminal) {
              onToggleTerminal();
              return true;
            }

            return false;
          };

          const updateLineBuffer = (data: string): void => {
            if (isEnterKey(data)) {
              currentLineRef.current = '';
            } else if (isBackspace(data)) {
              if (currentLineRef.current.length > 0) {
                currentLineRef.current = currentLineRef.current.slice(0, -1);
              }
            } else if (isEscapeSequence(data)) {
              // Reset line buffer for escape sequences (arrow keys, function keys, etc.)
              currentLineRef.current = '';
            } else if (isPrintableAscii(data)) {
              // Add printable character, limit length to prevent memory issues
              if (currentLineRef.current.length < 1000) {
                currentLineRef.current += data;
              }
            }
          };

          // onData is called when xterm processes input
          term.onData((data) => {
            // "exit" command when Enter is pressed
            if (isEnterKey(data)) {
              const command = currentLineRef.current.trim().toLowerCase();
              if (command === 'exit') {
                // handle closing like CTRL+D/CMD+D
                currentLineRef.current = '';
                if (handleExitCommand()) {
                  return;
                }
              }
            }

            updateLineBuffer(data);

            // Send all input to backend
            safeSendMessage({
              action: 'terminal',
              data: {
                value: data,
                terminalId
              }
            });
          });
        }

        term.onResize((size) => {
          safeSendMessage({
            action: 'terminal_resize',
            data: {
              cols: size.cols,
              rows: size.rows,
              terminalId
            }
          });
        });
      }

      terminalInstanceRef.current = term;
      setTerminalInstance(term);

      // Flush any buffered output we received before xterm was ready.
      if (pendingOutputRef.current.length > 0) {
        const buffered = pendingOutputRef.current.join('');
        pendingOutputRef.current = [];
        term.write(buffered);
      }

      // instance is created and buffered output is flushed
      if (allowInput) {
        setTimeout(() => {
          safeSendMessage({
            action: 'terminal',
            data: {
              value: '\n',
              terminalId
            }
          });
        }, 100);
      }
    } catch (error) {
      console.error('Error initializing terminal:', error);
    }
  }, [safeSendMessage, terminalRef, terminalInstance, allowInput, terminalId]);

  return {
    terminalRef,
    initializeTerminal,
    destroyTerminal,
    fitAddonRef,
    terminalInstance,
    isWebSocketReady: isReady
  };
};

const StopExecution = () => {
  const [isStopped, setIsStopped] = useState(false);
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'c' && (e.metaKey || e.ctrlKey) && e.shiftKey) {
        e.preventDefault();
        console.log('Stopped execution');
        setIsStopped(true);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  return { isStopped, setIsStopped };
};
