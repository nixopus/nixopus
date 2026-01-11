'use client';

import React, { useEffect, useRef } from 'react';
import { useTerminal } from '../utils/useTerminal';
import { useContainerReady } from '../utils/isContainerReady';
import { cn } from '@/lib/utils';
import type { ExitHandler, SessionStatus } from '../types';

type TerminalSessionProps = {
  isActive: boolean;
  isTerminalOpen: boolean;
  dimensions: { width: number; height: number };
  canCreate: boolean;
  canUpdate: boolean;
  setFitAddonRef: React.Dispatch<React.SetStateAction<any | null>>;
  terminalId: string;
  onStatusChange?: (status: SessionStatus) => void;
  exitHandler?: ExitHandler;
};

export const TerminalSession: React.FC<TerminalSessionProps> = ({
  isActive,
  isTerminalOpen,
  dimensions,
  canCreate,
  canUpdate,
  setFitAddonRef,
  terminalId,
  onStatusChange,
  exitHandler
}) => {
  const { terminalRef, fitAddonRef, initializeTerminal, destroyTerminal } = useTerminal(
    isTerminalOpen,
    dimensions.width,
    dimensions.height,
    canCreate || canUpdate,
    terminalId,
    exitHandler
  );

  const isContainerReady = useContainerReady(
    isTerminalOpen && isActive,
    terminalRef as React.RefObject<HTMLDivElement>
  );

  const onStatusChangeRef = useRef(onStatusChange);
  onStatusChangeRef.current = onStatusChange;

  useEffect(() => {
    // Initialize terminal when visible and active, but keep it alive when inactive to preserve state.
    if (isTerminalOpen && isActive && isContainerReady) {
      onStatusChangeRef.current?.('loading');
      initializeTerminal();
      const timer = setTimeout(() => onStatusChangeRef.current?.('active'), 500);
      return () => clearTimeout(timer);
    }

    // When inactive, just mark as idle but keep terminal instance alive to preserve state.
    // Output will continue to be processed in the background.
    if (!isActive && isTerminalOpen) {
      onStatusChangeRef.current?.('idle');
    }
  }, [isTerminalOpen, isActive, isContainerReady, initializeTerminal]);

  useEffect(() => {
    if (fitAddonRef) {
      setFitAddonRef(fitAddonRef);
    }
  }, [fitAddonRef, setFitAddonRef]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      destroyTerminal();
    };
  }, [destroyTerminal]);

  return (
    <div
      ref={terminalRef}
      className={cn(
        'flex-1 relative transition-opacity duration-300',
        isTerminalOpen && isActive ? 'opacity-100' : 'opacity-0 pointer-events-none'
      )}
      style={{
        visibility: isTerminalOpen && isActive ? 'visible' : 'hidden',
        minHeight: '200px',
        padding: '8px',
        overflow: 'hidden',
        backgroundColor: 'var(--terminal-bg)',
        scrollbarWidth: 'thin',
        height: '100%',
        width: '100%',
        maxWidth: '100%',
        boxSizing: 'border-box',
        contain: 'inline-size',
        animation: isActive ? 'terminalFadeIn 0.3s ease-out' : 'none'
      }}
    />
  );
};
