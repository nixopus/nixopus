import { useState, useCallback } from 'react';

export function useResourceLimitsDialog(resetToCurrentValues: () => void) {
  const [open, setOpen] = useState(false);

  const handleOpenChange = useCallback(
    (newOpen: boolean) => {
      setOpen(newOpen);
      if (newOpen) {
        resetToCurrentValues();
      }
    },
    [resetToCurrentValues]
  );

  const handleCancel = useCallback(() => {
    setOpen(false);
  }, []);

  const closeDialog = useCallback(() => {
    setOpen(false);
  }, []);

  return {
    open,
    setOpen,
    handleOpenChange,
    handleCancel,
    closeDialog
  };
}
