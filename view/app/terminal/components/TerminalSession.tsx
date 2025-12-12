'use client';

import React, { useEffect, useRef } from 'react';
import { useTerminal } from '../utils/useTerminal';
import { useContainerReady } from '../utils/isContainerReady';
import { cn } from '@/lib/utils';

export type SessionStatus = 'active' | 'idle' | 'loading';

type TerminalSessionProps = {
  isActive: boolean;
  isTerminalOpen: boolean;
  dimensions: { width: number; height: number };
  canCreate: boolean;
  canUpdate: boolean;
  setFitAddonRef: React.Dispatch<React.SetStateAction<any | null>>;
  terminalId: string;
  onStatusChange?: (status: SessionStatus) => void;
};

export const TerminalSession: React.FC<TerminalSessionProps> = ({
  isActive,
  isTerminalOpen,
  dimensions,
  canCreate,
  canUpdate,
  setFitAddonRef,
  terminalId,
  onStatusChange
}) => {
  const { terminalRef, fitAddonRef, initializeTerminal, destroyTerminal } = useTerminal(
    isTerminalOpen && isActive,
    dimensions.width,
    dimensions.height,
    canCreate || canUpdate,
    terminalId
  );

  const isContainerReady = useContainerReady(
    isTerminalOpen && isActive,
    terminalRef as React.RefObject<HTMLDivElement>
  );

  const onStatusChangeRef = useRef(onStatusChange);
  onStatusChangeRef.current = onStatusChange;

  useEffect(() => {
    // Keep terminal lifecycle tightly coupled to visibility/activity to avoid “unsync”/ghost sessions.
    if (isTerminalOpen && isActive && isContainerReady) {
      onStatusChangeRef.current?.('loading');
      initializeTerminal();
      const timer = setTimeout(() => onStatusChangeRef.current?.('active'), 500);
      return () => clearTimeout(timer);
    }

    // If the session is not visible/active, tear it down to prevent hidden instances consuming output.
    destroyTerminal();
    onStatusChangeRef.current?.('idle');
  }, [isTerminalOpen, isActive, isContainerReady, initializeTerminal, destroyTerminal]);

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
