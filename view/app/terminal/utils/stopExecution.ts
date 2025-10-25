import { useEffect, useState } from 'react';

export const StopExecution = () => {
  const [isStopped, setIsStopped] = useState(false);
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'c' && (e.metaKey || e.ctrlKey) && e.shiftKey) {
        e.preventDefault();
        console.log('Stopped execution');
        setIsStopped(true);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  return {
    isStopped,
    setIsStopped
  };
};
