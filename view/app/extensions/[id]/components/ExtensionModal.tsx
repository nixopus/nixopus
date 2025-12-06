'use client';

import ExtensionInput from '@/app/extensions/components/extension-input';
import type { Extension } from '@/redux/types/extension';

interface ExtensionModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  extension?: Extension;
  onSubmit: (values: Record<string, unknown>) => Promise<void>;
}

export function ExtensionModal({ open, onOpenChange, extension, onSubmit }: ExtensionModalProps) {
  return (
    <ExtensionInput
      open={open}
      onOpenChange={onOpenChange}
      extension={extension}
      onSubmit={onSubmit}
    />
  );
}
