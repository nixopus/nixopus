'use client';
import React, { useEffect, useState, useRef, useCallback } from 'react';
import '@xterm/xterm/css/xterm.css';
import { useTerminal } from './utils/useTerminal';
import { useContainerReady } from './utils/isContainerReady';
import { X } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { useFeatureFlags } from '@/hooks/features_provider';
import DisabledFeature from '@/components/features/disabled-feature';
import Skeleton from '@/app/file-manager/components/skeleton/Skeleton';
import { FeatureNames } from '@/types/feature-flags';
import { AnyPermissionGuard, ResourceGuard } from '@/components/rbac/PermissionGuard';
import { useRBAC } from '@/lib/rbac';

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
  const { t } = useTranslation();
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });
  const containerRef = useRef<HTMLDivElement>(null);
  const resizeTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
  const { canAccessResource } = useRBAC();

  const canCreate = canAccessResource('terminal', 'create');
  const canUpdate = canAccessResource('terminal', 'update');

  const { terminalRef, fitAddonRef, initializeTerminal, destroyTerminal } = useTerminal(
    isTerminalOpen,
    dimensions.width,
    dimensions.height,
    canCreate || canUpdate // Only allow input if user can create or update
  ) as {
    terminalRef: React.RefObject<HTMLDivElement>;
    fitAddonRef: any;
    initializeTerminal: () => void;
    destroyTerminal: () => void;
  };

  const isContainerReady = useContainerReady(isTerminalOpen, terminalRef);
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();
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

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureTerminal)) {
    return <DisabledFeature />;
  }

  return (
    <AnyPermissionGuard 
      permissions={['terminal:create', 'terminal:read', 'terminal:update']}
      loadingFallback={<Skeleton />}
    >
      <div
        className="flex h-full flex-col overflow-hidden bg-[#1e1e1e]"
        ref={containerRef}
        data-slot="terminal"
      >
        <div className="flex h-8 items-center justify-between border-b border-[#2d2d2d] px-3">
          <div className="flex items-center gap-2">
            <span className="text-xs font-medium text-[#cccccc]">{t('terminal.title')}</span>
            <span className="text-xs text-[#666666]">{t('terminal.shortcut')}</span>
          </div>
          <div className="flex items-center gap-2">
            <button
              className="flex h-4 w-4 items-center justify-center rounded hover:bg-[#2d2d2d]"
              onClick={toggleTerminal}
              title={t('terminal.close')}
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
    </AnyPermissionGuard>
  );
};
