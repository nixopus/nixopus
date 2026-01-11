'use client';
import React, { useRef } from 'react';
import '@xterm/xterm/css/xterm.css';
import { useContainerTerminal } from '@/packages/hooks/terminal/use-container-terminal';
import { TerminalProps } from '@/packages/types/containers';

export const Terminal: React.FC<TerminalProps> = ({ containerId }) => {
  const terminalRef = useRef<HTMLDivElement | null>(null);
  const { terminalRef: termRef } = useContainerTerminal(containerId);

  return (
    <div
      ref={(el) => {
        terminalRef.current = el;
        // @ts-ignore
        if (termRef) termRef.current = el;
      }}
      className="relative"
      style={{ height: '60vh', minHeight: 300, backgroundColor: '#1e1e1e' }}
    />
  );
};

export default Terminal;
