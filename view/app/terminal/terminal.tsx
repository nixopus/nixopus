'use client';
import React, { useEffect, useState, useRef, useCallback } from 'react';
import '@xterm/xterm/css/xterm.css';
import { useTerminal } from './utils/useTerminal';
import { useContainerReady } from './utils/isContainerReady';
import { X } from 'lucide-react';

const globalStyles = `
  .xterm-viewport::-webkit-scrollbar {
    display: none;
  }
  .xterm-viewport {
    scrollbar-width: none;
    -ms-overflow-style: none;
  }
`;

type TerminalProps = {
  isOpen: boolean;
  toggleTerminal: () => void;
  isTerminalOpen: boolean;
  setFitAddonRef: React.Dispatch<React.SetStateAction<any | null>>;
};

export const Terminal: React.FC<TerminalProps> = ({
  isOpen,
  toggleTerminal,
  isTerminalOpen,
  setFitAddonRef
}) => {
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });
  const containerRef = useRef<HTMLDivElement>(null);
  const resizeTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);

  const { terminalRef, fitAddonRef, initializeTerminal, destroyTerminal } = useTerminal(
    isTerminalOpen,
    dimensions.width,
    dimensions.height
  ) as {
    terminalRef: React.RefObject<HTMLDivElement>;
    fitAddonRef: any;
    initializeTerminal: () => void;
    destroyTerminal: () => void;
  };

  const isContainerReady = useContainerReady(isTerminalOpen, terminalRef);

  const updateDimensions = useCallback(() => {
    if (!containerRef.current) return;

    if (resizeTimeoutRef.current) {
      clearTimeout(resizeTimeoutRef.current);
    }

    resizeTimeoutRef.current = setTimeout(() => {
      if (containerRef.current) {
        setDimensions({
          width: containerRef.current.offsetWidth,
          height: containerRef.current.offsetHeight
        });
      }
    }, 100);
  }, []);

  useEffect(() => {
    if (!containerRef.current) return;

    updateDimensions();

    const resizeObserver = new ResizeObserver(updateDimensions);
    resizeObserver.observe(containerRef.current);

    return () => {
      resizeObserver.disconnect();
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
      }
    };
  }, [isTerminalOpen, updateDimensions]);

  useEffect(() => {
    if (isTerminalOpen && isContainerReady) {
      initializeTerminal();
    } else {
      destroyTerminal();
    }
  }, [isTerminalOpen, isContainerReady, initializeTerminal, destroyTerminal]);

  useEffect(() => {
    if (fitAddonRef) {
      setFitAddonRef(fitAddonRef);
    }
  }, [fitAddonRef, setFitAddonRef]);

  useEffect(() => {
    const style = document.createElement('style');
    style.textContent = globalStyles;
    document.head.appendChild(style);
    return () => {
      document.head.removeChild(style);
    };
  }, []);

  return (
    <div className="flex h-full flex-col overflow-hidden bg-[#1e1e1e]" ref={containerRef}>
      <div className="flex h-8 items-center justify-between border-b border-[#2d2d2d] px-3">
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium text-[#cccccc]">Terminal</span>
          <span className="text-xs text-[#666666]">⌘J</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            className="flex h-4 w-4 items-center justify-center rounded hover:bg-[#2d2d2d]"
            onClick={toggleTerminal}
          >
            <X className="h-3 w-3 text-[#666666] hover:text-[#cccccc]" />
          </button>
        </div>
      </div>
      <div
        ref={terminalRef}
        className="flex-1 relative"
        style={{
          visibility: isTerminalOpen ? 'visible' : 'hidden',
          minHeight: '200px',
          padding: '4px',
          overflow: 'hidden',
          backgroundColor: '#1e1e1e',
          scrollbarWidth: 'none',
          msOverflowStyle: 'none'
        }}
      />
    </div>
  );
};
