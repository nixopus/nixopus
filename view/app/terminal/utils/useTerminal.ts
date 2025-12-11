import { useState, useRef, useCallback, useEffect } from 'react';
import { StopExecution } from './stopExecution';
import { useWebSocket } from '@/hooks/socket-provider';

const CTRL_C = '\x03';

enum OutputType {
  STDOUT = 'stdout',
  STDERR = 'stderr',
  EXIT = 'exit'
}

type TerminalOutput = {
  data: {
    output_type: string;
    content: string;
  };
  topic: string;
};

export const useTerminal = (
  isTerminalOpen: boolean,
  width: number,
  height: number,
  allowInput: boolean = true,
  terminalId: string = 'terminal_id'
) => {
  const terminalRef = useRef<HTMLDivElement | null>(null);
  const fitAddonRef = useRef<any | null>(null);
  const { isStopped, setIsStopped } = StopExecution();
  const { sendJsonMessage, message, isReady } = useWebSocket();
  const [terminalInstance, setTerminalInstance] = useState<any | null>(null);
  const resizeTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);

  const destroyTerminal = useCallback(() => {
    if (terminalInstance) {
      terminalInstance.dispose();
      setTerminalInstance(null);
    }
    if (resizeTimeoutRef.current) {
      clearTimeout(resizeTimeoutRef.current);
    }
  }, [terminalInstance, terminalId]);

  useEffect(() => {
    if (isStopped && terminalInstance) {
      sendJsonMessage({ action: 'terminal', data: { value: CTRL_C, terminalId } });
      setIsStopped(false);
    }
  }, [isStopped, sendJsonMessage, setIsStopped, terminalInstance, terminalId]);

  useEffect(() => {
    if (!message || !terminalInstance) return;

    try {
      const parsedMessage =
        typeof message === 'string' && message.startsWith('{') ? JSON.parse(message) : message;

      if (parsedMessage.terminal_id !== terminalId) {
        console.log('Message is not for this terminal session');
        return;
      }

      if (parsedMessage.action === 'error') {
        console.error('Terminal error:', parsedMessage.data);
        return;
      }

      if (parsedMessage.type) {
        if (parsedMessage.type === OutputType.EXIT) {
          destroyTerminal();
        } else if (parsedMessage.data) {
          terminalInstance.write(parsedMessage.data);
        }
      }
    } catch (error) {
      console.error('Error processing WebSocket message:', error);
    }
  }, [message, terminalInstance, destroyTerminal, terminalId]);

  const initializeTerminal = useCallback(async () => {
    if (!terminalRef.current || terminalInstance || !isReady) return;

    try {
      const { Terminal } = await import('@xterm/xterm');
      const { FitAddon } = await import('xterm-addon-fit');
      const { WebLinksAddon } = await import('xterm-addon-web-links');

      const term = new Terminal({
        cursorBlink: true,
        cursorStyle: 'bar',
        cursorWidth: 2,
        fontFamily:
          '"JetBrains Mono", "Fira Code", "Cascadia Code", "SF Mono", Menlo, Monaco, "Courier New", monospace',
        fontSize: 13,
        fontWeight: '400',
        fontWeightBold: '600',
        letterSpacing: 0,
        lineHeight: 1.4,
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
        scrollback: 5000,
        tabStopWidth: 4,
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

        if (allowInput) {
          sendJsonMessage({
            action: 'terminal',
            data: { value: '\r', terminalId }
          });
        }

        requestAnimationFrame(() => {
          fitAddon.fit();
          const dimensions = fitAddon.proposeDimensions();
          if (dimensions) {
            sendJsonMessage({
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
            if (key === 'j' && (event.ctrlKey || event.metaKey)) {
              return false;
            } else if (key === 'c' && (event.ctrlKey || event.metaKey) && !event.shiftKey) {
              if (event.type === 'keydown') {
                try {
                  const selection = term.getSelection();
                  if (selection) {
                    navigator.clipboard.writeText(selection).then(() => {
                      term.clearSelection(); // Clear selection after successful copy
                    });
                    return false;
                  }
                } catch (error) {
                  console.error('Error in Ctrl+C handler:', error);
                }
              }
              return false;
            }
            return true;
          });
          term.onData((data) => {
            sendJsonMessage({
              action: 'terminal',
              data: { value: data, terminalId }
            });
          });
        }

        term.onResize((size) => {
          sendJsonMessage({
            action: 'terminal_resize',
            data: {
              cols: size.cols,
              rows: size.rows,
              terminalId
            }
          });
        });
      }

      setTerminalInstance(term);
    } catch (error) {
      console.error('Error initializing terminal:', error);
    }
  }, [sendJsonMessage, isReady, terminalRef, terminalInstance, allowInput, terminalId]);

  return {
    terminalRef,
    initializeTerminal,
    destroyTerminal,
    fitAddonRef,
    terminalInstance
  };
};
