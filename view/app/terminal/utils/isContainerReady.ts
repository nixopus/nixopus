import { useState, useLayoutEffect } from 'react';

export const useContainerReady = (
  isTerminalOpen: boolean,
  terminalRef?: React.RefObject<HTMLDivElement> | null
) => {
  const [isContainerReady, setIsContainerReady] = useState(false);

  useLayoutEffect(() => {
    let timeoutId: NodeJS.Timeout;
    let rafId: number;

    const checkSize = () => {
      timeoutId = setTimeout(() => {
        if (terminalRef?.current && isTerminalOpen) {
          const { offsetHeight, offsetWidth } = terminalRef.current;
          if (offsetHeight > 0 && offsetWidth > 0) {
            setIsContainerReady(true);
          } else {
            rafId = requestAnimationFrame(checkSize);
          }
        } else {
          setIsContainerReady(false);
        }
      }, 50);
    };

    if (isTerminalOpen) {
      checkSize();
    } else {
      setIsContainerReady(false);
    }

    return () => {
      clearTimeout(timeoutId);
      cancelAnimationFrame(rafId);
    };
  }, [isTerminalOpen, terminalRef]);

  return isContainerReady;
};
