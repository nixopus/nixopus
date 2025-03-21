'use client';
import React, { useEffect } from 'react';
import '@xterm/xterm/css/xterm.css';
import { useTerminal } from './utils/useTerminal';
import { useContainerReady } from './utils/isContainerReady';
import { X } from 'lucide-react';

type terminalProps = {
    isOpen: boolean;
    toggleTerminal: () => void;
    isTerminalOpen: boolean;
    setFitAddonRef: React.Dispatch<React.SetStateAction<any | null>>;
};

export const IntegratedTerminal: React.FC<terminalProps> = ({
    isOpen,
    toggleTerminal,
    isTerminalOpen,
    setFitAddonRef,
}) => {
    const { terminalRef, fitAddonRef, initializeTerminal } = useTerminal() as {
        terminalRef: React.RefObject<HTMLDivElement>;
        fitAddonRef: any;
        initializeTerminal: () => void;
    };
    const isContainerReady = useContainerReady(isTerminalOpen, terminalRef);

    useEffect(() => {
        if (isTerminalOpen && isContainerReady) {
            setTimeout(initializeTerminal, 0);
        }
    }, [isTerminalOpen, isContainerReady, initializeTerminal]);

    useEffect(() => {
        if (fitAddonRef) {
            setFitAddonRef(fitAddonRef);
        }
    }, [fitAddonRef]);

    return (
        <div className="flex h-full flex-col border border-t-0">
            <div className="flex h-5 items-center justify-between bg-secondary px-1 py-2 border-b-2 opacity-50">
                <span className="text-xs">
                    Terminal <span className="text-xs">⌘</span>J
                </span>
                <X className="h-4 w-4 hover:text-destructive" onClick={toggleTerminal} />
            </div>
            <div
                ref={terminalRef}
                className="flex-grow overflow-hidden bg-secondary"
                style={{
                    height: isTerminalOpen ? 'calc(100% - 32px)' : '0',
                    visibility: isTerminalOpen ? 'visible' : 'hidden',
                }}
            />
        </div>
    );
};
