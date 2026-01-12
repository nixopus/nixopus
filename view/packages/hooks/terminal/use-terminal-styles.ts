import { useEffect } from 'react';
import { terminalStyles } from '../../styles/terminal.styles';

export const useTerminalStyles = () => {
  useEffect(() => {
    const style = document.createElement('style');
    style.textContent = terminalStyles;
    document.head.appendChild(style);
    return () => {
      document.head.removeChild(style);
    };
  }, []);
};
