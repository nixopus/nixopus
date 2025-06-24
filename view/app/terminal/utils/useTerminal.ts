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
  allowInput: boolean = true
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
  }, [terminalInstance]);

  useEffect(() => {
    if (isStopped && terminalInstance) {
      sendJsonMessage({ action: 'terminal', data: CTRL_C });
      setIsStopped(false);
    }
  }, [isStopped, sendJsonMessage, setIsStopped, terminalInstance]);

  useEffect(() => {
    if (!isTerminalOpen) {
      destroyTerminal();
    }
  }, [isTerminalOpen, destroyTerminal]);

  useEffect(() => {
    if (!message || !terminalInstance) return;

    try {
      const parsedMessage =
        typeof message === 'string' && message.startsWith('{') ? JSON.parse(message) : message;

      if (parsedMessage.action === 'error') {
        console.error('Terminal error:', parsedMessage.data);
        return;
      }

      if (parsedMessage.data && parsedMessage.data.output_type) {
        const { output_type, content } = parsedMessage.data;
        if (output_type === OutputType.EXIT) {
          destroyTerminal();
        } else {
          terminalInstance.write(content);
        }
      }
    } catch (error) {
      console.error('Error processing WebSocket message:', error);
    }
  }, [message, terminalInstance, destroyTerminal]);

  const initializeTerminal = useCallback(async () => {
    if (!terminalRef.current || terminalInstance || !isReady) return;

    try {
      const { Terminal } = await import('@xterm/xterm');
      const { FitAddon } = await import('xterm-addon-fit');
      const { WebLinksAddon } = await import('xterm-addon-web-links');

      const term = new Terminal({
        cursorBlink: true,
        fontFamily: '"Menlo", "DejaVu Sans Mono", "Consolas", monospace',
        fontSize: 14,
        theme: {
          foreground: '#cccccc',
          background: '#1e1e1e',
          cursor: '#cccccc',
          black: '#000000',
          red: '#cd3131',
          green: '#0dbc79',
          yellow: '#e5e510',
          blue: '#2472c8',
          magenta: '#bc3fbc',
          cyan: '#11a8cd',
          white: '#e5e5e5',
          brightBlack: '#666666',
          brightRed: '#f14c4c',
          brightGreen: '#23d18b',
          brightYellow: '#f5f543',
          brightBlue: '#3b8eea',
          brightMagenta: '#d670d6',
          brightCyan: '#29b8db',
          brightWhite: '#e5e5e5'
        },
        allowTransparency: true,
        rightClickSelectsWord: true,
        disableStdin: !allowInput,
        convertEol: true,
        scrollback: 1000,
        tabStopWidth: 8,
        macOptionIsMeta: true,
        macOptionClickForcesSelection: true
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
            data: '\r'
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
                rows: dimensions.rows
              }
            });
          }
        });

        if (allowInput) {
          term.onData((data) => {
            sendJsonMessage({
              action: 'terminal',
              data
            });
          });
        }

        term.onResize((size) => {
          sendJsonMessage({
            action: 'terminal_resize',
            data: {
              cols: size.cols,
              rows: size.rows
            }
          });
        });
      }

      setTerminalInstance(term);
    } catch (error) {
      console.error('Error initializing terminal:', error);
    }
  }, [sendJsonMessage, isReady, terminalRef, terminalInstance, allowInput]);

  useEffect(() => {
    return () => {
      destroyTerminal();
    };
  }, [destroyTerminal]);

  return {
    terminalRef,
    initializeTerminal,
    destroyTerminal,
    fitAddonRef,
    terminalInstance
  };
};
