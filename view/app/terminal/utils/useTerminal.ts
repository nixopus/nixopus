import { useState, useRef, useCallback, useEffect } from 'react';
import { StopExecution } from './stopExecution';
import { useWebSocket } from '@/hooks/socket_provider';

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
};

export const useTerminal = () => {
  const terminalRef = useRef<HTMLDivElement | null>(null);
  const fitAddonRef = useRef<any | null>(null);
  const { isStopped, setIsStopped } = StopExecution();
  const { sendJsonMessage, message, isReady } = useWebSocket();
  const [isTerminalReady, setIsTerminalReady] = useState(false);
  const [isInitializing, setIsInitializing] = useState(false);
  const [terminalInstance, setTerminalInstance] = useState<any | null>(null);

  useEffect(() => {
    if (isStopped && terminalInstance) {
      sendJsonMessage({ action: 'terminal', data: CTRL_C });
      setIsStopped(false);
    }
  }, [isStopped, sendJsonMessage, setIsStopped]);

  useEffect(() => {
    if (!message || !terminalInstance) return;

    try {
      const parsedMessage: TerminalOutput =
        typeof message === 'string' && message.startsWith('{') ? JSON.parse(message) : message;

      if (!parsedMessage.data) {
        // terminalInstance.write(message);
        return;
      }

      const { output_type, content } = parsedMessage.data;

      if (output_type === OutputType.EXIT) {
        terminalInstance.dispose();
        setTerminalInstance(null);
        setIsTerminalReady(false);
      } else {
        const formattedContent =
          output_type === OutputType.STDERR ? `\x1B[31m${content}\x1B[0m` : content;
        terminalInstance.write(formattedContent);
      }
    } catch (error) {
      console.error('Error processing WebSocket message:', error);
    }
  }, [message, terminalInstance]);

  const initializeTerminal = useCallback(async () => {
    if (!terminalRef.current || terminalInstance) return;
    try {
      const { Terminal } = await import('@xterm/xterm');
      const { FitAddon } = await import('xterm-addon-fit');
      const { WebLinksAddon } = await import('xterm-addon-web-links');

      const term = new Terminal({
        cursorBlink: true,
        fontFamily: '"Menlo", "DejaVu Sans Mono", "Consolas", monospace',
        fontSize: 14,
        theme: { foreground: 'hsl(142.1 71% 45%)', background: 'hsl(240 4% 16%)', cursor: 'red' },
        allowTransparency: true,
        rightClickSelectsWord: true,
        disableStdin: false,
        convertEol: false
      });

      const fitAddon = new FitAddon();
      const webLinksAddon = new WebLinksAddon();

      term.loadAddon(fitAddon);
      term.loadAddon(webLinksAddon);
      fitAddonRef.current = fitAddon;
      term.open(terminalRef.current);
      fitAddon.activate(term);
      fitAddon.fit();

      if (terminalRef.current) {
        terminalRef.current.style.padding = '5px';
      }

      term.onData((data) => {
        sendJsonMessage({ action: 'terminal', data });
      });

      setTerminalInstance(term);
    } catch (error) {
      console.error('Error initializing terminal:', error);
      setIsInitializing(false);
    }
  }, [sendJsonMessage, isInitializing, terminalRef]);

  const destroyTerminal = useCallback(() => {
    if (terminalInstance) {
      terminalInstance.dispose();
      setTerminalInstance(null);
    }
  }, [terminalInstance]);

  useEffect(() => {
    return destroyTerminal;
  }, [destroyTerminal]);

  return {
    terminalRef,
    initializeTerminal,
    destroyTerminal,
    fitAddonRef,
    terminalInstance,
    isTerminalReady
  };
};
