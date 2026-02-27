import { useState, useEffect, useCallback, useRef } from 'react';

export const useTerminalDimensions = (
  containerRef: React.RefObject<HTMLDivElement | null>,
  isTerminalOpen: boolean
) => {
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });
  const resizeTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);

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
  }, [containerRef]);

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
  }, [isTerminalOpen, updateDimensions, containerRef]);

  return dimensions;
};
