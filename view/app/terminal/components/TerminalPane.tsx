'use client';

import React, { useEffect, useRef, useState, useCallback } from 'react';
import { useTerminal } from '../utils/useTerminal';
import { useContainerReady } from '../utils/isContainerReady';
import { cn } from '@/lib/utils';
import type { SessionStatus } from './TerminalSession';

type TerminalPaneProps = {
  isActive: boolean;
  isTerminalOpen: boolean;
  canCreate: boolean;
  canUpdate: boolean;
  setFitAddonRef: React.Dispatch<React.SetStateAction<any | null>>;
  terminalId: string;
  onFocus: () => void;
  onStatusChange?: (status: SessionStatus) => void;
};

export const TerminalPane: React.FC<TerminalPaneProps> = ({
  isActive,
  isTerminalOpen,
  canCreate,
  canUpdate,
  setFitAddonRef,
  terminalId,
  onFocus,
  onStatusChange
}) => {
  const paneRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });
  const resizeTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);

  const updateDimensions = useCallback(() => {
    if (!paneRef.current) return;

    if (resizeTimeoutRef.current) {
      clearTimeout(resizeTimeoutRef.current);
    }

    resizeTimeoutRef.current = setTimeout(() => {
      if (paneRef.current) {
        setDimensions({
          width: paneRef.current.offsetWidth,
          height: paneRef.current.offsetHeight
        });
      }
    }, 100);
  }, []);

  useEffect(() => {
    if (!paneRef.current) return;

    // Force immediate dimension check
    const immediateCheck = () => {
      if (paneRef.current) {
        const rect = paneRef.current.getBoundingClientRect();
        if (rect.width > 0 && rect.height > 0) {
          setDimensions({
            width: rect.width,
            height: rect.height
          });
        }
      }
    };

    // Check immediately
    immediateCheck();

    // Also use the callback-based update
    updateDimensions();

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        if (entry.target === paneRef.current) {
          updateDimensions();
        }
      }
    });
    resizeObserver.observe(paneRef.current);

    // Multiple delayed checks to catch cases where ResizeObserver doesn't fire immediately
    const delayedChecks = [100, 200, 500].map((delay) =>
      setTimeout(() => {
        immediateCheck();
        updateDimensions();
      }, delay)
    );

    return () => {
      resizeObserver.disconnect();
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
      }
      delayedChecks.forEach((timeout) => clearTimeout(timeout));
    };
  }, [isTerminalOpen, updateDimensions]);

  const { terminalRef, fitAddonRef, initializeTerminal, destroyTerminal, isWebSocketReady } =
    useTerminal(
      isTerminalOpen,
      dimensions.width,
      dimensions.height,
      canCreate || canUpdate,
      terminalId
    );

  const isContainerReady = useContainerReady(
    isTerminalOpen,
    terminalRef as React.RefObject<HTMLDivElement>
  );

  const onStatusChangeRef = useRef(onStatusChange);
  onStatusChangeRef.current = onStatusChange;
  const hasInitializedRef = useRef(false);

  useEffect(() => {
    if (!isTerminalOpen || !isWebSocketReady) return;

    let initialized = false;

    const attemptInitialization = () => {
      // Prevent multiple initializations
      if (initialized || hasInitializedRef.current) return true;

      // Check if terminalRef is attached
      if (!terminalRef?.current) return false;

      // Try to get dimensions from paneRef if state dimensions are 0
      let finalWidth = dimensions.width;
      let finalHeight = dimensions.height;

      if (finalWidth === 0 || finalHeight === 0) {
        if (paneRef.current) {
          const rect = paneRef.current.getBoundingClientRect();
          if (rect.width > 0 && rect.height > 0) {
            finalWidth = rect.width;
            finalHeight = rect.height;
            // Update dimensions state immediately
            setDimensions({
              width: finalWidth,
              height: finalHeight
            });
          }
        }
      }

      // Also check terminalRef dimensions as fallback
      if (finalWidth === 0 || finalHeight === 0) {
        if (terminalRef.current) {
          const rect = terminalRef.current.getBoundingClientRect();
          if (rect.width > 0 && rect.height > 0) {
            finalWidth = rect.width;
            finalHeight = rect.height;
          }
        }
      }

      // Initialize if we have valid dimensions and container is ready
      if (finalWidth > 0 && finalHeight > 0 && (isContainerReady || terminalRef.current)) {
        onStatusChangeRef.current?.('loading');
        initializeTerminal();
        initialized = true;
        hasInitializedRef.current = true;
        setTimeout(() => onStatusChangeRef.current?.('active'), 500);
        return true;
      }
      return false;
    };

    // Try immediately
    if (attemptInitialization()) {
      return;
    }

    // Retry with delays to handle async updates
    const retryDelays = [50, 100, 200, 500];
    const timeouts: NodeJS.Timeout[] = [];

    retryDelays.forEach((delay) => {
      const timeout = setTimeout(() => {
        if (attemptInitialization()) {
          // Clear remaining timeouts if initialization succeeds
          timeouts.forEach((t) => clearTimeout(t));
        }
      }, delay);
      timeouts.push(timeout);
    });

    return () => {
      timeouts.forEach((timeout) => clearTimeout(timeout));
    };
  }, [
    isTerminalOpen,
    isContainerReady,
    initializeTerminal,
    dimensions.width,
    dimensions.height,
    isWebSocketReady,
    terminalRef
  ]);

  // Cleanup: destroy terminal when component unmounts
  useEffect(() => {
    return () => {
      destroyTerminal();
      hasInitializedRef.current = false;
    };
  }, [destroyTerminal]);

  // Re-fit terminal when dimensions change - but only if WebSocket is ready
  useEffect(() => {
    if (fitAddonRef?.current && dimensions.width > 0 && dimensions.height > 0 && isWebSocketReady) {
      requestAnimationFrame(() => {
        try {
          fitAddonRef.current?.fit();
        } catch (error) {
          // Ignore fit errors
        }
      });
    }
  }, [fitAddonRef, dimensions.width, dimensions.height, isWebSocketReady]);

  useEffect(() => {
    if (fitAddonRef && isActive) {
      setFitAddonRef(fitAddonRef);
    }
  }, [fitAddonRef, setFitAddonRef, isActive]);

  return (
    <div
      ref={paneRef}
      className="flex-1 relative"
      style={{
        minHeight: '200px',
        minWidth: '200px',
        padding: '4px',
        overflow: 'hidden',
        backgroundColor: '#0c0c0f',
        scrollbarWidth: 'none',
        msOverflowStyle: 'none',
        height: '100%',
        width: '100%',
        position: 'relative'
      }}
      onClick={onFocus}
      onFocus={onFocus}
      tabIndex={0}
    >
      <div
        ref={terminalRef}
        className={cn(
          'absolute inset-0 transition-opacity duration-200',
          isActive ? 'opacity-100' : 'opacity-70'
        )}
        style={{
          padding: '4px',
          height: '100%',
          width: '100%'
        }}
      />
    </div>
  );
};
