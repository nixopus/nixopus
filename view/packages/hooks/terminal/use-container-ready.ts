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
      if (!isTerminalOpen) {
        setIsContainerReady(false);
        return;
      }

      if (!terminalRef?.current) {
        // Retry if ref is not attached yet
        rafRef.current = requestAnimationFrame(checkSize);
        return;
      }

      // Check both the ref element and its parent for dimensions
      const element = terminalRef.current;
      const parent = element.parentElement;

      // Try multiple methods to get dimensions
      const elementHeight = element.offsetHeight || element.clientHeight;
      const elementWidth = element.offsetWidth || element.clientWidth;
      const parentHeight = parent?.offsetHeight || parent?.clientHeight || 0;
      const parentWidth = parent?.offsetWidth || parent?.clientWidth || 0;

      // Use parent dimensions if element dimensions are 0
      let height = elementHeight > 0 ? elementHeight : parentHeight;
      let width = elementWidth > 0 ? elementWidth : parentWidth;

      // Also check getBoundingClientRect as a fallback
      if (height === 0 || width === 0) {
        const rect = element.getBoundingClientRect();
        height = height > 0 ? height : rect.height;
        width = width > 0 ? width : rect.width;
      }

      // Check parent rect as well
      if (height === 0 || width === 0) {
        const parentRect = parent?.getBoundingClientRect();
        if (parentRect) {
          height = height > 0 ? height : parentRect.height;
          width = width > 0 ? width : parentRect.width;
        }
      }

      if (height > 0 && width > 0) {
        setIsContainerReady(true);
      } else {
        // Retry with requestAnimationFrame
        rafRef.current = requestAnimationFrame(checkSize);
      }
    };

    if (isTerminalOpen) {
      // Check immediately
      checkSize();
      // Also check after delays to catch async updates
      checkTimeoutRef.current = setTimeout(() => {
        checkSize();
      }, 100);
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
