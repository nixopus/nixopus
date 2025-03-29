import { useState, useLayoutEffect, useRef } from 'react';

export const useContainerReady = (
  isTerminalOpen: boolean,
  terminalRef?: React.RefObject<HTMLDivElement> | null
) => {
  const [isContainerReady, setIsContainerReady] = useState(false);
  const checkTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
  const rafRef = useRef<number | undefined>(undefined);

  useLayoutEffect(() => {
    const checkSize = () => {
      if (terminalRef?.current && isTerminalOpen) {
        const { offsetHeight, offsetWidth } = terminalRef.current;
        if (offsetHeight > 0 && offsetWidth > 0) {
          setIsContainerReady(true);
        } else {
          rafRef.current = requestAnimationFrame(checkSize);
        }
      } else {
        setIsContainerReady(false);
      }
    };

    if (isTerminalOpen) {
      checkTimeoutRef.current = setTimeout(() => {
        checkSize();
      }, 50);
    } else {
      setIsContainerReady(false);
    }

    return () => {
      if (checkTimeoutRef.current) {
        clearTimeout(checkTimeoutRef.current);
      }
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
      }
    };
  }, [isTerminalOpen, terminalRef]);

  return isContainerReady;
};
