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
  
  // Use refs to always have access to current values in callbacks
  const isReadyRef = useRef(isReady);
  const sendJsonMessageRef = useRef(sendJsonMessage);
  
  // Keep refs updated with current values
  useEffect(() => {
    isReadyRef.current = isReady;
  }, [isReady]);
  
  useEffect(() => {
    sendJsonMessageRef.current = sendJsonMessage;
  }, [sendJsonMessage]);
  
  // Safe send function that always checks current WebSocket state
  const safeSendMessage = useCallback((data: any) => {
    if (isReadyRef.current) {
      sendJsonMessageRef.current(data);
    }
  }, []);

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
      safeSendMessage({ action: 'terminal', data: { value: CTRL_C, terminalId } });
      setIsStopped(false);
    }
  }, [isStopped, safeSendMessage, setIsStopped, terminalInstance, terminalId]);

  useEffect(() => {
    if (!message) return;
    
    // Debug: Log when terminal receives any message
    console.log(`[Terminal ${terminalId.slice(0, 8)}] Received message, hasInstance: ${!!terminalInstance}`);
    
    if (!terminalInstance) {
      console.log(`[Terminal ${terminalId.slice(0, 8)}] No terminal instance, skipping message`);
      return;
    }

    try {
      const parsedMessage =
        typeof message === 'string' && message.startsWith('{') ? JSON.parse(message) : message;

      console.log(`[Terminal ${terminalId.slice(0, 8)}] Message terminal_id: ${parsedMessage.terminal_id?.slice(0, 8)}, My terminalId: ${terminalId.slice(0, 8)}`);

      if (parsedMessage.terminal_id !== terminalId) {
        console.log(`[Terminal ${terminalId.slice(0, 8)}] Message NOT for this terminal, ignoring`);
        return;
      }

      console.log(`[Terminal ${terminalId.slice(0, 8)}] Message IS for this terminal, processing...`);

      if (parsedMessage.action === 'error') {
        console.error('Terminal error:', parsedMessage.data);
        return;
      }

      if (parsedMessage.type) {
        if (parsedMessage.type === OutputType.EXIT) {
          destroyTerminal();
        } else if (parsedMessage.data) {
          console.log(`[Terminal ${terminalId.slice(0, 8)}] Writing data to terminal`);
          terminalInstance.write(parsedMessage.data);
        }
      }
    } catch (error) {
      console.error('Error processing WebSocket message:', error);
    }
  }, [message, terminalInstance, destroyTerminal, terminalId]);

  const initializeTerminal = useCallback(async () => {
    console.log(`[Terminal ${terminalId.slice(0, 8)}] initializeTerminal called, ref: ${!!terminalRef.current}, wsReady: ${isReadyRef.current}, hasInstance: ${!!terminalInstance}`);
    
    // Check current isReady value from ref
    if (!terminalRef.current || !isReadyRef.current) {
      console.log(`[Terminal ${terminalId.slice(0, 8)}] Cannot initialize: ref=${!!terminalRef.current}, wsReady=${isReadyRef.current}`);
      return;
    }
    
    // If terminal already exists, don't reinitialize unless it was disposed
    if (terminalInstance) {
      console.log(`[Terminal ${terminalId.slice(0, 8)}] Terminal already exists, skipping initialization`);
      return;
    }
    
    console.log(`[Terminal ${terminalId.slice(0, 8)}] Starting terminal initialization...`);

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
        console.log(`[Terminal ${terminalId.slice(0, 8)}] Opening terminal on DOM element:`, terminalRef.current);
        const containerRect = terminalRef.current.getBoundingClientRect();
        console.log(`[Terminal ${terminalId.slice(0, 8)}] Container dimensions: ${containerRect.width}x${containerRect.height}`);
        
        terminalRef.current.innerHTML = '';
        term.open(terminalRef.current);
        fitAddon.activate(term);

        requestAnimationFrame(() => {
          fitAddon.fit();
          const dimensions = fitAddon.proposeDimensions();
          console.log(`[Terminal ${terminalId.slice(0, 8)}] Terminal fitted, dimensions: ${dimensions?.cols}x${dimensions?.rows}`);
          if (dimensions) {
            console.log(`[Terminal ${terminalId.slice(0, 8)}] Sending resize message`);
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
          console.log(`[Terminal ${terminalId.slice(0, 8)}] Sending initial terminal message to start session`);
          safeSendMessage({
            action: 'terminal',
            data: { value: '\r', terminalId }
          });
        }

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
            // Use ref-based send to always have current WebSocket state
            console.log(`[Terminal ${terminalId.slice(0, 8)}] User input: ${JSON.stringify(data)}`);
            safeSendMessage({
              action: 'terminal',
              data: { value: data, terminalId }
            });
          });
        }

        term.onResize((size) => {
          // Use ref-based send to always have current WebSocket state
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

      console.log(`[Terminal ${terminalId.slice(0, 8)}] Terminal initialized successfully, setting instance`);
      setTerminalInstance(term);
    } catch (error) {
      console.error(`[Terminal ${terminalId.slice(0, 8)}] Error initializing terminal:`, error);
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
