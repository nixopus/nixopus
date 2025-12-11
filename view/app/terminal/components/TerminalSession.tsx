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

  const hasInitializedRef = useRef(false);

  useEffect(() => {
    if (isTerminalOpen && isActive && isContainerReady && !hasInitializedRef.current) {
      hasInitializedRef.current = true;
      onStatusChangeRef.current?.('loading');
      initializeTerminal();
      const timer = setTimeout(() => onStatusChangeRef.current?.('active'), 500);
      return () => clearTimeout(timer);
    }
  }, [isTerminalOpen, isActive, isContainerReady, initializeTerminal]);

  useEffect(() => {
    if (fitAddonRef) {
      setFitAddonRef(fitAddonRef);
    }
  }, [fitAddonRef, setFitAddonRef]);

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
