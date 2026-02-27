import { useState, useCallback, useEffect } from 'react';

export const useTerminalState = () => {
  const [isTerminalOpen, setIsTerminalOpen] = useState<boolean>(() => {
    if (typeof window !== 'undefined') {
      const savedState = localStorage.getItem('terminalOpen');
      return savedState !== null ? JSON.parse(savedState) : true;
    }
    return true;
  });

  const toggleTerminal = useCallback((): void => {
    setIsTerminalOpen((prev) => {
      const newState = !prev;
      localStorage.setItem('terminalOpen', JSON.stringify(newState));
      return newState;
    });

    setTimeout(() => {
      window.dispatchEvent(new Event('resize'));
    }, 500);
  }, []);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'j' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        toggleTerminal();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [toggleTerminal]);

  return { isTerminalOpen, toggleTerminal };
};
